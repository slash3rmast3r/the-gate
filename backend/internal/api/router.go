// Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it
// SPDX-License-Identifier: MIT

package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/overthecloud/the-gate/internal/applogs"
	"github.com/overthecloud/the-gate/internal/authz"
	"github.com/overthecloud/the-gate/internal/config"
	"github.com/overthecloud/the-gate/internal/crowdsec"
	"github.com/overthecloud/the-gate/internal/models"
	"github.com/overthecloud/the-gate/internal/notify"
	"github.com/overthecloud/the-gate/internal/protections"
	"github.com/overthecloud/the-gate/internal/retention"
	"github.com/overthecloud/the-gate/internal/settings"
	"github.com/overthecloud/the-gate/internal/store"
	"github.com/overthecloud/the-gate/internal/traefiksync"
	"gorm.io/gorm"
)

type Server struct {
	cfg    config.Config
	db     *gorm.DB
	syncer *traefiksync.Syncer
	cs     *crowdsec.Client
	bf     *authz.BruteForce
	oidc   *authz.OIDC
	logs   *applogs.Reader
}

func NewRouter(cfg config.Config, db *gorm.DB, syncer *traefiksync.Syncer) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	s := &Server{
		cfg:    cfg,
		db:     db,
		syncer: syncer,
		cs:     crowdsec.New(cfg.CrowdsecLAPI, cfg.CrowdsecLAPIKey),
		bf:     authz.NewBruteForce(5, 15*time.Minute, 15*time.Minute),
		oidc:   authz.NewOIDC(cfg),
		logs:   applogs.New(cfg.LogDir),
	}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	storeCookie := cookie.NewStore([]byte(cfg.SessionSecret))
	storeCookie.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("gate_session", storeCookie))

	r.GET("/api/health", s.health)
	r.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":      "The Gate",
			"version":   "0.4.0",
			"author":    "Carlo Savino",
			"company":   "Over The Cloud",
			"website":   "https://overthecloud.it",
			"license":   "MIT",
			"copyright": "Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it",
		})
	})
	r.GET("/api/auth/options", s.authOptions)
	// Internal Traefik forwardAuth (docker network only; gate-api is not published)
	r.GET("/internal/geo-check", s.geoCheck)
	r.HEAD("/internal/geo-check", s.geoCheck)
	r.GET("/internal/rules-check", s.rulesCheck)
	r.HEAD("/internal/rules-check", s.rulesCheck)

	r.POST("/api/login", s.login)
	r.POST("/api/logout", s.logout)
	r.GET("/api/me", s.authRequired, s.me)
	r.GET("/api/auth/oidc/login", s.oidcLogin)
	r.GET("/api/auth/oidc/callback", s.oidcCallback)

	api := r.Group("/api")
	api.Use(s.authRequired)
	{
		api.GET("/applications", s.listApps)
		api.POST("/applications", s.createApp)
		api.GET("/applications/:id", s.getApp)
		api.PUT("/applications/:id", s.updateApp)
		api.DELETE("/applications/:id", s.deleteApp)
		api.POST("/applications/:id/toggle", s.toggleApp)
		api.POST("/applications/:id/defense", s.setDefenseMode)
		api.GET("/applications/:id/logs", s.appLogs)

		api.GET("/decisions", s.listDecisions)
		api.DELETE("/decisions", s.deleteDecision)
		api.GET("/audit", s.listAudit)
		api.GET("/dashboard", s.dashboard)
		api.POST("/sync", s.forceSync)
		api.GET("/export", s.exportApps)
		api.POST("/import", s.importApps)
		api.GET("/settings", s.getSettings)
		api.PUT("/settings", s.putSettings)
		api.POST("/backup", s.backup)
		api.GET("/ip-groups", s.listIPGroups)
		api.POST("/ip-groups", s.createIPGroup)
		api.PUT("/ip-groups/:id", s.updateIPGroup)
		api.DELETE("/ip-groups/:id", s.deleteIPGroup)
		api.GET("/security-rules", s.listSecurityRules)
		api.POST("/security-rules", s.createSecurityRule)
		api.PUT("/security-rules/:id", s.updateSecurityRule)
		api.DELETE("/security-rules/:id", s.deleteSecurityRule)
		api.GET("/certificates", s.listCertificates)
		api.POST("/certificates/renew", s.renewCertificate)
		api.GET("/traffic", s.traffic)
		api.POST("/notifications/test", s.testNotification)
		api.GET("/errorpages", s.listErrorPages)
		api.POST("/errorpages/theme", s.applyErrorTheme)
		api.GET("/errorpages/:code", s.getErrorPage)
		api.PUT("/errorpages/:code", s.putErrorPage)

		api.POST("/protections/update-crs", s.updateCRS)
		api.POST("/protections/update-crowdsec", s.updateCrowdSec)
		api.GET("/protections/status", s.protectionsStatus)
	}

	go s.autoUpdateLoop()
	go s.retentionLoop()
	return r
}

