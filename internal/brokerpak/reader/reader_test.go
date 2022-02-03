package reader_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/reader"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/platform"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/packer"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/stream"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
					withProvider("terraform-provider-google-beta", "1.19.0", "x4"),
					withProvider("terraform-provider-google", "1.19.0", "x5"),
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
					withProvider("terraform-provider-google-beta", "1.19.0", "x4"),
					withProvider("terraform-provider-google", "1.19.0", "x5"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				plat := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
				binOutput = filepath.Join(binOutput, "registry.terraform.io", "hashicorp")
				Expect(filepath.Join(binOutput, "google-beta", "1.19.0", plat, "terraform-provider-google-beta_v1.19.0_x4")).To(BeAnExistingFile())
				Expect(filepath.Join(binOutput, "google", "1.19.0", plat, "terraform-provider-google_v1.19.0_x5")).To(BeAnExistingFile())
			})
		})

		Context("single version of terraform", func() {
			It("extracts correctly", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV13),
					withProvider("terraform-provider-google-beta", "1.19.0", "x4"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				By("checking for the default version link")
				details, err := os.Readlink(filepath.Join(binOutput, "terraform"))
				Expect(err).NotTo(HaveOccurred())
				Expect(details).To(Equal("versions/0.13.0/terraform"))

				By("checking that the link points to something")
				Expect(filepath.Join(binOutput, "versions", terraformV13, "terraform")).To(BeAnExistingFile())
			})
		})

		Context("multiple terraform versions", func() {
			It("extracts terraform versions into different directories", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV12),
					withTerraform(terraformV13),
					withDefaultTerraform("1.1.1"),
					withProvider("terraform-provider-google-beta", "1.19.0", "x4"),
				)

				pakReader, err := reader.OpenBrokerPak(pk)
				Expect(err).NotTo(HaveOccurred())

				binOutput := GinkgoT().TempDir()
				Expect(pakReader.ExtractPlatformBins(binOutput)).NotTo(HaveOccurred())

				By("checking for the default version")
				details, err := os.Readlink(filepath.Join(binOutput, "terraform"))
				Expect(err).NotTo(HaveOccurred())
				Expect(details).To(Equal("versions/1.1.1/terraform"))

				By("checking for all the versions")
				Expect(filepath.Join(binOutput, "versions", terraformV12, "terraform")).To(BeAnExistingFile())
				Expect(filepath.Join(binOutput, "versions", terraformV13, "terraform")).To(BeAnExistingFile())
				Expect(filepath.Join(binOutput, "versions", "1.1.1", "terraform")).To(BeAnExistingFile())
			})
		})

		Context("multiple providers share same name and version", func() {
			It("should return an error", func() {
				pk := fakeBrokerpak(
					withTerraform(terraformV13),
					withProvider("terraform-provider-google-beta", "1.19.0", "x4"),
					withProvider("terraform-provider-google-beta", "1.19.0", "x5"),
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
})

type option func(dir string, m *manifest.Manifest)

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

	for _, o := range opts {
		o(dir, m)
	}

	Expect(stream.Copy(stream.FromYaml(m), stream.ToFile(dir, "manifest.yml"))).NotTo(HaveOccurred())

	for _, path := range m.ServiceDefinitions {
		Expect(stream.Copy(stream.FromYaml(tf.NewExampleTfServiceDefinition()), stream.ToFile(dir, path))).NotTo(HaveOccurred())
	}

	packName := path.Join(GinkgoT().TempDir(), "fake.brokerpak")
	Expect(packer.Pack(m, dir, packName)).NotTo(HaveOccurred())
	return packName
}

func withTerraform(tfVersion string) option {
	return func(dir string, m *manifest.Manifest) {
		fakeFile := filepath.Join(dir, "terraform")
		Expect(stream.Copy(stream.FromString(tfVersion), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		m.TerraformResources = append(m.TerraformResources, manifest.TerraformResource{
			Name:        "terraform",
			Version:     tfVersion,
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}

func withDefaultTerraform(tfVersion string) option {
	return func(dir string, m *manifest.Manifest) {
		fakeFile := filepath.Join(dir, "terraform")
		Expect(stream.Copy(stream.FromString(tfVersion), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		m.TerraformResources = append(m.TerraformResources, manifest.TerraformResource{
			Name:        "terraform",
			Version:     tfVersion,
			Source:      fakeFile,
			URLTemplate: fakeFile,
			Default:     true,
		})
	}
}

func withProvider(name, providerVersion, suffix string) option {
	return func(dir string, m *manifest.Manifest) {
		fakeFile := filepath.Join(dir, fmt.Sprintf("%s_v%s_%s", name, providerVersion, suffix))
		Expect(stream.Copy(stream.FromString("dummy-file"), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		m.TerraformResources = append(m.TerraformResources, manifest.TerraformResource{
			Name:        name,
			Version:     providerVersion,
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}

func withMissingProvider(name, providerVersion string) option {
	return func(dir string, m *manifest.Manifest) {
		fakeFile := filepath.Join(dir, "file-name-does-not-match")
		Expect(stream.Copy(stream.FromString("dummy-file"), stream.ToFile(fakeFile))).NotTo(HaveOccurred())

		m.TerraformResources = append(m.TerraformResources, manifest.TerraformResource{
			Name:        name,
			Version:     providerVersion,
			Source:      fakeFile,
			URLTemplate: fakeFile,
		})
	}
}
