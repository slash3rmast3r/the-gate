// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/store"
)

func (s *Server) deleteDecision(c *gin.Context) {
	scope := c.Query("scope")
	value := c.Query("value")
	if value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value required"})
		return
	}
	if err := s.cs.DeleteDecision(scope, value); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "unblock", value, scope)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
