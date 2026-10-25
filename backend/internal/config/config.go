package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Listen            string
	DBURL             string
	AdminUser         string
	AdminPassword     string
	SessionSecret     string
	TraefikDynamicDir string
	BouncerKey        string
	CrowdsecLAPI      string
	CrowdsecLAPIKey   string

	// Authentik / OIDC for Gate admin UI
	AuthMode          string // local | oidc | both
	OIDCIssuer        string
	OIDCClientID      string
	OIDCClientSecret  string
	OIDCRedirectURL   string

	// Forward-auth URL for protecting applications (Authentik outpost)
	AuthentikForwardURL string

	// Protection feeds
	CorazaDir              string
	CrowdsecContainer      string
	AutoUpdateProtections  bool
	AutoUpdateHours        int
	PublicBaseURL          string // e.g. https://gate.example.com:9444 for OIDC redirects
	LogDir                 string // Traefik access/error logs
}

func FromEnv() Config {
	dbURL := strings.TrimSpace(os.Getenv("GATE_DB_URL"))
	if dbURL == "" {
		user := getenv("POSTGRES_USER", "gate")
		pass := mustEnv("POSTGRES_PASSWORD")
		host := getenv("POSTGRES_HOST", "postgres")
		name := getenv("POSTGRES_DB", "gate")
		u := &url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(user, pass),
			Host:     fmt.Sprintf("%s:5432", host),
			Path:     "/" + name,
			RawQuery: "sslmode=disable",
		}
		dbURL = u.String()
	}
	bouncer := mustEnv("BOUNCER_KEY_TRAEFIK")
	authMode := strings.ToLower(getenv("GATE_AUTH_MODE", "local"))
	if authMode != "local" && authMode != "oidc" && authMode != "both" {
		authMode = "local"
	}
	hours, _ := strconv.Atoi(getenv("GATE_AUTO_UPDATE_HOURS", "24"))
	if hours < 1 {
		hours = 24
	}
	return Config{
		Listen:                 getenv("GATE_LISTEN", ":8080"),
		DBURL:                  dbURL,
		AdminUser:              getenv("GATE_ADMIN_USER", "admin"),
		AdminPassword:          mustEnv("GATE_ADMIN_PASSWORD"),
		SessionSecret:          mustEnv("GATE_SESSION_SECRET"),
		TraefikDynamicDir:      getenv("GATE_TRAEFIK_DYNAMIC_DIR", "/traefik-dynamic"),
		BouncerKey:             bouncer,
		CrowdsecLAPI:           getenv("CROWDSEC_LAPI", "http://crowdsec:8080"),
		CrowdsecLAPIKey:        getenv("CROWDSEC_LAPI_KEY", bouncer),
		AuthMode:               authMode,
		OIDCIssuer:             getenv("GATE_OIDC_ISSUER", ""),
		OIDCClientID:           getenv("GATE_OIDC_CLIENT_ID", ""),
		OIDCClientSecret:       getenv("GATE_OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:        getenv("GATE_OIDC_REDIRECT_URL", ""),
		AuthentikForwardURL:    getenv("AUTHENTIK_FORWARD_URL", "http://authentik-server:9000/outpost.goauthentik.io/auth/traefik"),
		CorazaDir:              getenv("GATE_CORAZA_DIR", "/coraza"),
		CrowdsecContainer:      getenv("GATE_CROWDSEC_CONTAINER", "gate-crowdsec"),
		AutoUpdateProtections:  getenv("GATE_AUTO_UPDATE_PROTECTIONS", "false") == "true",
		AutoUpdateHours:        hours,
		PublicBaseURL:          strings.TrimRight(getenv("GATE_PUBLIC_URL", ""), "/"),
		LogDir:                 getenv("GATE_LOG_DIR", "/var/log/traefik"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustEnv(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" || strings.HasPrefix(v, "CHANGE_ME") {
		panic("missing or placeholder env: " + k)
	}
	return v
}

func (c Config) OIDCEnabled() bool {
	return (c.AuthMode == "oidc" || c.AuthMode == "both") &&
		c.OIDCIssuer != "" && c.OIDCClientID != "" && c.OIDCClientSecret != "" && c.OIDCRedirectURL != ""
}

func (c Config) LocalAuthEnabled() bool {
	return c.AuthMode == "local" || c.AuthMode == "both" || !c.OIDCEnabled()
}