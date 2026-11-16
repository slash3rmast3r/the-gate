package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/overthecloud/the-gate/internal/models"
	"github.com/overthecloud/the-gate/internal/settings"
	"gorm.io/gorm"
)

func WriteZip(w io.Writer, db *gorm.DB, traefikDynamicDir string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	stamp := time.Now().UTC().Format("20060102-150405")
	meta := map[string]any{
		"product":   "The Gate",
		"created":   time.Now().UTC().Format(time.RFC3339),
		"note":      "Control-plane backup (apps, settings, IP groups, audit, traefik dynamic). Full DR: use scripts/backup.sh",
		"stamp":     stamp,
	}
	if err := writeJSON(zw, "meta.json", meta); err != nil {
		return err
	}

	apps, err := listApps(db)
	if err != nil {
		return err
	}
	if err := writeJSON(zw, "applications.json", apps); err != nil {
		return err
	}

	ops := settings.LoadOps(db)
	if err := writeJSON(zw, "settings.json", settings.RedactOps(ops)); err != nil {
		return err
	}

	var groups []models.IPGroup
	_ = db.Order("kind asc, name asc").Find(&groups).Error
	if err := writeJSON(zw, "ip_groups.json", groups); err != nil {
		return err
	}

	var audit []models.AuditEntry
	_ = db.Order("created_at desc").Limit(500).Find(&audit).Error
	if err := writeJSON(zw, "audit.json", audit); err != nil {
		return err
	}

	if traefikDynamicDir != "" {
		_ = filepath.Walk(traefikDynamicDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(traefikDynamicDir, path)
			return writeFile(zw, filepath.ToSlash(filepath.Join("traefik-dynamic", rel)), path)
		})
	}
	return nil
}

func listApps(db *gorm.DB) ([]models.Application, error) {
	var apps []models.Application
	err := db.Order("name asc").Find(&apps).Error
	out := make([]models.Application, len(apps))
	for i, a := range apps {
		out[i] = a.Redacted()
	}
	return out, err
}

func writeJSON(zw *zip.Writer, name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func writeFile(zw *zip.Writer, name, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func Filename() string {
	return fmt.Sprintf("the-gate-backup-%s.zip", time.Now().UTC().Format("20060102-150405"))
}
