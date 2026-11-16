package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/overthecloud/the-gate/internal/settings"
)

func SendTelegram(cfg settings.TelegramConfig, text string) error {
	if !cfg.Enabled {
		return fmt.Errorf("telegram notifications disabled")
	}
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return fmt.Errorf("telegram bot_token and chat_id required")
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.BotToken)
	body, _ := json.Marshal(map[string]any{
		"chat_id":    cfg.ChatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram HTTP %d", resp.StatusCode)
	}
	return nil
}

func SendSMTP(cfg settings.SMTPConfig, subject, body string) error {
	if !cfg.Enabled {
		return fmt.Errorf("smtp notifications disabled")
	}
	if cfg.Host == "" || cfg.From == "" || cfg.To == "" {
		return fmt.Errorf("smtp host, from, and to required")
	}
	port := cfg.Port
	if port == 0 {
		port = 587
	}
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(port))
	msg := strings.Builder{}
	msg.WriteString("From: " + cfg.From + "\r\n")
	msg.WriteString("To: " + cfg.To + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	msg.WriteString(body)

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	return smtp.SendMail(addr, auth, cfg.From, splitAddrs(cfg.To), []byte(msg.String()))
}

func NotifyAll(ops settings.OpsConfig, subject, body string) []string {
	var errs []string
	if ops.Telegram.Enabled {
		if err := SendTelegram(ops.Telegram, "<b>"+subject+"</b>\n"+body); err != nil {
			errs = append(errs, "telegram: "+err.Error())
		}
	}
	if ops.SMTP.Enabled {
		if err := SendSMTP(ops.SMTP, subject, body); err != nil {
			errs = append(errs, "smtp: "+err.Error())
		}
	}
	return errs
}

func splitAddrs(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ';' })
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
