package traefiksync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/overthecloud/the-gate/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type Syncer struct {
	dir                 string
	bouncerKey          string
	authentikForwardURL string
}

func New(dir, bouncerKey, authentikForwardURL string) *Syncer {
	return &Syncer{dir: dir, bouncerKey: bouncerKey, authentikForwardURL: authentikForwardURL}
}

func (s *Syncer) WriteBaseMiddlewares() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	tplPath := filepath.Join(s.dir, "middlewares.yml.template")
	outPath := filepath.Join(s.dir, "middlewares.yml")
	raw, err := os.ReadFile(tplPath)
	if err != nil {
		raw = []byte(minimalMiddlewares)
	}
	out := strings.ReplaceAll(string(raw), "${BOUNCER_KEY_TRAEFIK}", s.bouncerKey)
	out = strings.ReplaceAll(out, "${AUTHENTIK_FORWARD_URL}", s.authentikForwardURL)
	return os.WriteFile(outPath, []byte(out), 0o644)
}

func (s *Syncer) SyncApplications(apps []models.Application, groupSets ...[]models.IPGroup) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	groups := map[string]models.IPGroup{}
	for _, set := range groupSets {
		for _, g := range set {
			groups[g.ID.String()] = g
		}
	}
	entries, _ := os.ReadDir(s.dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "app-") && strings.HasSuffix(e.Name(), ".yml") {
			_ = os.Remove(filepath.Join(s.dir, e.Name()))
		}
	}

	for _, a := range apps {
		if !a.Enabled {
			continue
		}
		if err := s.writeApp(a, groups); err != nil {
			return err
		}
	}
	return nil
}

func routerName(a models.Application) string {
	id := strings.ReplaceAll(a.ID.String(), "-", "")
	if id == "" {
		id = sanitize(a.Host)
	}
	return "app-" + id
}

func (s *Syncer) writeApp(a models.Application, groups map[string]models.IPGroup) error {
	name := routerName(a)
	allowName := name + "-allowlist"
	basicName := name + "-basicauth"
	localMW := map[string]any{}

	cidrs := parseCIDRs(a.AllowCIDRs)
	if a.AllowGroupID != "" {
		if g, ok := groups[a.AllowGroupID]; ok && strings.EqualFold(g.Kind, "allow") {
			cidrs = append(cidrs, parseCIDRs(g.CIDRs)...)
		}
	}

	mws := middlewaresFor(a, allowName, basicName, len(cidrs) > 0)

	if len(cidrs) > 0 {
		localMW[allowName] = map[string]any{
			"ipAllowList": map[string]any{"sourceRange": cidrs},
		}
	}
	if strings.EqualFold(a.AuthMode, "basic") && a.BasicAuthUser != "" && a.BasicAuthPass != "" && a.BasicAuthPass != "********" {
		hash, err := bcrypt.GenerateFromPassword([]byte(a.BasicAuthPass), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		realm := a.BasicAuthRealm
		if realm == "" {
			realm = "The Gate"
		}
		localMW[basicName] = map[string]any{
			"basicAuth": map[string]any{
				"users":        []string{fmt.Sprintf("%s:%s", a.BasicAuthUser, string(hash))},
				"realm":        realm,
				"removeHeader": true,
			},
		}
	}

	hosts := []string{a.Host}
	for _, h := range parseCIDRs(a.ExtraHosts) {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "" || strings.EqualFold(h, a.Host) {
			continue
		}
		hosts = append(hosts, h)
	}
	hostRules := make([]string, 0, len(hosts))
	for _, h := range hosts {
		hostRules = append(hostRules, fmt.Sprintf("Host(`%s`)", h))
	}
	rule := strings.Join(hostRules, " || ")
	if len(hostRules) > 1 {
		rule = "(" + rule + ")"
	}
	prefix := strings.TrimSpace(a.PathPrefix)
	if prefix != "" && prefix != "/" {
		rule = fmt.Sprintf("%s && PathPrefix(`%s`)", rule, prefix)
	}

	router := map[string]any{
		"rule":        rule,
		"entryPoints": []string{"websecure"},
		"middlewares": mws,
		"service":     name,
	}
	if a.TLSEnabled {
		router["tls"] = map[string]any{"certResolver": "le"}
	} else {
		router["tls"] = map[string]any{}
	}

	httpBlock := map[string]any{
		"routers": map[string]any{name: router},
		"services": map[string]any{
			name: map[string]any{
				"loadBalancer": map[string]any{
					"servers": []map[string]string{
						{"url": a.UpstreamURL},
					},
					"passHostHeader": true,
				},
			},
		},
	}
	if len(localMW) > 0 {
		httpBlock["middlewares"] = localMW
	}

	doc := map[string]any{"http": httpBlock}
	b, err := yaml.Marshal(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, name+".yml"), b, 0o644)
}

