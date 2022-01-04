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

package brokerpak_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/brokerpak"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("manifest", func() {
	Describe("validate", func() {
		It("should not error when manifest is valid", func() {
			exampleManifest := brokerpak.NewExampleManifest()

			err := exampleManifest.Validate()

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("unmarshalling", func() {
		ymlManifest := `packversion: 1
name: my-services-pack
version: 0.1.0
metadata:
  author: VMware
platforms:
- os: linux
  arch: amd64
terraform_binaries:
- name: terraform
  version: 0.12.23
  source: https://github.com/hashicorp/terraform/archive/v0.12.23.zip  
required_env_variables:
- ARM_SUBSCRIPTION_ID
service_definitions:
- example-service-definition.yml
env_config_mapping:
  ARM_SUBSCRIPTION_ID: azure.subscription_id`

		expectedManifest := brokerpak.Manifest{
			PackVersion: 1,
			Name:        "my-services-pack",
			Version:     "0.1.0",
			Metadata: map[string]string{
				"author": "VMware",
			},
			Platforms: []brokerpak.Platform{
				{Os: "linux", Arch: "amd64"},
			},
			TerraformResources: []brokerpak.TerraformResource{
				{
					Name:    "terraform",
					Version: "0.12.23",
					Source:  "https://github.com/hashicorp/terraform/archive/v0.12.23.zip",
				},
			},
			ServiceDefinitions: []string{"example-service-definition.yml"},
			RequiredEnvVars:    []string{"ARM_SUBSCRIPTION_ID"},
			EnvConfigMapping:   map[string]string{"ARM_SUBSCRIPTION_ID": "azure.subscription_id"},
		}

		var actual brokerpak.Manifest
		err := yaml.Unmarshal([]byte(ymlManifest), &actual)

		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expectedManifest))
	})

	Describe("GetTerraformVersion", func() {
		It("returns terraform version", func() {
			exampleManifest := brokerpak.NewExampleManifest()

			actualVersion, err := exampleManifest.GetTerraformVersion()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualVersion).To(Equal(version.Must(version.NewVersion("0.13.0"))))
		})

		It("it returns error when it cant parse the terraform version", func() {
			exampleManifest := brokerpak.Manifest{
				TerraformResources: []brokerpak.TerraformResource{
					{
						Name:    "terraform",
						Version: "non-semver",
						Source:  "https://github.com/hashicorp/terraform/archive/v0.13.0.zip",
					},
				},
			}

			_, err := exampleManifest.GetTerraformVersion()
			Expect(err).To(MatchError("Malformed version: non-semver"))
		})

		It("it returns error when it cant find terraform version", func() {
			exampleManifest := brokerpak.Manifest{
				TerraformResources: []brokerpak.TerraformResource{},
			}

			_, err := exampleManifest.GetTerraformVersion()
			Expect(err).To(MatchError("terraform provider not found"))
		})
	})

	Describe("ManifestParameter", func() {
		Describe("validate", func() {
			It("should not error when manifest parameters is valid", func() {
				manifestParam := brokerpak.ManifestParameter{
					Name:        "test-name",
					Description: "best manifest",
				}

				err := manifestParam.Validate()

				Expect(err).NotTo(HaveOccurred())
			})

			It("should error when manifest parameters is missing required fields", func() {
				manifestParam := brokerpak.ManifestParameter{}

				err := manifestParam.Validate()

				Expect(err.Error()).To(Equal("missing field(s): description, name"))
			})
		})
	})
})
