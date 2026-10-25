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
	return db.AutoMigrate(&models.User{}, &models.Application{}, &models.AuditEntry{})
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