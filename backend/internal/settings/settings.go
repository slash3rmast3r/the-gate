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

type OpsConfig struct {
	AutoUpdateEnabled bool   `json:"auto_update_enabled"`
	AutoUpdateHours   int    `json:"auto_update_hours"`
	LogRetentionDays  int    `json:"log_retention_days"`
	ErrorTipMessage   string         `json:"error_tip_message"`
	ErrorTheme        string         `json:"error_theme"`
	Telegram          TelegramConfig `json:"telegram"`
	SMTP              SMTPConfig     `json:"smtp"`
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
	return nil
}

func LoadOps(db *gorm.DB) OpsConfig {
	tg := TelegramConfig{}
	_ = json.Unmarshal([]byte(Get(db, KeyNotifyTelegram, "{}")), &tg)
	smtp := SMTPConfig{Port: 587}
	_ = json.Unmarshal([]byte(Get(db, KeyNotifySMTP, "{}")), &smtp)
	return OpsConfig{
		AutoUpdateEnabled: GetBool(db, KeyAutoUpdateEnabled, false),
		AutoUpdateHours:   GetInt(db, KeyAutoUpdateHours, 24),
		LogRetentionDays:  GetInt(db, KeyLogRetentionDays, 7),
		ErrorTipMessage:   Get(db, KeyErrorTipMessage, ""),
		ErrorTheme:        Get(db, KeyErrorTheme, "forest"),
		Telegram:          tg,
		SMTP:              smtp,
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
	return Set(db, KeyNotifySMTP, string(sm))
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
