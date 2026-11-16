// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/applogs"
	"github.com/overthecloud/the-gate/internal/models"
	"github.com/overthecloud/the-gate/internal/store"
)

type ipGroupDTO struct {
	models.IPGroup
	HitCount int `json:"hit_count"`
}

func (s *Server) listIPGroups(c *gin.Context) {
	groups, err := store.ListIPGroups(s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	entries, _ := s.logs.AccessEntries(8000)
	out := make([]ipGroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, ipGroupDTO{IPGroup: g, HitCount: hitCountForGroup(g, entries)})
	}
	c.JSON(http.StatusOK, out)
}

func (s *Server) createIPGroup(c *gin.Context) {
	var g models.IPGroup
	if err := c.BindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if err := validateIPGroup(g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := store.CreateIPGroup(s.db, &g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.EqualFold(g.Kind, "deny") {
		if err := s.syncDenyGroup(g, true); err != nil {
			store.AddAudit(s.db, c.GetString("user"), "crowdsec-warn", g.Name, err.Error())
			c.Header("X-Crowdsec-Warning", err.Error())
		}
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "create", "ip-group", g.Name)
	c.JSON(http.StatusCreated, g)
}

func (s *Server) updateIPGroup(c *gin.Context) {
	g, err := store.GetIPGroup(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if strings.EqualFold(g.Kind, "deny") {
		_ = s.syncDenyGroup(*g, false)
	}
	var body models.IPGroup
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	g.Name = body.Name
	g.Kind = body.Kind
	g.CIDRs = body.CIDRs
	g.Notes = body.Notes
	if err := validateIPGroup(*g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := store.UpdateIPGroup(s.db, g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strings.EqualFold(g.Kind, "deny") {
		if err := s.syncDenyGroup(*g, true); err != nil {
			store.AddAudit(s.db, c.GetString("user"), "crowdsec-warn", g.Name, err.Error())
			c.Header("X-Crowdsec-Warning", err.Error())
		}
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "update", "ip-group", g.Name)
	c.JSON(http.StatusOK, g)
}

func (s *Server) deleteIPGroup(c *gin.Context) {
	g, err := store.GetIPGroup(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if strings.EqualFold(g.Kind, "deny") {
		if err := s.syncDenyGroup(*g, false); err != nil {
			store.AddAudit(s.db, c.GetString("user"), "crowdsec-warn", g.Name, err.Error())
			c.Header("X-Crowdsec-Warning", err.Error())
		}
	}
	if err := store.DeleteIPGroup(s.db, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "delete", "ip-group", g.Name)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) syncDenyGroup(g models.IPGroup, add bool) error {
	var first error
	for _, value := range splitValues(g.CIDRs) {
		scope := "Ip"
		if strings.Contains(value, "/") {
			scope = "Range"
		}
		var err error
		if add {
			err = s.cs.AddDecision(scope, value, "ban", "87600h", "the-gate-deny-group")
		} else {
			err = s.cs.DeleteDecision(scope, value)
		}
		if err != nil && first == nil {
			first = err
		}
	}
	return first
}

func validateIPGroup(g models.IPGroup) error {
	if strings.TrimSpace(g.Name) == "" {
		return errText("name required")
	}
	kind := strings.ToLower(strings.TrimSpace(g.Kind))
	if kind != "allow" && kind != "deny" {
		return errText("kind must be allow or deny")
	}
	for _, value := range splitValues(g.CIDRs) {
		if strings.Contains(value, "/") {
			if _, _, err := net.ParseCIDR(value); err != nil {
				return errText("invalid CIDR: " + value)
			}
		} else if net.ParseIP(value) == nil {
			return errText("invalid IP: " + value)
		}
	}
	return nil
}

func splitValues(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' ' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hitCountForGroup(g models.IPGroup, entries []applogs.Entry) int {
	nets := []*net.IPNet{}
	ips := map[string]bool{}
	for _, value := range splitValues(g.CIDRs) {
		if strings.Contains(value, "/") {
			if _, n, err := net.ParseCIDR(value); err == nil {
				nets = append(nets, n)
			}
		} else if ip := net.ParseIP(value); ip != nil {
			ips[ip.String()] = true
		}
	}
	count := 0
	for _, e := range entries {
		client := strings.TrimSpace(e.Client)
		if host, _, err := net.SplitHostPort(client); err == nil {
			client = host
		}
		ip := net.ParseIP(client)
		if ip == nil {
			continue
		}
		if ips[ip.String()] {
			count++
			continue
		}
		for _, n := range nets {
			if n.Contains(ip) {
				count++
				break
			}
		}
	}
	return count
}

type errText string

func (e errText) Error() string { return string(e) }