func (s *Server) autoUpdateLoop() {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	last := time.Now()
	for range t.C {
		ops := settings.LoadOps(s.db)
		if !ops.AutoUpdateEnabled {
			continue
		}
		if ops.AutoUpdateHours < 1 {
			ops.AutoUpdateHours = s.cfg.AutoUpdateHours
		}
		if time.Since(last) < time.Duration(ops.AutoUpdateHours)*time.Hour {
			continue
		}
		last = time.Now()
		if msg, err := protections.UpdateCRS(s.cfg.CorazaDir, ""); err != nil {
			notify.NotifyAll(ops, "The Gate auto-update failed", "OWASP CRS: "+err.Error())
			store.AddAudit(s.db, "system", "auto-update", "coraza", err.Error())
		} else {
			store.AddAudit(s.db, "system", "auto-update", "coraza", msg)
		}
		if msg, err := protections.UpdateCrowdSecHub(s.cfg.CrowdsecContainer); err != nil {
			notify.NotifyAll(ops, "The Gate auto-update failed", "CrowdSec Hub: "+err.Error())
			store.AddAudit(s.db, "system", "auto-update", "crowdsec", err.Error())
		} else {
			store.AddAudit(s.db, "system", "auto-update", "crowdsec", msg)
		}
		_ = s.resync()
	}
}

func (s *Server) retentionLoop() {
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	for {
		ops := settings.LoadOps(s.db)
		_, _ = retention.PruneLogs(s.cfg.LogDir, ops.LogRetentionDays)
		_, _ = retention.PruneAudit(s.db, ops.LogRetentionDays)
		<-t.C
	}
}

func (s *Server) health(c *gin.Context) {
	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "db": "down"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "up"})
}

func (s *Server) authOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"local":          s.cfg.LocalAuthEnabled(),
		"oidc":           s.cfg.OIDCEnabled(),
		"auth_mode":      s.cfg.AuthMode,
		"authentik_apps": s.cfg.AuthentikForwardURL != "",
	})
}

func (s *Server) authRequired(c *gin.Context) {
	sess := sessions.Default(c)
	u := sess.Get("user")
	if u == nil || u == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}
	c.Set("user", u.(string))
	c.Next()
}

func clientKey(c *gin.Context) string {
	ip := c.ClientIP()
	return ip
}

func (s *Server) login(c *gin.Context) {
	if !s.cfg.LocalAuthEnabled() {
		c.JSON(http.StatusForbidden, gin.H{"error": "local login disabled — use Authentik SSO"})
		return
	}
	key := clientKey(c)
	if ok, wait := s.bf.Allowed(key); !ok {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": fmt.Sprintf("too many failed attempts — retry in %s", wait.Round(time.Second)),
		})
		return
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	u, err := store.Authenticate(s.db, body.Username, body.Password)
	if err != nil {
		s.bf.Fail(key)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	s.bf.Success(key)
	sess := sessions.Default(c)
	sess.Set("user", u.Username)
	_ = sess.Save()
	store.AddAudit(s.db, u.Username, "login", "session", "local")
	c.JSON(http.StatusOK, gin.H{"username": u.Username})
}

func (s *Server) logout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"username": c.GetString("user")})
}

func (s *Server) oidcLogin(c *gin.Context) {
	if !s.cfg.OIDCEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "OIDC not configured"})
		return
	}
	u, err := s.oidc.LoginURL()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, u)
}

func (s *Server) oidcCallback(c *gin.Context) {
	if !s.cfg.OIDCEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "OIDC not configured"})
		return
	}
	state := c.Query("state")
	code := c.Query("code")
	if code == "" || !s.oidc.ConsumeState(state) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}
	user, err := s.oidc.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	sess := sessions.Default(c)
	sess.Set("user", user)
	_ = sess.Save()
	store.AddAudit(s.db, user, "login", "session", "oidc")
	c.Redirect(http.StatusFound, "/")
}

