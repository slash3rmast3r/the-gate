package retention

import (
	"os"
	"path/filepath"
	"time"

	"github.com/overthecloud/the-gate/internal/models"
	"gorm.io/gorm"
)

// PruneLogs removes rotated log files older than days under logDir.
func PruneLogs(logDir string, days int) (int, error) {
	if days < 1 || logDir == "" {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	removed := 0
	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		name := info.Name()
		// Keep active access.log / traefik.log; prune rotated .gz / .1 / dated files
		if name == "access.log" || name == "traefik.log" {
			return nil
		}
		if info.ModTime().Before(cutoff) {
			if os.Remove(path) == nil {
				removed++
			}
		}
		return nil
	})
	return removed, err
}

func PruneAudit(db *gorm.DB, days int) (int64, error) {
	if days < 1 {
		return 0, nil
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	res := db.Where("created_at < ?", cutoff).Delete(&models.AuditEntry{})
	return res.RowsAffected, res.Error
}
