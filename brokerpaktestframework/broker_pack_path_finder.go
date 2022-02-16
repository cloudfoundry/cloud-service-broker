package brokerpaktestframework

import (
	"path/filepath"
	"runtime"
)

func PathToBrokerPack() string {
	_, file, _, _ := runtime.Caller(1)

	return filepath.Dir(filepath.Dir(file))
}