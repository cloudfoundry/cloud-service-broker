package zippy

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
)

// ExtractDirectory extracts a path within the zip file to a target directory
func (z ZipReader) ExtractDirectory(zipDirectory, targetDirectory string) error {
	files, err := z.listExtractableFiles()
	if err != nil {
		return err
	}

	for _, fd := range files {
		if !strings.HasPrefix(fd.Name, zipDirectory) {
			continue
		}

		destPath := filepath.Join(targetDirectory, strings.TrimPrefix(fd.Name, zipDirectory))
		if err := z.extractFile(fd, destPath); err != nil {
			return err
		}
	}

	return nil
}

func (z ZipReader) ExtractFile(filePath, targetDirectory string) error {
	files, err := z.listExtractableFiles()
	if err != nil {
		return err
	}

	for _, fd := range files {
		if fd.Name != filePath {
			continue
		}

		destPath := filepath.Join(targetDirectory, filepath.Base(filePath))
		return z.extractFile(fd, destPath)
	}

	return fmt.Errorf("file %q does not exist in the zip", filePath)
}

func (z ZipReader) listExtractableFiles() (result []*zip.File, err error) {
	for _, fd := range z.List() {
		if fd.UncompressedSize64 == 0 { // skip directories
			continue
		}

		if containsDotDot(fd.Name) {
			return nil, fmt.Errorf("potential zip slip extracting %q", fd.Name)
		}

		result = append(result, fd)
	}

	return result, nil
}

func (z ZipReader) extractFile(fd *zip.File, destPath string) error {
	src := stream.FromReadCloserError(fd.Open())
	dest := stream.ToModeFile(fd.Mode(), destPath)

	if err := stream.Copy(src, dest); err != nil {
		return fmt.Errorf("couldn't extract file %q: %v", fd.Name, err)
	}

	return nil
}

// containsDotDot checks if the filepath value v contains a ".." entry.
// This will check filepath components by splitting along / or \. This
// function is copied directly from the Go net/http implementation.
func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }
