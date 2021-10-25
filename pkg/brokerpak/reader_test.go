package brokerpak_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pborman/uuid"

	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/brokerpak"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("reader", func() {
	Describe("ExtractPlatformBins", func() {
		const terraformV13 = "0.13.0"
		var err error
		var binOutput string
		var pk string

		BeforeEach(func() {
			binOutput, err = os.MkdirTemp("/tmp", "brokerPakBinOutput")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(pk)
			Expect(err).NotTo(HaveOccurred())

			err = os.RemoveAll(binOutput)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("terraform is not in the manifest", func() {
			It("should return an error", func() {
				pk, err = fakeBrokerPakWithNoTerraform()
				Expect(err).NotTo(HaveOccurred())
				pakReader, err := brokerpak.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				err = pakReader.ExtractPlatformBins(binOutput)
				Expect(err).To(MatchError("terraform not found in manifest"))
			})
		})

		When("Using Terraform v0.13", func() {
			Context("multiple providers share same name and version", func() {
				It("should return an error", func() {
					pk, err = fakeBrokerPakWithDuplicateProviders(terraformV13)
					Expect(err).NotTo(HaveOccurred())
					pakReader, err := brokerpak.OpenBrokerPak(pk)
					Expect(err).NotTo(HaveOccurred())

					err = pakReader.ExtractPlatformBins(binOutput)

					filePrefix := fmt.Sprintf("bin/%s/%s", runtime.GOOS, runtime.GOARCH)
					Expect(err).To(MatchError(fmt.Sprintf("multiple files found with prefix \"%[1]s/terraform-provider-google-beta_v1.19.0\": %[1]s/terraform-provider-google-beta_v1.19.0_x4, %[1]s/terraform-provider-google-beta_v1.19.0_x5", filePrefix)))
				})
			})
			Context("terraform-provider in manifest not found in zip", func() {
				It("should return an error", func() {
					pk, err = fakeBrokerPakWithMissingTerraformProvider(terraformV13)
					Expect(err).NotTo(HaveOccurred())
					pakReader, err := brokerpak.OpenBrokerPak(pk)
					Expect(err).NotTo(HaveOccurred())

					err = pakReader.ExtractPlatformBins(binOutput)
					Expect(err).To(MatchError(fmt.Sprintf("file with prefix \"bin/%s/%s/terraform-provider-google-beta_v1.19.0\" not found in zip", runtime.GOOS, runtime.GOARCH)))
				})
			})
			Context("provider in manifest not found in zip", func() {
				It("should return an error", func() {
					pk, err = fakeBrokerPakWithMissingProvider(terraformV13)
					Expect(err).NotTo(HaveOccurred())
					pakReader, err := brokerpak.OpenBrokerPak(pk)
					Expect(err).NotTo(HaveOccurred())

					err = pakReader.ExtractPlatformBins(binOutput)

					filePrefix := fmt.Sprintf("bin/%s/%s", runtime.GOOS, runtime.GOARCH)
					Expect(err).To(MatchError(fmt.Errorf(
						"error extracting %q to %q: %w",
						fmt.Sprintf("%s/some-provider", filePrefix),
						binOutput,
						fmt.Errorf("file \"%s/some-provider\" does not exist in the zip", filePrefix))))
				})
			})
		})
	})
})

