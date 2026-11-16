// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/backup"
	"github.com/overthecloud/the-gate/internal/errorpages"
	"github.com/overthecloud/the-gate/internal/geoip"
	"github.com/overthecloud/the-gate/internal/notify"
	"github.com/overthecloud/the-gate/internal/settings"
	"github.com/overthecloud/the-gate/internal/store"
	"github.com/overthecloud/the-gate/internal/traffic"
)

func (s *Server) getSettings(c *gin.Context) {
	c.JSON(http.StatusOK, settings.RedactOps(settings.LoadOps(s.db)))
}

func (s *Server) putSettings(c *gin.Context) {
	var ops settings.OpsConfig
	if err := c.BindJSON(&ops); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	current := settings.LoadOps(s.db)
	if err := settings.SaveOps(s.db, ops, current.Telegram.BotToken, current.SMTP.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = errorpages.RenderAll(s.cfg.ErrorPagesDir, ops.ErrorTipMessage)
	store.AddAudit(s.db, c.GetString("user"), "settings", "ops", "updated")
	c.JSON(http.StatusOK, settings.RedactOps(settings.LoadOps(s.db)))
}

func (s *Server) backup(c *gin.Context) {
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", `attachment; filename="`+backup.Filename()+`"`)
	if err := backup.WriteZip(c.Writer, s.db, s.cfg.TraefikDynamicDir); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "backup", "control-plane", "zip")
}

func (s *Server) testNotification(c *gin.Context) {
	errs := notify.NotifyAll(settings.LoadOps(s.db), "The Gate notification test", "Notifications are configured and reachable.")
	if len(errs) > 0 {
		c.JSON(http.StatusBadGateway, gin.H{"ok": false, "errors": errs})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) traffic(c *gin.Context) {
	minutes, _ := strconv.Atoi(c.DefaultQuery("minutes", "60"))
	entries, err := s.logs.AccessEntries(20000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	apps, err := store.ListApplications(s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, traffic.Analyze(apps, entries, geoip.New(s.cfg.GeoIPDBPath), minutes))
}
