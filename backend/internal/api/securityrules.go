// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/models"
	"github.com/overthecloud/the-gate/internal/store"
)

func (s *Server) listSecurityRules(c *gin.Context) {
	rows, err := store.ListSecurityRules(s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []models.SecurityRule{}
	}
	c.JSON(http.StatusOK, rows)
}

func (s *Server) createSecurityRule(c *gin.Context) {
	var r models.SecurityRule
	if err := c.BindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if err := validateSecurityRule(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := store.CreateSecurityRule(s.db, &r); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "create", "security-rule", r.Name)
	c.JSON(http.StatusCreated, r)
}

func (s *Server) updateSecurityRule(c *gin.Context) {
	existing, err := store.GetSecurityRule(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body models.SecurityRule
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	existing.Name = body.Name
	existing.Enabled = body.Enabled
	existing.Priority = body.Priority
	existing.Scope = body.Scope
	existing.AppID = body.AppID
	existing.Field = body.Field
	existing.HeaderName = body.HeaderName
	existing.Operator = body.Operator
	existing.Value = body.Value
	existing.Action = body.Action
	existing.Notes = body.Notes
	if err := validateSecurityRule(existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := store.UpdateSecurityRule(s.db, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "update", "security-rule", existing.Name)
	c.JSON(http.StatusOK, existing)
}

func (s *Server) deleteSecurityRule(c *gin.Context) {
	existing, err := store.GetSecurityRule(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := store.DeleteSecurityRule(s.db, existing.ID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "delete", "security-rule", existing.Name)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func validateSecurityRule(r *models.SecurityRule) error {
	if strings.TrimSpace(r.Name) == "" {
		return errMsg("name required")
	}
	scope := strings.ToLower(strings.TrimSpace(r.Scope))
	if scope != "global" && scope != "app" {
		return errMsg("scope must be global or app")
	}
	r.Scope = scope
	if scope == "app" && strings.TrimSpace(r.AppID) == "" {
		return errMsg("app_id required for app scope")
	}
	if scope == "global" {
		r.AppID = ""
	}
	field := strings.ToLower(strings.TrimSpace(r.Field))
	switch field {
	case "uri", "path", "query", "method", "header", "ip", "host", "user_agent":
	default:
		return errMsg("invalid field")
	}
	r.Field = field
	if field == "header" && strings.TrimSpace(r.HeaderName) == "" {
		return errMsg("header_name required when field=header")
	}
	op := strings.ToLower(strings.TrimSpace(r.Operator))
	switch op {
	case "eq", "contains", "starts_with", "ends_with", "regex", "in":
	default:
		return errMsg("invalid operator")
	}
	r.Operator = op
	if strings.TrimSpace(r.Value) == "" {
		return errMsg("value required")
	}
	action := strings.ToLower(strings.TrimSpace(r.Action))
	switch action {
	case "block", "allow", "log":
	default:
		return errMsg("action must be block, allow, or log")
	}
	r.Action = action
	if r.Priority <= 0 {
		r.Priority = 100
	}
	return nil
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }

func errMsg(s string) error { return simpleErr(s) }
