package reader_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/reader"
	"github.com/cloudfoundry/cloud-service-broker/internal/tfproviderfqn"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
	"github.com/cloudfoundry/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry/cloud-service-broker/utils/stream"
)

var _ = Describe("reader", func() {
	Describe("ExtractPlatformBins", func() {
		const (
			terraformV12 = "0.12.0"
			terraformV13 = "0.13.0"
		)

		Context("providers in terraform 0.12 and lower", func() {
			It("extracts providers to same directory", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV12),
					withProvider("", "terraform-provider-google-beta", "1.19.0", "x4"),
					withProvider("", "terraform-provider-google", "1.19.0", "x5"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				Expect(filepath.Join(binOutput, "terraform-provider-google-beta_v1.19.0_x4")).To(BeAnExistingFile())
				Expect(filepath.Join(binOutput, "terraform-provider-google_v1.19.0_x5")).To(BeAnExistingFile())
			})
		})

		Context("providers in terraform 0.13 and higher", func() {
			It("extracts providers to a directory hierarchy", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV13),
					withProvider("", "terraform-provider-google-beta", "1.19.0", "x4"),
					withProvider("other-namespace/google", "terraform-provider-google", "1.19.0", "x5"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				plat := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
				hashicorpBinOutput := filepath.Join(binOutput, "registry.terraform.io", "hashicorp")
				Expect(filepath.Join(hashicorpBinOutput, "google-beta", "1.19.0", plat, "terraform-provider-google-beta_v1.19.0_x4")).To(BeAnExistingFile())

				otherNamespaceBinOutput := filepath.Join(binOutput, "registry.terraform.io", "other-namespace")
				Expect(filepath.Join(otherNamespaceBinOutput, "google", "1.19.0", plat, "terraform-provider-google_v1.19.0_x5")).To(BeAnExistingFile())
			})
		})

		Context("single version of terraform", func() {
			It("extracts correctly", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV13),
					withProvider("", "terraform-provider-google-beta", "1.19.0", "x4"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				data, err := os.ReadFile(filepath.Join(binOutput, "versions", terraformV13, "terraform"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal([]byte(terraformV13)))
			})
		})

		Context("multiple terraform versions", func() {
			It("extracts terraform versions into different directories", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV12),
					withTerraform(terraformV13),
					withDefaultTerraform("1.1.1"),
					withProvider("", "terraform-provider-google-beta", "1.19.0", "x4"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				By("checking for v0.12")
				data, err := os.ReadFile(filepath.Join(binOutput, "versions", terraformV12, "terraform"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal([]byte(terraformV12)))

				By("checking for v0.13")
				data, err = os.ReadFile(filepath.Join(binOutput, "versions", terraformV13, "terraform"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal([]byte(terraformV13)))

				By("checking for v1.1.1")
				data, err = os.ReadFile(filepath.Join(binOutput, "versions", "1.1.1", "terraform"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal([]byte("1.1.1")))
			})
		})

		Context("multiple providers share same name and version", func() {
			It("should return an error", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV13),
					withProvider("", "terraform-provider-google-beta", "1.19.0", "x4"),
					withProvider("", "terraform-provider-google-beta", "1.19.0", "x5"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				filePrefix := fmt.Sprintf("bin/%s/%s", runtime.GOOS, runtime.GOARCH)
				binOutput := GinkgoT().TempDir()

				err = pakReader.ExtractPlatformBins(binOutput)
				Expect(err).To(MatchError(fmt.Sprintf("multiple files found with prefix \"%[1]s/terraform-provider-google-beta_v1.19.0\": %[1]s/terraform-provider-google-beta_v1.19.0_x4, %[1]s/terraform-provider-google-beta_v1.19.0_x5", filePrefix)))
			})
		})

		Context("terraform provider in manifest not found in zip", func() {
			It("should return an error", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV13),
					withMissingProvider("terraform-provider-google-beta", "1.19.0"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				err = pakReader.ExtractPlatformBins(binOutput)
				Expect(err).To(MatchError(fmt.Sprintf(`file with prefix "bin/%s/%s/terraform-provider-google-beta_v1.19.0" not found in zip`, runtime.GOOS, runtime.GOARCH)))
			})
		})
	})

	Describe("including source", func() {
		It("does not include source by default", func() {
			pk := fakeBrokerpak(withProvider("", "terraform-provider-fake", "1.2.3", "x1"))

			reader, err := zippy.Open(pk)
			Expect(err).NotTo(HaveOccurred())
			output := GinkgoT().TempDir()
			_ = reader.ExtractDirectory("", output)

			Expect(filepath.Join(output, "src")).NotTo(BeAnExistingFile())
		})

		It("can include the source", func() {
			pk := fakeBrokerpak(withSource(), withProvider("", "terraform-provider-fake", "1.2.3", "x1"))

			reader, err := zippy.Open(pk)
			Expect(err).NotTo(HaveOccurred())
			output := GinkgoT().TempDir()
			_ = reader.ExtractDirectory("", output)

			Expect(filepath.Join(output, "src", "terraform-provider-fake.zip")).To(BeAnExistingFile())
		})
	})
})

