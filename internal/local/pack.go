package local

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/pkg/brokerpak"
)

func pack(cachePath string) (string, func()) {
	pakPath, err := brokerpak.Pack("", cachePath, false, false, platform.CurrentPlatform())
	if err != nil {
		log.Fatalf("error while packing: %v", err)
	}

	if err := brokerpak.Validate(pakPath); err != nil {
		log.Fatalf("created: %v, but it failed validity checking: %v\n", pakPath, err)
	} else {
		fmt.Printf("created: %v\n", pakPath)
	}

	pakPath, err = filepath.Abs(pakPath)
	if err != nil {
		log.Fatal(err)
	}

	pakDir, err := os.MkdirTemp("", "csb-*")
	if err != nil {
		log.Fatal(err)
	}

	if err := os.Symlink(pakPath, filepath.Join(pakDir, filepath.Base(pakPath))); err != nil {
		log.Fatal(err)
	}

	return pakDir, func() {
		os.RemoveAll(pakDir)
		os.RemoveAll(pakPath)
	}
}
