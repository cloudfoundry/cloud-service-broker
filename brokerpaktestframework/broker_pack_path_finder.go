package brokerpaktestframework

import (
	"path/filepath"
	"runtime"
)

func PathToBrokerPack(skips ...int) string {
	skip := 1
	switch len(skips) {
	case 0:
	case 1:
		skip += skips[0]
	default:
		panic("too many skips")
	}
	_, file, _, _ := runtime.Caller(skip)

	return filepath.Dir(filepath.Dir(file))
}
