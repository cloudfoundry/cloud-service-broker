package zippy

import "archive/zip"

func Open(path string) (ZipReader, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return ZipReader{}, err
	}

	return ZipReader{reader: reader}, nil
}
