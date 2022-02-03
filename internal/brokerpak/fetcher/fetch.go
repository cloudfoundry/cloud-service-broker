// Package fetcher has logic for fetching a file from a source
// (which may be a file or URL) and saving it to a destination.
package fetcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
)

func FetchArchive(src, dest string) error {
	return newFileGetterClient(src, dest).Get()
}

func FetchBrokerpak(src, dest string) error {
	execWd := filepath.Dir(os.Args[0])
	execDir, err := filepath.Abs(execWd)
	if err != nil {
		return fmt.Errorf("couldn't turn dir %q into abs path: %v", execWd, err)
	}

	client := newFileGetterClient(src, dest)
	client.Pwd = execDir

	return client.Get()
}

func newFileGetterClient(src, dest string) *getter.Client {
	return &getter.Client{
		Src: src,
		Dst: dest,

		Mode:          getter.ClientModeFile,
		Getters:       defaultGetters(),
		Decompressors: map[string]getter.Decompressor{},
	}
}

func defaultGetters() map[string]getter.Getter {
	getters := map[string]getter.Getter{}
	for k, g := range getter.Getters {
		getters[k] = g
	}

	return getters
}
