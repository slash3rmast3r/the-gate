// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/models"
	"github.com/overthecloud/the-gate/internal/store"
)

// rulesCheck is Traefik forwardAuth for Cloudflare-like security rules.
func (s *Server) rulesCheck(c *gin.Context) {
	host := forwardedHost(c)
	uri := firstNonEmpty(c.GetHeader("X-Forwarded-Uri"), c.GetHeader("X-Forwarded-URI"), c.Request.URL.RequestURI())
	method := strings.ToUpper(firstNonEmpty(c.GetHeader("X-Forwarded-Method"), c.Request.Method))
	ip := clientIPFromRequest(c)
	ua := c.GetHeader("User-Agent")
	path, query := splitURI(uri)

	apps, err := store.ListApplications(s.db)
	if err != nil {
		c.Status(http.StatusOK)
		return
	}
	var app *models.Application
	for i := range apps {
		a := &apps[i]
		if !a.Enabled {
			continue
		}
		if hostMatchesApp(*a, host) {
			app = a
			break
		}
	}

	rules, err := store.ListSecurityRules(s.db)
	if err != nil || len(rules) == 0 {
		c.Status(http.StatusOK)
		return
	}

	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		scope := strings.ToLower(r.Scope)
		if scope == "global" || scope == "" {
			// ok
		} else if scope == "app" {
			if app == nil || r.AppID == "" || r.AppID != app.ID.String() {
				continue
			}
		} else {
			continue
		}
		if !ruleMatches(r, host, uri, path, query, method, ip, ua, c) {
			continue
		}
		store.BumpSecurityRuleHit(s.db, r.ID.String())
		switch strings.ToLower(r.Action) {
		case "allow":
			c.Status(http.StatusOK)
			return
		case "log":
			detail := method + " " + uri + " ip=" + ip
			store.AddAudit(s.db, "rules", "security-rule-log", r.Name, detail)
			c.Status(http.StatusOK)
			return
		default: // block
			c.Header("Content-Type", "text/plain")
			c.String(http.StatusForbidden, "blocked by security rule")
			return
		}
	}
	c.Status(http.StatusOK)
}

func hostMatchesApp(a models.Application, host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	if strings.EqualFold(a.Host, host) {
		return true
	}
	for _, h := range splitCSV(a.ExtraHosts) {
		if strings.EqualFold(h, host) {
			return true
		}
	}
	// www redirect pair (service may run on apex or www while Host field is the other)
	mode := strings.ToLower(strings.TrimSpace(a.WWWRedirect))
	if mode == "force_www" || mode == "force_apex" {
		h := strings.ToLower(strings.TrimSpace(a.Host))
		if strings.HasPrefix(h, "www.") && strings.TrimPrefix(h, "www.") == host {
			return true
		}
		if !strings.HasPrefix(h, "www.") && ("www."+h) == host {
			return true
		}
	}
	return false
}

func ruleMatches(r models.SecurityRule, host, uri, path, query, method, ip, ua string, c *gin.Context) bool {
	hay := ""
	switch strings.ToLower(r.Field) {
	case "uri":
		hay = uri
	case "path":
		hay = path
	case "query":
		hay = query
	case "method":
		hay = method
	case "host":
		hay = host
	case "ip":
		hay = ip
	case "user_agent", "user-agent", "ua":
		hay = ua
	case "header":
		name := strings.TrimSpace(r.HeaderName)
		if name == "" {
			return false
		}
		hay = c.GetHeader(name)
		if hay == "" {
			hay = c.Request.Header.Get(name)
		}
	default:
		return false
	}
	return matchOperator(r.Operator, hay, r.Value)
}

func matchOperator(op, hay, want string) bool {
	op = strings.ToLower(strings.TrimSpace(op))
	switch op {
	case "eq", "equals", "=":
		return strings.EqualFold(hay, want)
	case "contains":
		return strings.Contains(strings.ToLower(hay), strings.ToLower(want))
	case "starts_with", "prefix":
		return strings.HasPrefix(strings.ToLower(hay), strings.ToLower(want))
	case "ends_with", "suffix":
		return strings.HasSuffix(strings.ToLower(hay), strings.ToLower(want))
	case "in":
		for _, p := range splitCSV(want) {
			if strings.EqualFold(hay, p) {
				return true
			}
		}
		return false
	case "regex", "matches":
		re, err := regexp.Compile("(?i)" + want)
		if err != nil {
			return false
		}
		return re.MatchString(hay)
	default:
		return false
	}
}

func splitURI(uri string) (path, query string) {
	if i := strings.Index(uri, "?"); i >= 0 {
		return uri[:i], uri[i+1:]
	}
	return uri, ""
}

func splitCSV(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == ';'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func forwardedHost(c *gin.Context) string {
	host := firstNonEmpty(c.GetHeader("X-Forwarded-Host"), c.Request.Host)
	if i := strings.Index(host, ":"); i > 0 {
		host = host[:i]
	}
	return strings.ToLower(strings.TrimSpace(host))
}