type config struct {
	manifest      *manifest.Manifest
	dir           string
	includeSource bool
}

type option func(c *config)

func fakeBrokerpak(opts ...option) string {
	dir := GinkgoT().TempDir()

	m := &manifest.Manifest{
		PackVersion: 1,
		Name:        "fake-brokerpack",
		Version:     "1.0.0",
		Metadata: map[string]string{
			"author": "me@example.com",
		},
		Platforms: []platform.Platform{
			{Os: "linux", Arch: "amd64"},
			{Os: "darwin", Arch: "amd64"},
		},
		ServiceDefinitions: []string{"example-service-definition.yml"},
		Parameters: []manifest.Parameter{
			{Name: "TEST_PARAM", Description: "An example paramater that will be injected into Terraform's environment variables."},
		},
		EnvConfigMapping: map[string]string{"ENV_VAR": "env.var"},
	}

	c := config{
		manifest: m,
		dir:      dir,
	}
	for _, o := range opts {
		o(&c)
	}

	Expect(stream.Copy(stream.FromYaml(m), stream.ToFile(dir, "manifest.yml"))).NotTo(HaveOccurred())

	for _, sdPath := range m.ServiceDefinitions {
		Expect(stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition()), stream.ToFile(dir, sdPath))).NotTo(HaveOccurred())
	}

	packName := path.Join(GinkgoT().TempDir(), "fake.brokerpak")
	Expect(packer.Pack(m, dir, packName, "", c.includeSource, false)).NotTo(HaveOccurred())
	return packName
}

func withTerraform(tfVersion string) option {
	return func(c *config) {
		fakeFile := filepath.Join(c.dir, tfVersion, "terraform")
		Expect(stream.Copy(stream.FromString(tfVersion), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		c.manifest.TerraformVersions = append(c.manifest.TerraformVersions, manifest.TerraformVersion{
			Version:     version.Must(version.NewVersion(tfVersion)),
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}

func withDefaultTerraform(tfVersion string) option {
	return func(c *config) {
		fakeFile := filepath.Join(c.dir, tfVersion, "terraform")
		Expect(stream.Copy(stream.FromString(tfVersion), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		c.manifest.TerraformVersions = append(c.manifest.TerraformVersions, manifest.TerraformVersion{
			Version:     version.Must(version.NewVersion(tfVersion)),
			Default:     true,
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}

func withProvider(provider, name, providerVersion, suffix string) option {
	return func(c *config) {
		fakeFile := filepath.Join(c.dir, fmt.Sprintf("%s_v%s_%s", name, providerVersion, suffix))
		Expect(stream.Copy(stream.FromString("dummy-file"), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		c.manifest.TerraformProviders = append(c.manifest.TerraformProviders, manifest.TerraformProvider{
			Name:        name,
			Version:     version.Must(version.NewVersion(providerVersion)),
			Provider:    tfproviderfqn.Must(name, provider),
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}

func withMissingProvider(name, providerVersion string) option {
	return func(c *config) {
		fakeFile := filepath.Join(c.dir, "file-name-does-not-match")
		Expect(stream.Copy(stream.FromString("dummy-file"), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		c.manifest.TerraformProviders = append(c.manifest.TerraformProviders, manifest.TerraformProvider{
			Name:        name,
			Version:     version.Must(version.NewVersion(providerVersion)),
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}

func withSource() option {
	return func(c *config) {
		c.includeSource = true
	}
}
