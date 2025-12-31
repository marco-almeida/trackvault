package core

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

func DefaultBackupDir() string {
	return filepath.Join(xdg.DataHome, "trackvault", "backups")
}
