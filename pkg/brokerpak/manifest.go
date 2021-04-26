// Copyright (c) 2020-Present Pivotal Software, Inc. All Rights Reserved.
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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/ziputil"

	getter "github.com/hashicorp/go-getter"
)

const manifestName = "manifest.yml"

type Manifest struct {
	// Package metadata
	PackVersion int `yaml:"packversion"`

	// User modifiable values
	Name               string              `yaml:"name"`
	Version            string              `yaml:"version"`
	Metadata           map[string]string   `yaml:"metadata"`
	Platforms          []Platform          `yaml:"platforms"`
	TerraformResources []TerraformResource `yaml:"terraform_binaries"`
	ServiceDefinitions []string            `yaml:"service_definitions"`
	Parameters         []ManifestParameter `yaml:"parameters"`
	RequiredEnvVars    []string            `yaml:"required_env_variables"`
	EnvConfigMapping   map[string]string   `yaml:"env_config_mapping"`
}

var _ validation.Validatable = (*Manifest)(nil)

// Validate will run struct validation on the fields of this manifest.
func (m *Manifest) Validate() (errs *validation.FieldError) {
	if m.PackVersion != 1 {
		errs = errs.Also(validation.ErrInvalidValue(m.PackVersion, "packversion"))
	}

	errs = errs.Also(
		validation.ErrIfBlank(m.Name, "name"),
		validation.ErrIfBlank(m.Version, "version"),
	)

	// Platforms
	if len(m.Platforms) == 0 {
		errs = errs.Also(validation.ErrMissingField("platforms"))
	}

	for i, platform := range m.Platforms {
		errs = errs.Also(platform.Validate().ViaFieldIndex("platforms", i))
	}

	// Terraform Resources
	if len(m.TerraformResources) == 0 {
		errs = errs.Also(validation.ErrMissingField("terraform_binaries"))
	}

	for i, resource := range m.TerraformResources {
		errs = errs.Also(resource.Validate().ViaFieldIndex("terraform_binaries", i))
	}

	// Service Definitions
	if len(m.ServiceDefinitions) == 0 {
		errs = errs.Also(validation.ErrMissingField("service_definitions"))
	}

	// Params
	for i, param := range m.Parameters {
		errs = errs.Also(param.Validate().ViaFieldIndex("parameters", i))
	}

	return errs
}

// AppliesToCurrentPlatform returns true if the one of the platforms in the
// manifest match the current GOOS and GOARCH.
func (m *Manifest) AppliesToCurrentPlatform() bool {
	for _, platform := range m.Platforms {
		if platform.MatchesCurrent() {
			return true
		}
	}

	return false
}

// Pack creates a brokerpak from the manifest and definitions.
func (m *Manifest) Pack(base, dest string) error {
	// NOTE: we use "log" rather than Lager because this is used by the CLI and
	// needs to be human readable rather than JSON.
	log.Println("Packing...")

	dir, err := ioutil.TempDir("", "brokerpak")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up
	log.Println("Using temp directory:", dir)

	log.Println("Packing sources...")
	if err := m.packSources(dir); err != nil {
		return err
	}

	log.Println("Packing binaries...")
	if err := m.packBinaries(dir); err != nil {
		return err
	}

	log.Println("Packing definitions...")
	if err := m.packDefinitions(dir, base); err != nil {
		return err
	}

	log.Println("Packing terraform providers...")
	if err := m.packProviders(dir, base); err != nil {
		return err
	}

	log.Println("Creating archive:", dest)
	return ziputil.Archive(dir, dest)
}

func (m *Manifest) packSources(tmp string) error {
	for _, resource := range m.TerraformResources {
		destination := filepath.Join(tmp, "src", resource.Name+".zip")

		log.Println("\t", resource.Source, "->", destination)
		if err := fetchArchive(resource.Source, destination); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manifest) packBinaries(tmp string) error {
	for _, platform := range m.Platforms {
		platformPath := filepath.Join(tmp, "bin", platform.Os, platform.Arch)
		for _, resource := range m.TerraformResources {
			log.Println("\t", resource.Url(platform), "->", platformPath)
			if err := getter.GetAny(platformPath, resource.Url(platform)); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyProvider(sourcePath, destPath string) error {
	// open source file
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}

	// create a new one
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}

	// copy source content to destination
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to dest file failed: %s", err)
	}

	// change permission
	err = os.Chmod(destPath, 0755)
	if err != nil {
		return fmt.Errorf("Change perm on dest file failed: %s", err)
	}

	return nil
}

func visitTfFolder(tfCur string, tfTmp string, providers *map[string]bool) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ignore the .terraform folder
		if path == tfCur+"/.terraform/" {
			return nil
		}

		// split to kept path
		s := strings.Split(path, tfCur+"/.terraform/")

		// create folder if not exists in tmp folder
		if info.IsDir() == true {
			_, err := os.Stat(filepath.Join(tfTmp, s[1]))
			if os.IsNotExist(err) {
				err := os.MkdirAll(filepath.Join(tfTmp, s[1]), 0755)
				if err != nil {
					return fmt.Errorf("unable to create folder in tmp %v", err)
				}
			}
		}

		// file ?
		if info.IsDir() == false {
			tfBin := filepath.Base(path)

			// just take in account file starting with terraform-provider
			if strings.HasPrefix(tfBin, "terraform-provider-") == true {
				exists := (*providers)[tfBin]
				if exists == false {
					log.Printf("\t%s", tfBin)
					(*providers)[tfBin] = true

					// move the provider binary in the temp folder
					err := CopyProvider(path, filepath.Join(tfTmp, s[1]))
					if err != nil {
						return fmt.Errorf("unable to copy provider in tmp %v", err)
					}
				}
			}
		}

		return nil
	}
}

