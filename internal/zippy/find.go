package zippy

import (
	"archive/zip"
)

func (z ZipReader) Find(path string) *zip.File {
	for _, f := range z.reader.File {
		if f.Name == path {
			return f
		}
	}

	return nil
}
