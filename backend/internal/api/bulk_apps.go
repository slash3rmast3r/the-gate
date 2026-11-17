// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/models"
	"github.com/overthecloud/the-gate/internal/store"
)

type bulkAppsBody struct {
	IDs []string               `json:"ids"`
	Set map[string]interface{} `json:"set"`
}

func (s *Server) bulkUpdateApps(c *gin.Context) {
	var body bulkAppsBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	if len(body.Set) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "set required — include at least one field to change"})
		return
	}

	updated := 0
	var names []string
	for _, id := range body.IDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		app, err := store.GetApplication(s.db, id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found: " + id})
			return
		}
		if err := applyBulkSet(app, body.Set); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "application_id": id})
			return
		}
		if strings.EqualFold(app.AuthMode, "basic") && (app.BasicAuthUser == "" || app.BasicAuthPass == "") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "cannot set auth_mode=basic in bulk without existing basic credentials — configure basic auth per-app first, or use none/authentik",
				"application_id": id,
				"host":           app.Host,
			})
			return
		}
		if err := store.UpdateApplication(s.db, app); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		updated++
		names = append(names, app.Host)
	}

	if !s.mustResync(c) {
		return
	}
	detail := fmt.Sprintf("updated=%d fields=%v hosts=%s", updated, keysOf(body.Set), strings.Join(names, ","))
	store.AddAudit(s.db, c.GetString("user"), "bulk-update", "applications", detail)
	c.JSON(http.StatusOK, gin.H{"ok": true, "updated": updated, "ids": body.IDs})
}

func applyBulkSet(app *models.Application, set map[string]interface{}) error {
	for key, raw := range set {
		switch key {
		case "allow_group_id":
			app.AllowGroupID = asString(raw)
		case "deny_group_id":
			app.DenyGroupID = asString(raw)
		case "auth_mode":
			v := strings.ToLower(asString(raw))
			switch v {
			case "none", "basic", "authentik":
				app.AuthMode = v
			default:
				return fmt.Errorf("invalid auth_mode")
			}
		case "rate_limit":
			v := strings.ToLower(asString(raw))
			switch v {
			case "soft", "strict", "dos", "off":
				app.RateLimit = v
			default:
				return fmt.Errorf("invalid rate_limit")
			}
		case "waf_mode":
			v := strings.ToLower(asString(raw))
			switch v {
			case "full", "light", "crowdsec-only", "off":
				app.WAFMode = v
			default:
				return fmt.Errorf("invalid waf_mode")
			}
		case "defense_mode":
			v := strings.ToLower(asString(raw))
			switch v {
			case "normal", "audit", "under_attack":
				app.DefenseMode = v
			default:
				return fmt.Errorf("invalid defense_mode")
			}
		case "blocked_countries":
			app.BlockedCountries = strings.ToUpper(asString(raw))
		case "http_flood":
			b, ok := asBool(raw)
			if !ok {
				return fmt.Errorf("http_flood must be boolean")
			}
			app.HTTPFlood = b
		case "bot_protect":
			b, ok := asBool(raw)
			if !ok {
				return fmt.Errorf("bot_protect must be boolean")
			}
			app.BotProtect = b
		case "attacks":
			b, ok := asBool(raw)
			if !ok {
				return fmt.Errorf("attacks must be boolean")
			}
			app.Attacks = b
		case "tls_enabled":
			b, ok := asBool(raw)
			if !ok {
				return fmt.Errorf("tls_enabled must be boolean")
			}
			app.TLSEnabled = b
		case "custom_errors":
			b, ok := asBool(raw)
			if !ok {
				return fmt.Errorf("custom_errors must be boolean")
			}
			app.CustomErrors = b
		case "enabled":
			b, ok := asBool(raw)
			if !ok {
				return fmt.Errorf("enabled must be boolean")
			}
			app.Enabled = b
		default:
			return fmt.Errorf("unsupported bulk field: %s", key)
		}
	}
	return nil
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func asBool(v interface{}) (bool, bool) {
	switch t := v.(type) {
	case bool:
		return t, true
	case string:
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "true", "1", "yes", "on":
			return true, true
		case "false", "0", "no", "off":
			return false, true
		}
	}
	return false, false
}

func keysOf(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
