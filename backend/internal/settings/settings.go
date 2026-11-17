package settings

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/overthecloud/the-gate/internal/models"
	"gorm.io/gorm"
)

const (
	KeyAutoUpdateEnabled = "auto_update_enabled"
	KeyAutoUpdateHours   = "auto_update_hours"
	KeyLogRetentionDays  = "log_retention_days"
	KeyErrorTipMessage   = "error_tip_message"
	KeyErrorTheme        = "error_theme"
	KeyNotifyTelegram    = "notify_telegram"
	KeyNotifySMTP        = "notify_smtp"
	KeyBadBehavior       = "bad_behavior"
	KeyDNSBLZones        = "dnsbl_zones"
)

type TelegramConfig struct {
	Enabled  bool   `json:"enabled"`
	BotToken string `json:"bot_token"`
	ChatID   string `json:"chat_id"`
}

type SMTPConfig struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type BadBehaviorConfig struct {
	Enabled     bool   `json:"enabled"`
	Threshold   int    `json:"threshold"`    // errors per IP in window
	WindowSec   int    `json:"window_sec"`   // lookback window
	MinStatus   int    `json:"min_status"`   // count statuses >= this (default 400)
	BanDuration string `json:"ban_duration"` // CrowdSec duration e.g. 4h
}

type OpsConfig struct {
	AutoUpdateEnabled bool   `json:"auto_update_enabled"`
	AutoUpdateHours   int    `json:"auto_update_hours"`
	LogRetentionDays  int    `json:"log_retention_days"`
	ErrorTipMessage   string         `json:"error_tip_message"`
	ErrorTheme        string         `json:"error_theme"`
	Telegram          TelegramConfig `json:"telegram"`
	SMTP              SMTPConfig     `json:"smtp"`
	BadBehavior       BadBehaviorConfig `json:"bad_behavior"`
	DNSBLZones        string            `json:"dnsbl_zones"` // CSV of DNSBL zones
}

func Get(db *gorm.DB, key, def string) string {
	var s models.Setting
	if err := db.First(&s, "key = ?", key).Error; err != nil {
		return def
	}
	return s.Value
}

func Set(db *gorm.DB, key, value string) error {
	s := models.Setting{Key: key, Value: value, UpdatedAt: time.Now()}
	return db.Save(&s).Error
}

