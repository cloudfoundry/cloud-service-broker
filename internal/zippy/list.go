package zippy

import "archive/zip"

func (z ZipReader) List() (result []*zip.File) {
	return z.reader.File
}
