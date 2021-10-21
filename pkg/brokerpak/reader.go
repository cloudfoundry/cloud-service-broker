// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package brokerpak

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
)

// BrokerPakReader reads bundled together Terraform and service definitions.
type BrokerPakReader struct {
	contents zippy.ZipReader
}

func (pak *BrokerPakReader) readYaml(name string, v interface{}) error {
	fd := pak.contents.Find(name)
	if fd == nil {
		return fmt.Errorf("couldn't find the file with the name %q", name)
	}

	return stream.Copy(stream.FromReadCloserError(fd.Open()), stream.ToYaml(v))
}

// Manifest fetches the manifest out of the package.
func (pak *BrokerPakReader) Manifest() (*Manifest, error) {
	manifest := &Manifest{}

	if err := pak.readYaml(manifestName, manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}

// Services gets the list of services included in the pack.
func (pak *BrokerPakReader) Services() ([]tf.TfServiceDefinitionV1, error) {
	manifest, err := pak.Manifest()
	if err != nil {
		return nil, err
	}

	var services []tf.TfServiceDefinitionV1
	for _, serviceDefinition := range manifest.ServiceDefinitions {
		tmp := tf.TfServiceDefinitionV1{}
		if err := pak.readYaml(serviceDefinition, &tmp); err != nil {
			return nil, err
		}

		tmp.RequiredEnvVars = manifest.RequiredEnvVars
		services = append(services, tmp)
	}

	return services, nil
}

// Validate checks the manifest and service definitions for syntactic and
// limited semantic errors.
func (pak *BrokerPakReader) Validate() error {
	manifest, err := pak.Manifest()
	if err != nil {
		return fmt.Errorf("couldn't open brokerpak manifest: %v", err)
	}

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("couldn't validate brokerpak manifest: %v", err)
	}

	services, err := pak.Services()
	if err != nil {
		return fmt.Errorf("couldn't list services: %v", err)
	}

	for _, svc := range services {
		if err := svc.Validate(); err != nil {
			return fmt.Errorf("service %q failed validation: %v", svc.Name, err)
		}
	}

	return nil
}

// Close closes the underlying reader for the BrokerPakReader.
func (pak *BrokerPakReader) Close() error {
	pak.contents.Close()
	return nil
}

// ExtractPlatformBins extracts the binaries for the current platform to the
// given destination.
func (pak *BrokerPakReader) ExtractPlatformBins(destination string) error {
	mf, err := pak.Manifest()
	if err != nil {
		return err
	}

	if !mf.AppliesToCurrentPlatform() {
		return fmt.Errorf("the package %q doesn't contain binaries compatible with the current platform %q", mf.Name, CurrentPlatform().String())
	}

	terraformVersion, err := getTerraformVersion(mf)
	if err != nil {
		return err
	}

	switch {
	case terraformVersion.LessThan(version.Must(version.NewVersion("0.12.0"))):
		return errors.New("terraform version too low")
	case terraformVersion.LessThan(version.Must(version.NewVersion("0.13.0"))):
		return pak.extractPlatformBins12(destination, mf)
	case terraformVersion.LessThan(version.Must(version.NewVersion("0.14.0"))):
		return pak.extractPlatformBins13(destination, mf)
	default:
		return errors.New("terraform version too high")
	}
}

func getTerraformVersion(mf *Manifest) (*version.Version, error) {
	for _, r := range mf.TerraformResources {
		if r.Name == "terraform" {
			return version.NewVersion(r.Version)
		}
	}
	return nil, errors.New("terraform not found in manifest")
}

func (pak *BrokerPakReader) extractPlatformBins12(destination string, mf *Manifest) error {
	plat := CurrentPlatform()
	bindir := path.Join("bin", plat.Os, plat.Arch)
	return pak.contents.ExtractDirectory(bindir, destination)
}

func (pak *BrokerPakReader) extractPlatformBins13(destination string, mf *Manifest) error {
	plat := CurrentPlatform()
	for _, r := range mf.TerraformResources {
		if strings.HasPrefix(r.Name, "terraform-provider-") {
			filePath, err := pak.findFileInZip(r.Name, r.Version)
			if err != nil {
				return err
			}
			if err := pak.contents.ExtractFile(filePath, providerInstallPath(destination, r.Name, r.Version)); err != nil {
				return fmt.Errorf("error extracting terraform-provider file: %w", err)
			}
		} else {
			file := filepath.Join("bin", plat.Os, plat.Arch, r.Name)
			if err := pak.contents.ExtractFile(file, destination); err != nil {
				return fmt.Errorf("error extracting %q to %q: %w", file, destination, err)
			}
		}
	}

	return nil
}

func (pak *BrokerPakReader) findFileInZip(name, version string) (string, error) {
	plat := CurrentPlatform()
	prefix := path.Join("bin", plat.Os, plat.Arch, fmt.Sprintf("%s_v%s", name, version))
	var found []string

	for _, f := range pak.contents.List() {
		if strings.HasPrefix(f.Name, prefix) {
			found = append(found, f.Name)
		}
	}

	switch len(found) {
	case 1:
		return found[0], nil
	case 0:
		return "", fmt.Errorf("file with prefix %q not found in zip", prefix)
	default:
		return "", fmt.Errorf("multiple files found with prefix %q: %s", prefix, strings.Join(found, ", "))
	}
}

func providerInstallPath(destination, name, version string) string {
	suffix := name[19:] // chop off 'terraform-provider-'
	plat := CurrentPlatform()
	target := fmt.Sprintf("%s_%s", plat.Os, plat.Arch)
	return filepath.Join(destination, "registry.terraform.io", "hashicorp", suffix, version, target)
}

// OpenBrokerPak opens the file at the given path as a BrokerPakReader.
func OpenBrokerPak(pakPath string) (*BrokerPakReader, error) {
	rc, err := zippy.Open(pakPath)
	if err != nil {
		return nil, err
	}
	return &BrokerPakReader{contents: rc}, nil
}

// DownloadAndOpenBrokerpak downloads a (potentially remote) brokerpak to
// the local filesystem and opens it.
func DownloadAndOpenBrokerpak(pakUri string) (*BrokerPakReader, error) {
	// create a temp directory to hold the pak
	pakDir, err := os.MkdirTemp("", "brokerpak-staging")
	if err != nil {
		return nil, fmt.Errorf("couldn't create brokerpak staging area for %q: %v", pakUri, err)
	}

	// Download the brokerpak
	localLocation := filepath.Join(pakDir, "pack.brokerpak")
	if err := fetchBrokerpak(pakUri, localLocation); err != nil {
		return nil, fmt.Errorf("couldn't download brokerpak %q: %v", pakUri, err)
	}

	return OpenBrokerPak(localLocation)
}
