// Package storage implements a Database Access Object (DAO)
package storage

import (
	"os"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const lockfileDir = "lockfiledir"

func init() {
	viper.BindEnv(lockfileDir, "CSB_LOCKFILE_DIR")
}

type Storage struct {
	db          *gorm.DB
	encryptor   Encryptor
	lockFileDir string
}

func New(db *gorm.DB, encryptor Encryptor) *Storage {
	// the VM based HA deployment requires a drain mechanism. LockFiles are a simple solution.
	// but not every environment will opt for using VM based deployments. So detect if the lockfile
	// director is present.

	dirDefault := viper.GetString(lockfileDir)
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
