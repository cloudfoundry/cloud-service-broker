// Package zippy is a basic API to zip and unzip files that uses archive/zip
// Name inspired by: https://en.wikipedia.org/wiki/Zippy_(Rainbow)
package zippy

import "archive/zip"

type ZipReader struct {
	reader *zip.ReadCloser
}
