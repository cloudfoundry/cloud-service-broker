// Package fetcher has logic for fetching a file from a source
// (which may be a file or URL) and saving it to a destination.
package fetcher

import (
	"github.com/hashicorp/go-getter"
)

func FetchArchive(src, dest string) error {
	return newFileGetterClient(src, dest).Get()
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