func GetBool(db *gorm.DB, key string, def bool) bool {
	v := Get(db, key, "")
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func GetInt(db *gorm.DB, key string, def int) int {
	v := Get(db, key, "")
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func EnsureDefaults(db *gorm.DB, autoEnabled bool, autoHours, retentionDays int) error {
	if Get(db, KeyAutoUpdateEnabled, "") == "" {
		_ = Set(db, KeyAutoUpdateEnabled, strconv.FormatBool(autoEnabled))
	}
	if Get(db, KeyAutoUpdateHours, "") == "" {
		if autoHours < 1 {
			autoHours = 24
		}
		_ = Set(db, KeyAutoUpdateHours, strconv.Itoa(autoHours))
	}
	if Get(db, KeyLogRetentionDays, "") == "" {
		if retentionDays < 1 {
			retentionDays = 7
		}
		_ = Set(db, KeyLogRetentionDays, strconv.Itoa(retentionDays))
	}
	if Get(db, KeyErrorTipMessage, "") == "" {
		_ = Set(db, KeyErrorTipMessage, "Access is now checked for threats. For more, please contact support.")
	}
	if Get(db, KeyErrorTheme, "") == "" {
		_ = Set(db, KeyErrorTheme, "forest")
	}
	if Get(db, KeyNotifyTelegram, "") == "" {
		b, _ := json.Marshal(TelegramConfig{})
		_ = Set(db, KeyNotifyTelegram, string(b))
	}
	if Get(db, KeyNotifySMTP, "") == "" {
		b, _ := json.Marshal(SMTPConfig{Port: 587})
		_ = Set(db, KeyNotifySMTP, string(b))
	}
	if Get(db, KeyBadBehavior, "") == "" {
		b, _ := json.Marshal(BadBehaviorConfig{
			Enabled: false, Threshold: 30, WindowSec: 60, MinStatus: 400, BanDuration: "4h",
		})
		_ = Set(db, KeyBadBehavior, string(b))
	}
	if Get(db, KeyDNSBLZones, "") == "" {
		_ = Set(db, KeyDNSBLZones, "dnsbl.sorbs.net,bl.spamcop.net")
	}
	return nil
}

func LoadOps(db *gorm.DB) OpsConfig {
	tg := TelegramConfig{}
	_ = json.Unmarshal([]byte(Get(db, KeyNotifyTelegram, "{}")), &tg)
	smtp := SMTPConfig{Port: 587}
	_ = json.Unmarshal([]byte(Get(db, KeyNotifySMTP, "{}")), &smtp)
	bb := BadBehaviorConfig{Enabled: false, Threshold: 30, WindowSec: 60, MinStatus: 400, BanDuration: "4h"}
	_ = json.Unmarshal([]byte(Get(db, KeyBadBehavior, "{}")), &bb)
	if bb.Threshold < 1 {
		bb.Threshold = 30
	}
	if bb.WindowSec < 10 {
		bb.WindowSec = 60
	}
	if bb.MinStatus < 400 {
		bb.MinStatus = 400
	}
	if strings.TrimSpace(bb.BanDuration) == "" {
		bb.BanDuration = "4h"
	}
	return OpsConfig{
		AutoUpdateEnabled: GetBool(db, KeyAutoUpdateEnabled, false),
		AutoUpdateHours:   GetInt(db, KeyAutoUpdateHours, 24),
		LogRetentionDays:  GetInt(db, KeyLogRetentionDays, 7),
		ErrorTipMessage:   Get(db, KeyErrorTipMessage, ""),
		ErrorTheme:        Get(db, KeyErrorTheme, "forest"),
		Telegram:          tg,
		SMTP:              smtp,
		BadBehavior:       bb,
		DNSBLZones:        Get(db, KeyDNSBLZones, "dnsbl.sorbs.net,bl.spamcop.net"),
	}
}

func SaveOps(db *gorm.DB, ops OpsConfig, keepTelegramToken, keepSMTPPassword string) error {
	if ops.AutoUpdateHours < 1 {
		ops.AutoUpdateHours = 24
	}
	if ops.LogRetentionDays < 1 {
		ops.LogRetentionDays = 7
	}
	if ops.Telegram.BotToken == "" || ops.Telegram.BotToken == "********" {
		ops.Telegram.BotToken = keepTelegramToken
	}
	if ops.SMTP.Password == "" || ops.SMTP.Password == "********" {
		ops.SMTP.Password = keepSMTPPassword
	}
	if err := Set(db, KeyAutoUpdateEnabled, strconv.FormatBool(ops.AutoUpdateEnabled)); err != nil {
		return err
	}
	if err := Set(db, KeyAutoUpdateHours, strconv.Itoa(ops.AutoUpdateHours)); err != nil {
		return err
	}
	if err := Set(db, KeyLogRetentionDays, strconv.Itoa(ops.LogRetentionDays)); err != nil {
		return err
	}
	if err := Set(db, KeyErrorTipMessage, ops.ErrorTipMessage); err != nil {
		return err
	}
	if strings.TrimSpace(ops.ErrorTheme) == "" {
		ops.ErrorTheme = "forest"
	}
	if err := Set(db, KeyErrorTheme, ops.ErrorTheme); err != nil {
		return err
	}
	tg, _ := json.Marshal(ops.Telegram)
	if err := Set(db, KeyNotifyTelegram, string(tg)); err != nil {
		return err
	}
	sm, _ := json.Marshal(ops.SMTP)
	if err := Set(db, KeyNotifySMTP, string(sm)); err != nil {
		return err
	}
	if ops.BadBehavior.Threshold < 1 {
		ops.BadBehavior.Threshold = 30
	}
	if ops.BadBehavior.WindowSec < 10 {
		ops.BadBehavior.WindowSec = 60
	}
	if ops.BadBehavior.MinStatus < 400 {
		ops.BadBehavior.MinStatus = 400
	}
	if strings.TrimSpace(ops.BadBehavior.BanDuration) == "" {
		ops.BadBehavior.BanDuration = "4h"
	}
	bb, _ := json.Marshal(ops.BadBehavior)
	if err := Set(db, KeyBadBehavior, string(bb)); err != nil {
		return err
	}
	zones := strings.TrimSpace(ops.DNSBLZones)
	if zones == "" {
		zones = "dnsbl.sorbs.net,bl.spamcop.net"
	}
	return Set(db, KeyDNSBLZones, zones)
}

func RedactOps(ops OpsConfig) OpsConfig {
	if ops.Telegram.BotToken != "" {
		ops.Telegram.BotToken = "********"
	}
	if ops.SMTP.Password != "" {
		ops.SMTP.Password = "********"
	}
	return ops
}