func (s *Server) resync() error {
	if err := s.syncer.WriteBaseMiddlewares(); err != nil {
		return err
	}
	apps, err := store.ListApplications(s.db)
	if err != nil {
		return err
	}
	groups, _ := store.ListIPGroups(s.db)
	return s.syncer.SyncApplications(apps, groups)
}

func (s *Server) mustResync(c *gin.Context) bool {
	if err := s.resync(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "saved but traefik sync failed: " + err.Error()})
		return false
	}
	return true
}

func redactList(apps []models.Application) []models.Application {
	out := make([]models.Application, len(apps))
	for i, a := range apps {
		out[i] = a.Redacted()
	}
	return out
}

func (s *Server) listApps(c *gin.Context) {
	apps, err := store.ListApplications(s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, redactList(apps))
}

func (s *Server) getApp(c *gin.Context) {
	a, err := store.GetApplication(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a.Redacted())
}

func (s *Server) createApp(c *gin.Context) {
	var a models.Application
	if err := c.BindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	if a.Name == "" || a.Host == "" || a.UpstreamURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, host, upstream_url required"})
		return
	}
	if !strings.HasPrefix(a.UpstreamURL, "http://") && !strings.HasPrefix(a.UpstreamURL, "https://") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "upstream_url must start with http:// or https://"})
		return
	}
	if strings.EqualFold(a.AuthMode, "basic") && (a.BasicAuthUser == "" || a.BasicAuthPass == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basic auth requires user and password"})
		return
	}
	if a.AuthMode == "" {
		a.AuthMode = "none"
	}
	if err := store.CreateApplication(s.db, &a); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "create", a.Host, a.UpstreamURL)
	c.JSON(http.StatusCreated, a.Redacted())
}

func (s *Server) updateApp(c *gin.Context) {
	existing, err := store.GetApplication(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body models.Application
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	existing.Name = body.Name
	existing.Host = body.Host
	existing.UpstreamURL = body.UpstreamURL
	existing.Enabled = body.Enabled
	existing.WAFMode = body.WAFMode
	existing.RateLimit = body.RateLimit
	existing.TLSEnabled = body.TLSEnabled
	existing.PathPrefix = body.PathPrefix
	existing.AllowCIDRs = body.AllowCIDRs
	existing.AuthMode = body.AuthMode
	existing.BasicAuthUser = body.BasicAuthUser
	existing.BasicAuthRealm = body.BasicAuthRealm
	existing.CustomErrors = body.CustomErrors
	existing.Notes = body.Notes
	existing.DefenseMode = body.DefenseMode
	existing.HTTPFlood = body.HTTPFlood
	existing.BotProtect = body.BotProtect
	existing.Attacks = body.Attacks
	existing.AllowGroupID = body.AllowGroupID
	existing.DenyGroupID = body.DenyGroupID
	existing.BlockedCountries = body.BlockedCountries
	existing.ExtraHosts = body.ExtraHosts
	if body.BasicAuthPass != "" && body.BasicAuthPass != "********" {
		existing.BasicAuthPass = body.BasicAuthPass
	}
	if strings.EqualFold(existing.AuthMode, "basic") && (existing.BasicAuthUser == "" || existing.BasicAuthPass == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "basic auth requires user and password"})
		return
	}
	if err := store.UpdateApplication(s.db, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "update", existing.Host, existing.UpstreamURL)
	c.JSON(http.StatusOK, existing.Redacted())
}

func (s *Server) deleteApp(c *gin.Context) {
	existing, err := store.GetApplication(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := store.DeleteApplication(s.db, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "delete", existing.Host, "")
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) appLogs(c *gin.Context) {
	a, err := store.GetApplication(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	kind := c.DefaultQuery("kind", "access") // access | error
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	q := c.Query("q")
	entries, err := s.logs.ForHost(a.Host, kind, limit, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"application_id": a.ID,
		"host":           a.Host,
		"kind":           kind,
		"count":          len(entries),
		"entries":        entries,
	})
}

func (s *Server) toggleApp(c *gin.Context) {
	existing, err := store.GetApplication(s.db, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	existing.Enabled = !existing.Enabled
	if err := store.UpdateApplication(s.db, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "toggle", existing.Host, boolStr(existing.Enabled))
	c.JSON(http.StatusOK, existing.Redacted())
}

func (s *Server) listDecisions(c *gin.Context) {
	d, err := s.cs.Decisions()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"decisions": []any{}, "warning": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"decisions": d})
}

