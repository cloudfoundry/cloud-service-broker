package packer

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

func cachedFetchFile(getter func(source string, destination string) error, source, destination, cachePath string) error {
	cacheKey := buildCacheKey(cachePath, source)

	switch {
	case cachePath == "":
		log.Println("\t", source, "->", destination, "(no cache)")
		return getter(source, destination)
	case exists(source):
		log.Println("\t", source, "->", destination, "(local file)")
		return copyLocalFile(source, destination)
	case exists(cacheKey):
		log.Println("\t", source, "->", destination, "(from cache)")
		return cp.Copy(cacheKey, destination)
	default:
		log.Println("\t", source, "->", destination, "(stored to cache)")
		return getAndCache(getter, source, destination, cacheKey)
	}
}

func buildCacheKey(cachePath string, source string) string {
	return path.Join(cachePath, fmt.Sprintf("%x", md5.Sum([]byte(source))))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func copyLocalFile(source, destination string) error {
	return cp.Copy(source, filepath.Join(destination, filepath.Base(source)))
}

// getAndCache is a "tee" function that copies the source file to both the destination and the cache.
// It downloads to a temporary directory first and then copies the data twice. This reduces
// the chance of partially downloaded files polluting the cache after failure, or concurrency issues.
func getAndCache(getter func(source string, destination string) error, source, destination, cacheKey string) error {
	var tmpdir string
	defer func() {
		_ = os.RemoveAll(tmpdir)
	}()

	for _, step := range []func() error{
		func() (err error) {
			tmpdir, err = os.MkdirTemp("", "")
			return
		},
		func() error {
			return getter(source, tmpdir)
		},
		func() error {
			return cp.Copy(tmpdir, cacheKey)
		},
		func() error {
			return cp.Copy(tmpdir, destination)
		},
	} {
		if err := step(); err != nil {
			return err
		}
	}

	return nil
}
