package brokerpaktestframework

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/hashicorp/go-version"
	"github.com/onsi/gomega/gexec"
	cp "github.com/otiai10/copy"
)

const (
	terraformVersion = "1.1.4"
)

var (
	brokerpakDefaultFoldersToCopy = []string{"terraform"}
)

func BuildTestInstance(brokerPackDir string, provider TerraformMock, logger io.Writer, brokerpakExtraFoldersToCopy ...string) (*TestInstance, error) {
	csbBuild, err := gexec.Build("github.com/cloudfoundry/cloud-service-broker")
	if err != nil {
		return nil, err
	}

	workingDir, err := os.MkdirTemp("", "prefix")
	if err != nil {
		return nil, fmt.Errorf("error creating temporal working directory %w", err)
	}

	if err := copyBrokerpakYMLFiles(brokerPackDir, workingDir); err != nil {
		return nil, err
	}

	folders := append(brokerpakDefaultFoldersToCopy, brokerpakExtraFoldersToCopy...)
	if err := copyBrokerpakFolders(brokerPackDir, workingDir, folders); err != nil {
		return nil, err
	}

	if err := writeManifest(brokerPackDir, provider.Binary, workingDir); err != nil {
		return nil, err
	}

	command := exec.Command(csbBuild, "pak", "build")
	command.Dir = workingDir
	session, err := gexec.Start(command, logger, logger)

	if err != nil {
		return nil, err
	}

	session.Wait(5 * time.Minute)

	return &TestInstance{brokerBuild: csbBuild, workspace: workingDir, username: "u", password: "p", port: "8080"}, nil
}

func copyBrokerpakYMLFiles(brokerPackDir string, workingDir string) error {
	yamlFiles, err := filepath.Glob(brokerPackDir + "/*.yml")
	if err != nil {
		return err
	}

	for _, canonicalFilePath := range yamlFiles {
		filename := filepath.Base(canonicalFilePath)
		if filename == "manifest.yml" {
			continue
		}

		if err = cp.Copy(canonicalFilePath, path.Join(workingDir, filename)); err != nil {
			return err
		}
	}

	return nil
}

func copyBrokerpakFolders(brokerPackDir string, workingDir string, folders []string) error {
	for _, directory := range folders {
		src := path.Join(brokerPackDir, directory)
		dst := path.Join(workingDir, directory)
		if err := cp.Copy(src, dst); err != nil {
			return fmt.Errorf("error in folder copy operation - src: %s - dst %s", src, dst)
		}
	}
	return nil
}

func writeManifest(brokerPackDir string, build string, workingDir string) (err error) {
	contents, err := os.ReadFile(path.Join(brokerPackDir, "manifest.yml"))
	if err != nil {
		return
	}
	parsedManifest, err := manifest.Parse(contents)
	if err != nil {
		return
	}

	parsedManifest.Platforms = []platform.Platform{{Os: runtime.GOOS, Arch: runtime.GOARCH}}
	versionOrPanic := version.Must(version.NewVersion(terraformVersion))
	parsedManifest.TerraformVersions = []manifest.TerraformVersion{{Version: versionOrPanic, URLTemplate: build}}
	parsedManifest.TerraformProviders = nil
	outputFile, err := os.Create(path.Join(workingDir, "manifest.yml"))
	if err != nil {
		return
	}
	defer func() {
		closeErr := outputFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	serializedManifest, err := parsedManifest.Serialize()
	if err != nil {
		return
	}

	_, err = outputFile.Write(serializedManifest)
	if err != nil {
		return
	}

	return
}