func middlewaresFor(a models.Application, allowName, basicName string, hasAllowlist ...bool) []string {
	var mws []string
	allowlist := parseCIDRs(a.AllowCIDRs) != nil
	if len(hasAllowlist) > 0 {
		allowlist = hasAllowlist[0]
	}
	if allowlist {
		mws = append(mws, allowName+"@file")
	}
	switch strings.ToLower(a.AuthMode) {
	case "basic":
		if a.BasicAuthUser != "" && a.BasicAuthPass != "" && a.BasicAuthPass != "********" {
			mws = append(mws, basicName+"@file")
		}
	case "authentik", "forward-auth", "sso":
		mws = append(mws, "authentik-forward@file")
	}

	mode := strings.ToLower(a.DefenseMode)
	if mode == "" {
		mode = "normal"
	}
	botProtect := a.BotProtect || mode == "under_attack"
	attacks := a.Attacks || mode == "under_attack"
	httpFlood := a.HTTPFlood || mode == "under_attack"

	if botProtect || strings.EqualFold(a.WAFMode, "crowdsec-only") {
		mws = appendUnique(mws, "crowdsec@file")
	}
	if attacks {
		switch mode {
		case "under_attack":
			mws = appendUnique(mws, "crowdsec@file")
			mws = appendUnique(mws, "coraza-crs@file")
		case "audit":
			mws = appendUnique(mws, "coraza-audit@file")
		default:
			switch a.WAFMode {
			case "off", "crowdsec-only":
			case "light":
				mws = appendUnique(mws, "coraza-light@file")
			default:
				mws = appendUnique(mws, "coraza-crs@file")
			}
		}
	}
	if httpFlood {
		if mode == "under_attack" {
			mws = append(mws, "rate-limit-dos@file")
		} else {
			switch a.RateLimit {
			case "dos":
				mws = append(mws, "rate-limit-dos@file")
			case "strict":
				mws = append(mws, "rate-limit-strict@file")
			case "soft":
				mws = append(mws, "rate-limit-soft@file")
			case "off", "":
			default:
				mws = append(mws, "rate-limit-soft@file")
			}
		}
	}
	if strings.TrimSpace(a.BlockedCountries) != "" {
		mws = appendUnique(mws, "gate-geocheck@file")
	}
	// Cloudflare-like rules: always evaluate (handler fails open when none match)
	mws = appendUnique(mws, "gate-rules@file")
	if a.CustomErrors {
		mws = append(mws, "gate-errors@file")
	}
	mws = append(mws, "secure-headers@file")
	return mws
}

func appendUnique(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

func parseCIDRs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func sanitize(host string) string {
	h := strings.ToLower(host)
	h = strings.ReplaceAll(h, ".", "-")
	h = strings.ReplaceAll(h, "*", "star")
	h = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, h)
	return h
}

const minimalMiddlewares = `
http:
  middlewares:
    secure-headers:
      headers:
        browserXssFilter: true
        contentTypeNosniff: true
        frameDeny: true
        stsSeconds: 31536000
    crowdsec:
      plugin:
        crowdsec:
          enabled: true
          crowdsecMode: stream
          crowdsecLapiScheme: http
          crowdsecLapiHost: crowdsec:8080
          crowdsecLapiKey: "${BOUNCER_KEY_TRAEFIK}"
          crowdsecAppsecEnabled: true
          crowdsecAppsecHost: crowdsec:7422
          updateIntervalSeconds: 60
    authentik-forward:
      forwardAuth:
        address: "${AUTHENTIK_FORWARD_URL}"
        trustForwardHeader: true
        authResponseHeaders:
          - X-authentik-username
          - X-authentik-groups
          - X-authentik-email
          - X-authentik-name
          - X-authentik-uid
    gate-geocheck:
      forwardAuth:
        address: "http://gate-api:8080/internal/geo-check"
        trustForwardHeader: true
    gate-rules:
      forwardAuth:
        address: "http://gate-api:8080/internal/rules-check"
        trustForwardHeader: true
    gate-errors:
      errors:
        status:
          - "401"
          - "403"
          - "404"
          - "429"
          - "500"
          - "502"
          - "503"
        service: gate-errorpages
        query: "/{status}.html"
    coraza-crs:
      plugin:
        coraza:
          directives:
            - SecRuleEngine On
            - Include /etc/coraza/crs-setup.conf
            - Include /etc/coraza/rules/*.conf
    coraza-light:
      plugin:
        coraza:
          directives:
            - SecRuleEngine On
            - SecRequestBodyAccess Off
            - Include /etc/coraza/crs-setup.conf
            - Include /etc/coraza/rules/*.conf
    coraza-audit:
      plugin:
        coraza:
          directives:
            - SecRuleEngine DetectionOnly
            - SecRequestBodyAccess On
            - SecResponseBodyAccess Off
            - Include /etc/coraza/crs-setup.conf
            - Include /etc/coraza/rules/*.conf
    rate-limit-soft:
      rateLimit:
        average: 100
        burst: 50
        period: 1s
    rate-limit-strict:
      rateLimit:
        average: 20
        burst: 10
        period: 1s
    rate-limit-dos:
      rateLimit:
        average: 5
        burst: 10
        period: 1s
  services:
    gate-errorpages:
      loadBalancer:
        servers:
          - url: "http://gate-errors:80"
`
