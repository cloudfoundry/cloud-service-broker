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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/providers/tf"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/v3/utils/stream"
	"github.com/hashicorp/go-version"
)

func fakeBrokerpak() (string, error) {
	dir, err := os.MkdirTemp("", "fakepak")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	tfSrc := filepath.Join(dir, "tofu")
	if err := os.WriteFile(tfSrc, []byte("dummy-file"), 0644); err != nil {
		return "", err
	}
	tfpSrc := filepath.Join(dir, "terraform-provider-google-beta_v1.19.0")
	if err := os.WriteFile(tfpSrc, []byte("dummy-file"), 0644); err != nil {
		return "", err
	}

	exampleManifest := &manifest.Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []platform.Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
		// These resources are stubbed with a local dummy file
		TerraformVersions: []manifest.TerraformVersion{
			{
				Version:     version.Must(version.NewVersion("1.6.0")),
				Source:      tfSrc,
				URLTemplate: tfSrc,
			},
		},
		TerraformProviders: []manifest.TerraformProvider{
			{
				Name:        "terraform-provider-google-beta",
				Version:     version.Must(version.NewVersion("1.19.0")),
				Source:      tfpSrc,
				URLTemplate: tfpSrc,
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []manifest.Parameter{
			{Name: "TEST_PARAM", Description: "An example parameter that will be injected into OpenTofu's environment variables."},
		},
		EnvConfigMapping: map[string]string{"ENV_VAR": "env.var"},
	}

	data, err := exampleManifest.Serialize()
	if err != nil {
		return "", err
	}

	if err := stream.Copy(stream.FromBytes(data), stream.ToFile(dir, ManifestName)); err != nil {
		return "", err
	}

	for _, path := range exampleManifest.ServiceDefinitions {
		if err := stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition("00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")), stream.ToFile(dir, path)); err != nil {
			return "", err
		}
	}

	return Pack(dir, "", true, false, platform.Platform{})
}

func ExampleValidate() {
	pk, err := fakeBrokerpak()
	defer os.Remove(pk)

	if err != nil {
		panic(err)
	}

	if err := Validate(pk); err != nil {
		panic(err)
	} else {
		fmt.Println("ok!")
	}

	// Output: ok!
}

func TestFinfo(t *testing.T) {
	pk, err := fakeBrokerpak()
	defer os.Remove(pk)

	if err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	if err := finfo(pk, buf); err != nil {
		t.Fatal(err)
	}

	// Check for "important strings" which MUST exist for this to be a valid
	// output
	importantStrings := []string{
		"Information",      // heading
		"my-services-pack", // name
		"1.0.0",            // version

		"Parameters", // heading
		"TEST_PARAM", // value

		"Dependencies",                   // heading
		"tofu",                           // dependency
		"terraform-provider-google-beta", // dependency

		"Services",                             // heading
		"00000000-0000-0000-0000-000000000000", // guid
		"example-service",                      // name

		"Contents",                               // heading
		"bin/",                                   // directory
		"definitions/",                           // directory
		"manifest.yml",                           // manifest
		"src/terraform-provider-google-beta.zip", // file
		"src/tofu.zip",                           // file

	}
	actual := buf.String()
	for _, str := range importantStrings {
		if !strings.Contains(actual, str) {
			t.Fatalf("Expected output to contain %s but it didn't", str)
		}
	}
}

func TestRegistryFromLocalBrokerpak(t *testing.T) {
	pk, err := fakeBrokerpak()
	defer os.Remove(pk)

	if err != nil {
		t.Fatalf("fakeBrokerpak: %v", err)
	}

	abs, err := filepath.Abs(pk)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}

	registry, err := registryFromLocalBrokerpak(abs)
	if err != nil {
		t.Fatalf("registryFromLocalBrokerpak: %v", err)
	}

	if len(registry) != 1 {
		t.Fatalf("Expected %d services but got %d", 1, len(registry))
	}

	svc, err := registry.GetServiceByID("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatal(err)
	}

	if svc.Name != "example-service" {
		t.Errorf("Expected exapmle-service, got %q", svc.Name)
	}
}
