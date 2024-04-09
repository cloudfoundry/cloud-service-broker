package zippy

import (
	"archive/zip"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/cloud-service-broker/v2/utils/stream"
)

func Archive(sourceDirectory, destinationZip string, compress bool) error {
	fd, err := os.Create(destinationZip)
	if err != nil {
		return fmt.Errorf("couldn't create archive %q: %v", destinationZip, err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(fd)

	w := zip.NewWriter(fd)
	defer func(zw *zip.Writer) {
		_ = zw.Close()
	}(w)

	sourceDirectory = path.Clean(sourceDirectory)
	return filepath.Walk(sourceDirectory, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == sourceDirectory {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if compress {
			header.Method = zip.Deflate
		}

		if info.IsDir() {
			header.Name = fmt.Sprintf("%s%c", clean(strings.TrimPrefix(path, sourceDirectory)), os.PathSeparator)
			if _, err = w.CreateHeader(header); err != nil {
				return err
			}
		} else {
			header.Name = clean(strings.TrimPrefix(path, sourceDirectory))
			fd, err := w.CreateHeader(header)
			if err != nil {
				return err
			}

			if err := stream.Copy(stream.FromFile(path), stream.ToWriter(fd)); err != nil {
				return err
			}
		}

		return nil
	})
}

func clean(path string) string {
	slashStrip := strings.TrimPrefix(path, "/")
	dotStrip := strings.TrimPrefix(slashStrip, "./")
	return dotStrip
}
