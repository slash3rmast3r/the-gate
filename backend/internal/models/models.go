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
	WAFMode     string    `gorm:"size:32;default:full" json:"waf_mode"`       // full | light | crowdsec-only | off
	RateLimit   string    `gorm:"size:32;default:soft" json:"rate_limit"`     // soft | strict | dos | off
	TLSEnabled  bool      `gorm:"default:true" json:"tls_enabled"`
	PathPrefix  string    `gorm:"size:255;default:/" json:"path_prefix"`
	AllowCIDRs  string    `gorm:"size:1024" json:"allow_cidrs"`
	// AuthMode: none | basic | authentik (forward-auth outpost)
	AuthMode       string `gorm:"size:32;default:none" json:"auth_mode"`
	BasicAuthUser  string `gorm:"size:128" json:"basic_auth_user"`
	BasicAuthPass  string `gorm:"size:256" json:"basic_auth_pass,omitempty"` // write-only on update if empty keep
	BasicAuthRealm string `gorm:"size:128;default:The Gate" json:"basic_auth_realm"`
	CustomErrors   bool   `gorm:"default:true" json:"custom_errors"`
	Notes          string `gorm:"size:1024" json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// PublicApplication hides password hash material in list responses if needed.
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