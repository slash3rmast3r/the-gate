// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/notify"
	"github.com/overthecloud/the-gate/internal/settings"
	"github.com/overthecloud/the-gate/internal/store"
)

type defenseSnapshot struct {
	DefenseMode string `json:"defense_mode"`
	WAFMode     string `json:"waf_mode"`
	RateLimit   string `json:"rate_limit"`
	HTTPFlood   bool   `json:"http_flood"`
	BotProtect  bool   `json:"bot_protect"`
	Attacks     bool   `json:"attacks"`
}

func (s *Server) setDefenseMode(c *gin.Context) {
	app, err := store.GetApplication(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		Mode string `json:"mode"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	mode := strings.ToLower(strings.TrimSpace(body.Mode))
	if mode != "normal" && mode != "audit" && mode != "under_attack" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be normal, audit, or under_attack"})
		return
	}

	if mode == "under_attack" && app.DefenseMode != "under_attack" {
		snap := defenseSnapshot{
			DefenseMode: app.DefenseMode,
			WAFMode:     app.WAFMode,
			RateLimit:   app.RateLimit,
			HTTPFlood:   app.HTTPFlood,
			BotProtect:  app.BotProtect,
			Attacks:     app.Attacks,
		}
		b, _ := json.Marshal(snap)
		app.DefenseSnapshot = string(b)
		app.WAFMode = "full"
		app.RateLimit = "dos"
		app.HTTPFlood = true
		app.BotProtect = true
		app.Attacks = true
		notify.NotifyAll(settings.LoadOps(s.db), "The Gate under attack mode enabled", app.Name+" ("+app.Host+") is now in under attack mode.")
	} else if app.DefenseMode == "under_attack" && mode != "under_attack" && app.DefenseSnapshot != "" {
		var snap defenseSnapshot
		if json.Unmarshal([]byte(app.DefenseSnapshot), &snap) == nil {
			app.WAFMode = snap.WAFMode
			app.RateLimit = snap.RateLimit
			app.HTTPFlood = snap.HTTPFlood
			app.BotProtect = snap.BotProtect
			app.Attacks = snap.Attacks
		}
		app.DefenseSnapshot = ""
	}
	app.DefenseMode = mode
	if err := store.UpdateApplication(s.db, app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "defense", app.Host, mode)
	c.JSON(http.StatusOK, app.Redacted())
}