func fakeBrokerPakWithNoTerraform() (string, error) {
	dir, err := os.MkdirTemp("", "fakepak")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	tfSrc := filepath.Join(dir, "terraform")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(tfSrc)); err != nil {
		return "", err
	}

	exampleManifest := &brokerpak.Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []brokerpak.Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
		// These resources are stubbed with a local dummy file
		TerraformResources: []brokerpak.TerraformResource{
			{
				Name:        "terraform-provider-google-beta",
				Version:     "1.19.0",
				Source:      tfSrc,
				UrlTemplate: tfSrc,
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []brokerpak.ManifestParameter{
			{Name: "TEST_PARAM", Description: "An example paramater that will be injected into Terraform's environment variables."},
		},
		EnvConfigMapping: map[string]string{"ENV_VAR": "env.var"},
	}

	if err := stream.Copy(stream.FromYaml(exampleManifest), stream.ToFile(dir, "manifest.yml")); err != nil {
		return "", err
	}

	for _, path := range exampleManifest.ServiceDefinitions {
		if err := stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition()), stream.ToFile(dir, path)); err != nil {
			return "", err
		}
	}

	packName := fmt.Sprintf("/tmp/%v-%s-%s.brokerpak", uuid.New(), exampleManifest.Name, "1.0.0")
	return packName, exampleManifest.Pack(dir, packName)
}
func fakeBrokerPakWithDuplicateProviders(terraformVersion string) (string, error) {
	dir, err := os.MkdirTemp("", "fakepak")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	tfSrc := filepath.Join(dir, "terraform")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(tfSrc)); err != nil {
		return "", err
	}

	providerOneSrc := filepath.Join(dir, "terraform-provider-google-beta_v1.19.0_x5")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(providerOneSrc)); err != nil {
		return "", err
	}

	providerTwoSrc := filepath.Join(dir, "terraform-provider-google-beta_v1.19.0_x4")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(providerTwoSrc)); err != nil {
		return "", err
	}

	exampleManifest := &brokerpak.Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []brokerpak.Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
		// These resources are stubbed with a local dummy file
		TerraformResources: []brokerpak.TerraformResource{
			{
				Name:        "terraform",
				Version:     terraformVersion,
				Source:      tfSrc,
				UrlTemplate: tfSrc,
			},
			{
				Name:        "terraform-provider-google-beta",
				Version:     "1.19.0",
				Source:      providerOneSrc,
				UrlTemplate: providerOneSrc,
			},
			{
				Name:        "terraform-provider-google-beta",
				Version:     "1.19.0",
				Source:      providerTwoSrc,
				UrlTemplate: providerTwoSrc,
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []brokerpak.ManifestParameter{
			{Name: "TEST_PARAM", Description: "An example paramater that will be injected into Terraform's environment variables."},
		},
		EnvConfigMapping: map[string]string{"ENV_VAR": "env.var"},
	}

	if err := stream.Copy(stream.FromYaml(exampleManifest), stream.ToFile(dir, "manifest.yml")); err != nil {
		return "", err
	}

	for _, path := range exampleManifest.ServiceDefinitions {
		if err := stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition()), stream.ToFile(dir, path)); err != nil {
			return "", err
		}
	}

	packName := fmt.Sprintf("/tmp/%v-%s-%s.brokerpak", uuid.New(), exampleManifest.Name, "1.0.0")
	return packName, exampleManifest.Pack(dir, packName)
}
func fakeBrokerPakWithMissingTerraformProvider(terraformVersion string) (string, error) {
	dir, err := os.MkdirTemp("", "fakepak")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	tfSrc := filepath.Join(dir, "terraform")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(tfSrc)); err != nil {
		return "", err
	}

	providerOneSrc := filepath.Join(dir, "some_other_provider_v1.19.0_x5")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(providerOneSrc)); err != nil {
		return "", err
	}

	exampleManifest := &brokerpak.Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []brokerpak.Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
		// These resources are stubbed with a local dummy file
		TerraformResources: []brokerpak.TerraformResource{
			{
				Name:        "terraform",
				Version:     terraformVersion,
				Source:      tfSrc,
				UrlTemplate: tfSrc,
			},
			{
				Name:        "terraform-provider-google-beta",
				Version:     "1.19.0",
				Source:      providerOneSrc,
				UrlTemplate: providerOneSrc,
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []brokerpak.ManifestParameter{
			{Name: "TEST_PARAM", Description: "An example paramater that will be injected into Terraform's environment variables."},
		},
		EnvConfigMapping: map[string]string{"ENV_VAR": "env.var"},
	}

	if err := stream.Copy(stream.FromYaml(exampleManifest), stream.ToFile(dir, "manifest.yml")); err != nil {
		return "", err
	}

	for _, path := range exampleManifest.ServiceDefinitions {
		if err := stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition()), stream.ToFile(dir, path)); err != nil {
			return "", err
		}
	}

	packName := fmt.Sprintf("/tmp/%v-%s-%s.brokerpak", uuid.New(), exampleManifest.Name, "1.0.0")
	return packName, exampleManifest.Pack(dir, packName)
}
func fakeBrokerPakWithMissingProvider(terraformVersion string) (string, error) {
	dir, err := os.MkdirTemp("", "fakepak")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	tfSrc := filepath.Join(dir, "terraform")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(tfSrc)); err != nil {
		return "", err
	}

	providerOneSrc := filepath.Join(dir, "some_other_provider_v1.19.0_x5")
	if err := stream.Copy(stream.FromString("dummy-file"), stream.ToFile(providerOneSrc)); err != nil {
		return "", err
	}

	exampleManifest := &brokerpak.Manifest{
		PackVersion: 1,
		Name:        "my-services-pack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []brokerpak.Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
		// These resources are stubbed with a local dummy file
		TerraformResources: []brokerpak.TerraformResource{
			{
				Name:        "terraform",
				Version:     terraformVersion,
				Source:      tfSrc,
				UrlTemplate: tfSrc,
			},
			{
				Name:        "some-provider",
				Version:     "1.19.0",
				Source:      providerOneSrc,
				UrlTemplate: providerOneSrc,
			},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []brokerpak.ManifestParameter{
			{Name: "TEST_PARAM", Description: "An example paramater that will be injected into Terraform's environment variables."},
		},
		EnvConfigMapping: map[string]string{"ENV_VAR": "env.var"},
	}

	if err := stream.Copy(stream.FromYaml(exampleManifest), stream.ToFile(dir, "manifest.yml")); err != nil {
		return "", err
	}

	for _, path := range exampleManifest.ServiceDefinitions {
		if err := stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition()), stream.ToFile(dir, path)); err != nil {
			return "", err
		}
	}

	packName := fmt.Sprintf("/tmp/%v-%s-%s.brokerpak", uuid.New(), exampleManifest.Name, "1.0.0")
	return packName, exampleManifest.Pack(dir, packName)
}
