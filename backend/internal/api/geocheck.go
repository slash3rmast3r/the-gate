// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/geoip"
	"github.com/overthecloud/the-gate/internal/store"
)

// geoCheck is called by Traefik forwardAuth for country blocking (internal docker network).
func (s *Server) geoCheck(c *gin.Context) {
	host := forwardedHost(c)
	ip := clientIPFromRequest(c)
	apps, err := store.ListApplications(s.db)
	if err != nil {
		c.Status(http.StatusOK) // fail open on control-plane errors
		return
	}
	var blocked string
	for _, a := range apps {
		if !a.Enabled {
			continue
		}
		if hostMatchesApp(a, host) {
			blocked = a.BlockedCountries
			break
		}
	}
	codes := parseCountryCodes(blocked)
	if len(codes) == 0 {
		c.Status(http.StatusOK)
		return
	}
	resolver := geoip.New(s.cfg.GeoIPDBPath)
	iso := strings.ToUpper(resolver.Country(ip))
	if iso == "" || iso == "UNKNOWN" || iso == "LOCAL" || iso == "XX" {
		// Without GeoIP DB we cannot enforce country blocks reliably — fail open.
		c.Status(http.StatusOK)
		return
	}
	for _, code := range codes {
		if iso == code {
			c.Header("Content-Type", "text/plain")
			c.String(http.StatusForbidden, "country blocked")
			return
		}
	}
	c.Status(http.StatusOK)
}

func clientIPFromRequest(c *gin.Context) string {
	for _, h := range []string{"X-Real-Ip", "X-Forwarded-For", "Cf-Connecting-Ip"} {
		v := strings.TrimSpace(c.GetHeader(h))
		if v == "" {
			continue
		}
		if h == "X-Forwarded-For" {
			parts := strings.Split(v, ",")
			v = strings.TrimSpace(parts[0])
		}
		if host, _, err := net.SplitHostPort(v); err == nil {
			return host
		}
		return v
	}
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	if ip != "" {
		return ip
	}
	return c.Request.RemoteAddr
}

func parseCountryCodes(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == ';'
	})
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.ToUpper(strings.TrimSpace(p))
		if len(p) != 2 || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}
