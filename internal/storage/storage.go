// Package storage implements a Database Access Object (DAO)
package storage

import (
	"os"

	"gorm.io/gorm"
)

type Storage struct {
	db          *gorm.DB
	encryptor   Encryptor
	lockFileDir string
}

func New(db *gorm.DB, encryptor Encryptor) *Storage {
	// the VM based HA deployment requires a drain mechanism. LockFiles are a simple solution.
	// but not every environment will opt for using VM based deployments. So detect if the lockfile
	// director is present.

	dirDefault := os.Getenv("CSB_LOCKFILE_DIR")
	if _, err := os.Stat(dirDefault); err != nil {
		dirDefault, _ = os.MkdirTemp("/tmp/", "lockfiles")
	}
	return &Storage{
		db:          db,
		encryptor:   encryptor,
		lockFileDir: dirDefault,
	}
}

type JSONObject map[string]any
