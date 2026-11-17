// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/overthecloud/the-gate/internal/store"
)

const antibotCookie = "gate_ab"
const antibotTTL = 24 * time.Hour

// antibotCheck is Traefik forwardAuth for JS challenge (bots that don't run JS stay blocked).
func (s *Server) antibotCheck(c *gin.Context) {
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
		if !hostMatchesApp(a, host) {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(a.AntibotMode), "js") {
			need = true
		}
		break
	}
	if !need {
		c.Status(http.StatusOK)
		return
	}

	ip := clientIPFromRequest(c)
	cookie, _ := c.Cookie(antibotCookie)
	if cookie != "" && s.verifyAntibotToken(cookie, ip) {
		c.Status(http.StatusOK)
		return
	}

	token := s.issueAntibotToken(ip)
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><title>The Gate — Browser check</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
body{font-family:system-ui,sans-serif;background:#0b1220;color:#e5e7eb;display:grid;place-items:center;min-height:100vh;margin:0}
.card{max-width:420px;padding:28px;border:1px solid #1f2937;border-radius:16px;background:#111827}
h1{font-size:1.2rem;margin:0 0 8px}p{color:#9ca3af;margin:0 0 16px;line-height:1.5}
</style></head><body><div class="card">
<h1>Checking your browser</h1>
<p>Please wait a moment while The Gate verifies you are not an automated client.</p>
</div>
<script>
document.cookie = %q;
setTimeout(function(){ location.reload(); }, 1200);
</script></body></html>`, antibotCookie+"="+token+"; Path=/; Max-Age=86400; Secure; SameSite=Lax")

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusForbidden, html)
}

func (s *Server) issueAntibotToken(ip string) string {
	exp := time.Now().Add(antibotTTL).Unix()
	payload := fmt.Sprintf("%s|%d", ip, exp)
	mac := hmac.New(sha256.New, []byte(s.cfg.SessionSecret))
	_, _ = mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	raw := payload + "|" + sig
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func (s *Server) verifyAntibotToken(token, ip string) bool {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	parts := strings.Split(string(b), "|")
	if len(parts) != 3 {
		return false
	}
	if parts[0] != ip {
		return false
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	payload := parts[0] + "|" + parts[1]
	mac := hmac.New(sha256.New, []byte(s.cfg.SessionSecret))
	_, _ = mac.Write([]byte(payload))
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(parts[2]))
}
