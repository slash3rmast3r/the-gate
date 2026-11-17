// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/overthecloud/the-gate/internal/settings"
	"github.com/overthecloud/the-gate/internal/store"
)

func (s *Server) badBehaviorLoop() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for range t.C {
		s.runBadBehaviorSweep()
	}
}

func (s *Server) runBadBehaviorSweep() {
	ops := settings.LoadOps(s.db)
	bb := ops.BadBehavior
	if !bb.Enabled {
		return
	}
	entries, err := s.logs.AccessEntries(4000)
	if err != nil || len(entries) == 0 {
		return
	}
	cutoff := time.Now().UTC().Add(-time.Duration(bb.WindowSec) * time.Second)
	counts := map[string]int{}
	for _, e := range entries {
		if e.Status < bb.MinStatus {
			continue
		}
		ip := normalizeClientIP(e.Client)
		if ip == "" || isLocalOrPrivate(ip) {
			continue
		}
		if ts := parseLogTime(e.Time); !ts.IsZero() && ts.Before(cutoff) {
			continue
		}
		counts[ip]++
	}
	for ip, n := range counts {
		if n < bb.Threshold {
			continue
		}
		if err := s.cs.AddDecision("Ip", ip, "ban", bb.BanDuration, "the-gate-bad-behavior"); err != nil {
			log.Printf("bad-behavior ban %s: %v", ip, err)
			continue
		}
		store.AddAudit(s.db, "system", "bad-behavior-ban", ip, "count="+strconv.Itoa(n)+" window="+strconv.Itoa(bb.WindowSec)+"s")
	}
}

func normalizeClientIP(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		return host
	}
	return raw
}

func isLocalOrPrivate(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	return parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() || parsed.IsUnspecified()
}

func parseLogTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
