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
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/platform"
)

const manifestName = "manifest.yml"

// NewExampleManifest creates a new manifest with sample values for the service broker suitable for giving a user a template to manually edit.
func NewExampleManifest() manifest.Manifest {
	return manifest.Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []platform.Platform{
			{Os: "linux", Arch: "386"},
			{Os: "linux", Arch: "amd64"},
		},
		TerraformResources: []manifest.TerraformResource{
			{
				Name:    "terraform",
				Version: "0.13.0",
				Source:  "https://github.com/hashicorp/terraform/archive/v0.13.0.zip",
			},
			{
				Name:    "terraform-provider-google-beta",
				Version: "1.19.0",
				Source:  "https://github.com/terraform-providers/terraform-provider-google/archive/v1.19.0.zip",
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []manifest.Parameter{
			{Name: "MY_ENVIRONMENT_VARIABLE", Description: "Set this to whatever you like."},
		},
	}
}