func (s *Server) listAudit(c *gin.Context) {
	rows, err := store.ListAudit(s.db, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (s *Server) dashboard(c *gin.Context) {
	apps, _ := store.ListApplications(s.db)
	enabled := 0
	for _, a := range apps {
		if a.Enabled {
			enabled++
		}
	}
	dec, _ := s.cs.Decisions()
	c.JSON(http.StatusOK, gin.H{
		"applications_total":   len(apps),
		"applications_enabled": enabled,
		"active_decisions":     len(dec),
		"product":              "The Gate",
		"version":              "0.4.0",
		"author":               "Carlo Savino",
		"company":              "Over The Cloud",
		"website":              "https://overthecloud.it",
		"license":              "MIT",
		"copyright":            "Copyright (c) 2026 Carlo Savino — Over The Cloud — https://overthecloud.it",
		"oidc_enabled":         s.cfg.OIDCEnabled(),
	})
}

func (s *Server) forceSync(c *gin.Context) {
	if !s.mustResync(c) {
		return
	}
	apps, _ := store.ListApplications(s.db)
	written := 0
	for _, a := range apps {
		if a.Enabled {
			written++
		}
	}
	store.AddAudit(s.db, c.GetString("user"), "sync", "traefik", "ok")
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"written": written,
		"message": fmt.Sprintf("Traefik config synced (%d enabled application(s)).", written),
	})
}

func (s *Server) exportApps(c *gin.Context) {
	apps, err := store.ListApplications(s.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=the-gate-apps.json")
	c.JSON(http.StatusOK, redactList(apps))
}

func (s *Server) importApps(c *gin.Context) {
	var apps []models.Application
	if err := c.BindJSON(&apps); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected JSON array of applications"})
		return
	}
	created := 0
	for _, src := range apps {
		if src.Name == "" || src.Host == "" || src.UpstreamURL == "" {
			continue
		}
		a := models.Application{
			ID:             uuid.Nil,
			Name:           src.Name,
			Host:           src.Host,
			UpstreamURL:    src.UpstreamURL,
			Enabled:        src.Enabled,
			WAFMode:        src.WAFMode,
			RateLimit:      src.RateLimit,
			TLSEnabled:     src.TLSEnabled,
			PathPrefix:     src.PathPrefix,
			AllowCIDRs:     src.AllowCIDRs,
			AuthMode:       src.AuthMode,
			BasicAuthUser:  src.BasicAuthUser,
			BasicAuthPass:  src.BasicAuthPass,
			BasicAuthRealm: src.BasicAuthRealm,
			CustomErrors:   src.CustomErrors,
			Notes:          src.Notes,
			DefenseMode:    src.DefenseMode,
			HTTPFlood:      src.HTTPFlood,
			BotProtect:     src.BotProtect,
			Attacks:        src.Attacks,
			AllowGroupID:   src.AllowGroupID,
			DenyGroupID:      src.DenyGroupID,
			BlockedCountries: src.BlockedCountries,
			ExtraHosts:       src.ExtraHosts,
		}
		if a.BasicAuthPass == "********" {
			a.BasicAuthPass = ""
		}
		if err := store.CreateApplication(s.db, &a); err != nil {
			continue
		}
		created++
	}
	if !s.mustResync(c) {
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "import", "applications", fmt.Sprintf("%d", created))
	c.JSON(http.StatusOK, gin.H{"imported": created})
}

func (s *Server) updateCRS(c *gin.Context) {
	msg, err := protections.UpdateCRS(s.cfg.CorazaDir, c.Query("version"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "update-crs", "coraza", msg)
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": msg})
}

func (s *Server) updateCrowdSec(c *gin.Context) {
	msg, err := protections.UpdateCrowdSecHub(s.cfg.CrowdsecContainer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "detail": msg})
		return
	}
	store.AddAudit(s.db, c.GetString("user"), "update-crowdsec", "hub", msg)
	c.JSON(http.StatusOK, gin.H{"ok": true, "message": msg})
}

func (s *Server) protectionsStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"crs_dir":               s.cfg.CorazaDir,
		"crowdsec_container":    s.cfg.CrowdsecContainer,
		"auto_update":           s.cfg.AutoUpdateProtections,
		"auto_update_hours":     s.cfg.AutoUpdateHours,
		"authentik_forward_url": s.cfg.AuthentikForwardURL,
		"oidc_admin":            s.cfg.OIDCEnabled(),
	})
}

func boolStr(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
