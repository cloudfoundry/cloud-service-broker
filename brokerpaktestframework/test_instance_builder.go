package brokerpaktestframework

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/yaml.v3"
)

func BuildTestInstance(brokerPackDir string, provider TerraformMock, logger io.Writer) (*TestInstance, error) {
	csbBuild, err := gexec.Build("github.com/cloudfoundry/cloud-service-broker")
	if err != nil {
		return nil, err
	}

	workingDir, err := createWorkspace(brokerPackDir, provider.Binary)
	if err != nil {
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

func createWorkspace(brokerPackDir string, build string) (string, error) {
	workingDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		return "", err
	}
	err = linkBrokerpackFiles(brokerPackDir, workingDir)
	if err != nil {
		return "", err
	}

	return workingDir, templateManifest(brokerPackDir, build, workingDir)
}

func linkBrokerpackFiles(brokerPackDir string, workingDir string) error {
	yamlFiles, err := filepath.Glob(brokerPackDir + "/*.yml")
	if err != nil {
		return err
	}
	for _, file := range yamlFiles {
		err = os.Link(file, path.Join(workingDir, filepath.Base(file)))
		if err != nil {
			return err
		}
	}
	err = os.Symlink(path.Join(brokerPackDir, "terraform"), path.Join(workingDir, "terraform"))
	if err != nil {
		return err
	}
	err = os.Remove(path.Join(workingDir, "manifest.yml"))
	if err != nil {
		return err
	}
	return nil
}

func templateManifest(brokerPackDir string, build string, workingDir string) error {
	contents, err := ioutil.ReadFile(path.Join(brokerPackDir, "manifest.yml"))
	if err != nil {
		return err
	}
	parsedManifest, err := manifest.Parse(contents)
	if err != nil {
		return err
	}
	setArch(parsedManifest)
	replaceTerraformBinaries(parsedManifest, build)

	outputFile, err := os.Create(path.Join(workingDir, "manifest.yml"))
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(outputFile)
	err = encoder.Encode(parsedManifest)
	if err != nil {
		return err
	}
	return outputFile.Close()
}

func replaceTerraformBinaries(parsedManifest *manifest.Manifest, terraformBuild string) error {
	parsedManifest.TerraformResources = []manifest.TerraformResource{
		{
			Name:        "terraform",
			Version:     "1.1.4",
			URLTemplate: terraformBuild,
		},
	}
	return nil
}

func setArch(parsedManifest *manifest.Manifest) {
	parsedManifest.Platforms = []platform.Platform{
		{
			Os:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}
}