func (m *Manifest) packProviders(tmp, base string) error {
	err := os.MkdirAll(filepath.Join(tmp, "/terraform.d/providers"), 0755)
	if err != nil {
		return fmt.Errorf("unable to create terraform providers folder in tmp %v", err)
	}

	// search terraform folders in all services definitions
	// duplicated folders are removed
	tfDirs := make(map[string]bool)
	for _, sd := range m.ServiceDefinitions {
		defn := &tf.TfServiceDefinitionV1{}
		if err := stream.Copy(stream.FromFile(base, sd), stream.ToYaml(defn)); err != nil {
			return fmt.Errorf("couldn't parse %s: %v", sd, err)
		}

		if defn.ProvisionSettings.TemplateRef != "" {
			tf_base := filepath.Dir(defn.ProvisionSettings.TemplateRef)
			tfDirs[tf_base] = true
		}

		for _, ref := range defn.ProvisionSettings.TemplateRefs {
			if ref != "" {
				tfDirs[filepath.Dir(ref)] = true
			}
		}

		if defn.BindSettings.TemplateRef != "" {
			tf_base := filepath.Dir(defn.BindSettings.TemplateRef)
			tfDirs[tf_base] = true
		}

		for _, ref := range defn.BindSettings.TemplateRefs {
			if ref != "" {
				tfDirs[filepath.Dir(ref)] = true
			}
		}

	}

	// execute terraform init on each folders
	// duplicated providers are removed
	providers := make(map[string]bool)
	for _, platform := range m.Platforms {
		for tfDir, _ := range tfDirs {
			platformPath := filepath.Join(tmp, "bin", platform.Os, platform.Arch)

			// terraform path
			tfDirPath := filepath.Join(base, tfDir)

			// clean up terraform folder
			os.RemoveAll(tfDirPath + "/.terraform/")

			// execute terraform
			cmd := exec.Command(platformPath+"/terraform", "init")
			cmd.Dir = tfDirPath

			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				return fmt.Errorf(fmt.Sprint(err) + ": " + stderr.String())
			}

			// search plugins in the hidden terraform directory
			err = filepath.Walk(tfDirPath+"/.terraform/", visitTfFolder(tfDirPath, tmp+"/terraform.d", &providers))
			if err != nil {
				return fmt.Errorf("walk on terraform failed: %v", err)
			}

			// clean up terraform folder
			os.RemoveAll(tfDirPath + "/.terraform/")
		}
	}

	return nil
}

func clearRefs(sd *tf.TfServiceDefinitionV1Action) {
	sd.TemplateRef = ""
	sd.TemplateRefs = make(map[string]string)
}

func (m *Manifest) packDefinitions(tmp, base string) error {
	// users can place definitions in any directory structure they like, even
	// above the current directory so we standardize their location and names
	// for the zip to avoid collisions
	//
	// provision and bind templates are loaded from any template ref and packed inline
	manifestCopy := *m

	var servicePaths []string
	for i, sd := range m.ServiceDefinitions {

		defn := &tf.TfServiceDefinitionV1{}
		if err := stream.Copy(stream.FromFile(base, sd), stream.ToYaml(defn)); err != nil {
			return fmt.Errorf("couldn't parse %s: %v", sd, err)
		}

		if err := defn.ProvisionSettings.LoadTemplate(base); err != nil {
			return fmt.Errorf("couldn't load provision template %s: %v", defn.ProvisionSettings.TemplateRef, err)
		}

		if err := defn.BindSettings.LoadTemplate(base); err != nil {
			return fmt.Errorf("couldn't load bind template %s: %v", defn.BindSettings.TemplateRef, err)
		}

		clearRefs(&defn.ProvisionSettings)
		clearRefs(&defn.BindSettings)

		packedName := fmt.Sprintf("service%d-%s.yml", i, defn.Name)
		log.Printf("\t%s/%s -> %s/definitions/%s\n", base, sd, tmp, packedName)
		if err := stream.Copy(stream.FromYaml(defn), stream.ToFile(tmp, "definitions", packedName)); err != nil {
			return err
		}

		servicePaths = append(servicePaths, "definitions/"+packedName)
	}

	manifestCopy.ServiceDefinitions = servicePaths

	return stream.Copy(stream.FromYaml(manifestCopy), stream.ToFile(tmp, manifestName))
}

// ManifestParameter holds environment variables that will be looked up and
// passed to the executed Terraform instance.
type ManifestParameter struct {
	// NOTE: Future fields should take inspiration from the CNAB spec because they
	// solve a similar problem. https://github.com/deislabs/cnab-spec
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

var _ validation.Validatable = (*ManifestParameter)(nil)

// Validate implements validation.Validatable.
func (param *ManifestParameter) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(param.Name, "name"),
		validation.ErrIfBlank(param.Description, "description"),
	)
}

// NewExampleManifest creates a new manifest with sample values for the service broker suitable for giving a user a template to manually edit.
func NewExampleManifest() Manifest {
	return Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []Platform{
			{Os: "linux", Arch: "386"},
			{Os: "linux", Arch: "amd64"},
		},
		TerraformResources: []TerraformResource{
			{
				Name:    "terraform",
				Version: "0.11.9",
				Source:  "https://github.com/hashicorp/terraform/archive/v0.11.9.zip",
			},
			{
				Name:    "terraform-provider-google-beta",
				Version: "1.19.0",
				Source:  "https://github.com/terraform-providers/terraform-provider-google/archive/v1.19.0.zip",
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []ManifestParameter{
			{Name: "MY_ENVIRONMENT_VARIABLE", Description: "Set this to whatever you like."},
		},
	}
}
