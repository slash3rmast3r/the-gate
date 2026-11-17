package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Application is a protected site/endpoint (SafeLine-like "Application").
type Application struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"size:128" json:"name"`
	Host        string    `gorm:"uniqueIndex;size:255" json:"host"`
	UpstreamURL string    `gorm:"size:512" json:"upstream_url"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	WAFMode     string    `gorm:"size:32;default:full" json:"waf_mode"`   // full | light | crowdsec-only | off
	RateLimit   string    `gorm:"size:32;default:soft" json:"rate_limit"` // soft | strict | dos | off
	TLSEnabled  bool      `gorm:"default:true" json:"tls_enabled"`
	PathPrefix  string    `gorm:"size:255;default:/" json:"path_prefix"`
	AllowCIDRs  string    `gorm:"size:1024" json:"allow_cidrs"`
	// AuthMode: none | basic | authentik
	AuthMode       string `gorm:"size:32;default:none" json:"auth_mode"`
	BasicAuthUser  string `gorm:"size:128" json:"basic_auth_user"`
	BasicAuthPass  string `gorm:"size:256" json:"basic_auth_pass,omitempty"`
	BasicAuthRealm string `gorm:"size:128;default:The Gate" json:"basic_auth_realm"`
	CustomErrors   bool   `gorm:"default:true" json:"custom_errors"`
	Notes          string `gorm:"size:1024" json:"notes"`

	DefenseMode     string `gorm:"size:32;default:normal" json:"defense_mode"` // normal | audit | under_attack
	HTTPFlood       bool   `gorm:"default:true" json:"http_flood"`
	BotProtect      bool   `gorm:"default:true" json:"bot_protect"`
	Attacks         bool   `gorm:"default:true" json:"attacks"`
	DefenseSnapshot string `gorm:"size:2048" json:"defense_snapshot,omitempty"`
	AllowGroupID     string `gorm:"size:36" json:"allow_group_id"`
	DenyGroupID      string `gorm:"size:36" json:"deny_group_id"`
	BlockedCountries string `gorm:"size:512" json:"blocked_countries"` // ISO codes CSV e.g. CN,RU,KP
	ExtraHosts       string `gorm:"size:1024" json:"extra_hosts"`       // extra FQDNs for LE SAN / routing (CSV)

	// Edge extras (BunkerWeb-inspired, Traefik-native)
	AntibotMode      string `gorm:"size:16;default:off" json:"antibot_mode"` // off | js
	DNSBLEnabled     bool   `gorm:"default:false" json:"dnsbl_enabled"`
	WWWRedirect      string `gorm:"size:16;default:off" json:"www_redirect"` // off | force_www | force_apex
	RequestHeaders   string `gorm:"type:text" json:"request_headers"`         // lines "Header: value"
	ResponseHeaders  string `gorm:"type:text" json:"response_headers"`
	WebsocketSticky  bool   `gorm:"default:false" json:"websocket_sticky"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SecurityRule is a Cloudflare-like conditional rule (global or per-application).
type SecurityRule struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name       string    `gorm:"size:128" json:"name"`
	Enabled    bool      `gorm:"default:true" json:"enabled"`
	Priority   int       `gorm:"default:100" json:"priority"` // lower runs first
	Scope      string    `gorm:"size:16;index;default:app" json:"scope"` // global | app
	AppID      string    `gorm:"size:36;index" json:"app_id"`            // empty when global
	Field      string    `gorm:"size:32" json:"field"`                   // uri|path|query|method|header|ip|host|user_agent
	HeaderName string    `gorm:"size:128" json:"header_name"`
	Operator   string    `gorm:"size:32" json:"operator"` // eq|contains|starts_with|ends_with|regex|in
	Value      string    `gorm:"size:1024" json:"value"`
	Action     string    `gorm:"size:16" json:"action"` // block|allow|log
	Notes      string    `gorm:"size:512" json:"notes"`
	HitCount   int64     `gorm:"default:0" json:"hit_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (r *SecurityRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.DefenseMode == "" {
		a.DefenseMode = "normal"
	}
	return nil
}

func (a Application) Redacted() Application {
	if a.BasicAuthPass != "" {
		a.BasicAuthPass = "********"
	}
	return a
}

type AuditEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Actor     string    `gorm:"size:64" json:"actor"`
	Action    string    `gorm:"size:64" json:"action"`
	Target    string    `gorm:"size:255" json:"target"`
	Detail    string    `gorm:"size:2048" json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *AuditEntry) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type Setting struct {
	Key       string    `gorm:"primaryKey;size:128" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IPGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:128" json:"name"`
	Kind      string    `gorm:"size:16;index" json:"kind"` // allow | deny
	CIDRs     string    `gorm:"type:text" json:"cidrs"`
	Notes     string    `gorm:"size:512" json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (g *IPGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}
