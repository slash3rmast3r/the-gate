package store

import (
	"errors"
	"strings"

	"github.com/overthecloud/the-gate/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{}, &models.Application{}, &models.AuditEntry{}, &models.Setting{}, &models.IPGroup{}, &models.SecurityRule{})
}

func EnsureAdmin(db *gorm.DB, user, pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	var existing models.User
	err = db.Where("username = ?", user).First(&existing).Error
	if err == nil {
		// Keep UI password in sync with GATE_ADMIN_PASSWORD on every boot
		existing.PasswordHash = string(hash)
		return db.Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&models.User{Username: user, PasswordHash: string(hash)}).Error
}

func Authenticate(db *gorm.DB, user, pass string) (*models.User, error) {
	var u models.User
	if err := db.Where("username = ?", user).First(&u).Error; err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)) != nil {
		return nil, errors.New("invalid credentials")
	}
	return &u, nil
}

func ListApplications(db *gorm.DB) ([]models.Application, error) {
	var apps []models.Application
	err := db.Order("name asc").Find(&apps).Error
	return apps, err
}

func GetApplication(db *gorm.DB, id string) (*models.Application, error) {
	var a models.Application
	if err := db.First(&a, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateApplication(db *gorm.DB, a *models.Application) error {
	a.Host = strings.ToLower(strings.TrimSpace(a.Host))
	a.UpstreamURL = strings.TrimSpace(a.UpstreamURL)
	if a.WAFMode == "" {
		a.WAFMode = "full"
	}
	if a.RateLimit == "" {
		a.RateLimit = "soft"
	}
	if strings.TrimSpace(a.PathPrefix) == "" {
		a.PathPrefix = "/"
	}
	if a.AuthMode == "" {
		a.AuthMode = "none"
	}
	if a.BasicAuthRealm == "" {
		a.BasicAuthRealm = "The Gate"
	}
	if a.DefenseMode == "" {
		a.DefenseMode = "normal"
	}
	if !a.HTTPFlood && !a.BotProtect && !a.Attacks {
		a.HTTPFlood = true
		a.BotProtect = true
		a.Attacks = true
	}
	return db.Create(a).Error
}

func UpdateApplication(db *gorm.DB, a *models.Application) error {
	a.Host = strings.ToLower(strings.TrimSpace(a.Host))
	a.UpstreamURL = strings.TrimSpace(a.UpstreamURL)
	return db.Save(a).Error
}

func DeleteApplication(db *gorm.DB, id string) error {
	return db.Delete(&models.Application{}, "id = ?", id).Error
}

func AddAudit(db *gorm.DB, actor, action, target, detail string) {
	_ = db.Create(&models.AuditEntry{
		Actor:  actor,
		Action: action,
		Target: target,
		Detail: detail,
	}).Error
}

func ListAudit(db *gorm.DB, limit int) ([]models.AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []models.AuditEntry
	err := db.Order("created_at desc").Limit(limit).Find(&rows).Error
	return rows, err
}

func ListIPGroups(db *gorm.DB) ([]models.IPGroup, error) {
	var groups []models.IPGroup
	err := db.Order("kind asc, name asc").Find(&groups).Error
	return groups, err
}

func GetIPGroup(db *gorm.DB, id string) (*models.IPGroup, error) {
	var g models.IPGroup
	if err := db.First(&g, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func CreateIPGroup(db *gorm.DB, g *models.IPGroup) error {
	normalizeIPGroup(g)
	return db.Create(g).Error
}

func UpdateIPGroup(db *gorm.DB, g *models.IPGroup) error {
	normalizeIPGroup(g)
	return db.Save(g).Error
}

func DeleteIPGroup(db *gorm.DB, id string) error {
	return db.Delete(&models.IPGroup{}, "id = ?", id).Error
}

func normalizeIPGroup(g *models.IPGroup) {
	g.Name = strings.TrimSpace(g.Name)
	g.Kind = strings.ToLower(strings.TrimSpace(g.Kind))
	g.CIDRs = strings.TrimSpace(g.CIDRs)
	g.Notes = strings.TrimSpace(g.Notes)
}

func ListSecurityRules(db *gorm.DB) ([]models.SecurityRule, error) {
	var rows []models.SecurityRule
	err := db.Order("priority asc, created_at asc").Find(&rows).Error
	return rows, err
}

func GetSecurityRule(db *gorm.DB, id string) (*models.SecurityRule, error) {
	var r models.SecurityRule
	if err := db.First(&r, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func CreateSecurityRule(db *gorm.DB, r *models.SecurityRule) error {
	normalizeSecurityRule(r)
	return db.Create(r).Error
}

func UpdateSecurityRule(db *gorm.DB, r *models.SecurityRule) error {
	normalizeSecurityRule(r)
	return db.Save(r).Error
}

func DeleteSecurityRule(db *gorm.DB, id string) error {
	return db.Delete(&models.SecurityRule{}, "id = ?", id).Error
}

func BumpSecurityRuleHit(db *gorm.DB, id string) {
	_ = db.Model(&models.SecurityRule{}).Where("id = ?", id).UpdateColumn("hit_count", gorm.Expr("hit_count + 1")).Error
}

func normalizeSecurityRule(r *models.SecurityRule) {
	r.Name = strings.TrimSpace(r.Name)
	r.Scope = strings.ToLower(strings.TrimSpace(r.Scope))
	r.AppID = strings.TrimSpace(r.AppID)
	r.Field = strings.ToLower(strings.TrimSpace(r.Field))
	r.HeaderName = strings.TrimSpace(r.HeaderName)
	r.Operator = strings.ToLower(strings.TrimSpace(r.Operator))
	r.Value = strings.TrimSpace(r.Value)
	r.Action = strings.ToLower(strings.TrimSpace(r.Action))
	r.Notes = strings.TrimSpace(r.Notes)
	if r.Scope == "" {
		r.Scope = "app"
	}
	if r.Priority <= 0 {
		r.Priority = 100
	}
}
