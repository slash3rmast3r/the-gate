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

func (s *Syncer) SyncApplications(apps []models.Application) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
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
		if err := s.writeApp(a); err != nil {
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

func (s *Syncer) writeApp(a models.Application) error {
	name := routerName(a)
	allowName := name + "-allowlist"
	basicName := name + "-basicauth"
	localMW := map[string]any{}

	mws := middlewaresFor(a, allowName, basicName)

	if cidrs := parseCIDRs(a.AllowCIDRs); len(cidrs) > 0 {
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

	rule := fmt.Sprintf("Host(`%s`)", a.Host)
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

func middlewaresFor(a models.Application, allowName, basicName string) []string {
	var mws []string
	if parseCIDRs(a.AllowCIDRs) != nil {
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

	switch a.WAFMode {
	case "off":
	case "crowdsec-only":
		mws = append(mws, "crowdsec@file")
	case "light":
		mws = append(mws, "crowdsec@file", "coraza-light@file")
	default:
		mws = append(mws, "crowdsec@file", "coraza-crs@file")
	}
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
	if a.CustomErrors {
		mws = append(mws, "gate-errors@file")
	}
	mws = append(mws, "secure-headers@file")
	return mws
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