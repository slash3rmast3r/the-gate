// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/acme"
	"github.com/overthecloud/the-gate/internal/store"
)

func (s *Server) listCertificates(c *gin.Context) {
	certs, err := acme.List(s.cfg.ACMEStoragePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	apps, _ := store.ListApplications(s.db)
	for i := range certs {
		for _, a := range apps {
			if certBelongsToApp(certs[i], a.Host, a.ExtraHosts) {
				certs[i].AppID = a.ID.String()
				certs[i].AppName = a.Name
				certs[i].AppHost = a.Host
				break
			}
		}
	}
	if certs == nil {
		certs = []acme.CertInfo{}
	}
	c.JSON(http.StatusOK, gin.H{
		"certificates": certs,
		"storage":      s.cfg.ACMEStoragePath,
		"note":         "Let's Encrypt issuance and renewal are handled automatically by Traefik. Force renew removes the stored cert so Traefik re-requests it.",
	})
}

func (s *Server) renewCertificate(c *gin.Context) {
	var body struct {
		Domain string `json:"domain"`
	}
	if err := c.BindJSON(&body); err != nil || strings.TrimSpace(body.Domain) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain required"})
		return
	}
	removed, err := acme.ForceRenew(s.cfg.ACMEStoragePath, body.Domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "cert-renew", body.Domain, "removed entries so Traefik re-issues")
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"domain":  strings.ToLower(strings.TrimSpace(body.Domain)),
		"removed": removed,
		"message": "Certificate entry removed. Traefik will re-issue on next request to the host.",
	})
}

func certBelongsToApp(cert acme.CertInfo, host, extra string) bool {
	wanted := map[string]bool{strings.ToLower(strings.TrimSpace(host)): true}
	for _, h := range splitCSV(extra) {
		wanted[strings.ToLower(h)] = true
	}
	if wanted[cert.Main] {
		return true
	}
	for _, d := range cert.Domains {
		if wanted[strings.ToLower(d)] {
			return true
		}
	}
	return false
}
