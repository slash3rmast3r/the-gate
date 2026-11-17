// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/settings"
	"github.com/overthecloud/the-gate/internal/store"
)

func (s *Server) dnsblCheck(c *gin.Context) {
	host := forwardedHost(c)
	apps, err := store.ListApplications(s.db)
	if err != nil {
		c.Status(http.StatusOK)
		return
	}
	need := false
	for _, a := range apps {
		if !a.Enabled {
			continue
		}
		if hostMatchesApp(a, host) {
			need = a.DNSBLEnabled
			break
		}
	}
	if !need {
		c.Status(http.StatusOK)
		return
	}

	ip := clientIPFromRequest(c)
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() {
		c.Status(http.StatusOK)
		return
	}
	v4 := parsed.To4()
	if v4 == nil {
		c.Status(http.StatusOK)
		return
	}

	ops := settings.LoadOps(s.db)
	zones := splitCSV(ops.DNSBLZones)
	if len(zones) == 0 {
		c.Status(http.StatusOK)
		return
	}
	rev := fmt.Sprintf("%d.%d.%d.%d", v4[3], v4[2], v4[1], v4[0])
	resolver := net.DefaultResolver
	ctx, cancel := context.WithTimeout(c.Request.Context(), 1500*time.Millisecond)
	defer cancel()
	for _, zone := range zones {
		zone = strings.TrimSpace(strings.ToLower(zone))
		if zone == "" {
			continue
		}
		q := rev + "." + zone
		addrs, err := resolver.LookupHost(ctx, q)
		if err != nil || len(addrs) == 0 {
			continue
		}
		c.Header("Content-Type", "text/plain")
		c.String(http.StatusForbidden, "blocked by dnsbl")
		return
	}
	c.Status(http.StatusOK)
}
