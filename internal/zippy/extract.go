package zippy

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
)

func (z ZipReader) Extract(zipDirectory, targetDirectory string) error {
	for _, fd := range z.reader.File {
		if fd.UncompressedSize64 == 0 { // skip directories
			continue
		}

		if !strings.HasPrefix(fd.Name, zipDirectory) {
			continue
		}

		if containsDotDot(fd.Name) {
			return fmt.Errorf("potential zip slip extracting %q", fd.Name)
		}

		src := stream.FromReadCloserError(fd.Open())

		newName := strings.TrimPrefix(fd.Name, zipDirectory)
		destPath := filepath.Join(targetDirectory, filepath.FromSlash(newName))
		dest := stream.ToModeFile(fd.Mode(), destPath)

		if err := stream.Copy(src, dest); err != nil {
			return fmt.Errorf("couldn't extract file %q: %v", fd.Name, err)
		}
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
