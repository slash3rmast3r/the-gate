// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/errorpages"
	"github.com/overthecloud/the-gate/internal/settings"
	"github.com/overthecloud/the-gate/internal/store"
)

var editableErrorCodes = map[string]bool{
	"401": true,
	"403": true,
	"404": true,
	"429": true,
	"500": true,
	"502": true,
	"503": true,
}

func (s *Server) listErrorPages(c *gin.Context) {
	codes := make([]string, 0, len(editableErrorCodes))
	for code := range editableErrorCodes {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	pages := map[string]string{}
	for _, code := range codes {
		html, _ := errorpages.LoadTemplate(s.cfg.ErrorPagesDir, code)
		pages[code] = html
	}
	ops := settings.LoadOps(s.db)
	c.JSON(http.StatusOK, gin.H{
		"codes":       codes,
		"pages":       pages,
		"tip_message": ops.ErrorTipMessage,
		"theme":       ops.ErrorTheme,
		"themes":      errorpages.Themes,
		"placeholder": errorpages.TipPlaceholder,
	})
}

func (s *Server) applyErrorTheme(c *gin.Context) {
	var body struct {
		Theme string `json:"theme"`
	}
	if err := c.BindJSON(&body); err != nil || strings.TrimSpace(body.Theme) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "theme required"})
		return
	}
	ops := settings.LoadOps(s.db)
	if err := errorpages.ApplyThemeAll(s.cfg.ErrorPagesDir, body.Theme, ops.ErrorTipMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = settings.Set(s.db, settings.KeyErrorTheme, body.Theme)
	store.AddAudit(s.db, c.GetString("user"), "error-theme", body.Theme, "applied")
	c.JSON(http.StatusOK, gin.H{"ok": true, "theme": body.Theme})
}

func (s *Server) getErrorPage(c *gin.Context) {
	code := c.Param("code")
	if !editableErrorCodes[code] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported error page code"})
		return
	}
	html, err := errorpages.LoadTemplate(s.cfg.ErrorPagesDir, code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": code, "html": ""})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": code, "html": html})
}

func (s *Server) putErrorPage(c *gin.Context) {
	code := c.Param("code")
	if !editableErrorCodes[code] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported error page code"})
		return
	}
	var body struct {
		HTML string `json:"html"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	tip := settings.LoadOps(s.db).ErrorTipMessage
	if err := errorpages.SaveTemplate(s.cfg.ErrorPagesDir, code, body.HTML, tip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "errorpage", code, "updated")
	c.JSON(http.StatusOK, gin.H{"ok": true, "code": code})
}
