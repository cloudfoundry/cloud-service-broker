package packer_test

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"

	"github.com/hashicorp/go-version"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testManifest() *manifest.Manifest {
	return &manifest.Manifest{}
}

// TODO: can we avoid passing URL template if we change the logic in the packer?
func fakeManifestTofu() *manifest.Manifest {
	return &manifest.Manifest{
		Platforms: []platform.Platform{
			{
				Os:   "linux",
				Arch: "amd64",
			},
		},
		TerraformVersions: []manifest.TerraformVersion{
			{
				Version:     version.Must(version.NewVersion("1.6.0")),
				Source:      "https://github.com/opentofu/opentofu/archive/refs/tags/v1.6.0.zip",
				URLTemplate: "https://github.com/opentofu/opentofu/releases/download/v${version}/tofu_${version}_${os}_${arch}.zip",
				Default:     true,
			},
		},
	}
}

func fakeManifestTwoTofus() *manifest.Manifest {
	return &manifest.Manifest{
		Platforms: []platform.Platform{
			{
				Os:   "linux",
				Arch: "amd64",
			},
		},
		TerraformVersions: []manifest.TerraformVersion{
			{
				Version:     version.Must(version.NewVersion("1.6.0")),
				Source:      "https://github.com/opentofu/opentofu/archive/refs/tags/v1.6.0.zip",
				URLTemplate: "https://github.com/opentofu/opentofu/releases/download/v${version}/tofu_${version}_${os}_${arch}.zip",
			},
			{
				Version:     version.Must(version.NewVersion("1.6.1")),
				Source:      "https://github.com/opentofu/opentofu/archive/refs/tags/v1.6.1.zip",
				URLTemplate: "https://github.com/opentofu/opentofu/releases/download/v${version}/tofu_${version}_${os}_${arch}.zip",
				Default:     true,
			},
		},
	}
}

func fakeManifestProviders() *manifest.Manifest {
	return &manifest.Manifest{
		Platforms: []platform.Platform{
			{
				Os:   "linux",
				Arch: "amd64",
			},
		},
		TerraformProviders: []manifest.TerraformProvider{{
			Name:    "terraform-provider-random",
			Version: version.Must(version.NewVersion("3.6.0")),
			Source:  "https://github.com/terraform-providers/terraform-provider-random/archive/v3.6.0.zip",
		}},
	}
}

var _ = Describe("Pack", func() {
	Context("packing binaries", func() {
		It("packs empty binaries", func() {
			zipOutputDir := GinkgoT().TempDir()
			zipOutputFile := filepath.Join(zipOutputDir, "packdest")

			err := packer.Pack(testManifest(), "testmanifest", zipOutputFile, "", false, false)
			Expect(err).ToNot(HaveOccurred())
			reader, err := zippy.Open(zipOutputFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader.List()).To(HaveLen(1))
			Expect(fileNames(reader)).To(ContainElements("manifest.yml"))

			data, err := readManifest(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(ContainSubstring("terraform_binaries: []"))

		})

		It("packs one tofu binary", func() {
			zipOutputDir := GinkgoT().TempDir()
			zipOutputFile := filepath.Join(zipOutputDir, "packdest")

			err := packer.Pack(fakeManifestTofu(), "testmanifest", zipOutputFile, "", false, false)
			Expect(err).ToNot(HaveOccurred())
			reader, err := zippy.Open(zipOutputFile)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileNames(reader)).To(ContainElements("manifest.yml", "bin/linux/amd64/1.6.0/tofu"))

			data, err := readManifest(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(ContainSubstring("- name: tofu"))
			Expect(data).To(ContainSubstring("version: 1.6.0"))

		})

		It("packs two tofu binaries", func() {
			zipOutputDir := GinkgoT().TempDir()
			zipOutputFile := filepath.Join(zipOutputDir, "packdest")

			err := packer.Pack(fakeManifestTwoTofus(), "testmanifest", zipOutputFile, "", false, false)
			Expect(err).ToNot(HaveOccurred())
			reader, err := zippy.Open(zipOutputFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileNames(reader)).To(ContainElements("manifest.yml", "bin/linux/amd64/1.6.0/tofu", "bin/linux/amd64/1.6.1/tofu"))

			data, err := readManifest(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(ContainSubstring("- name: tofu"))
			Expect(data).To(ContainSubstring("version: 1.6.0"))
			Expect(data).To(ContainSubstring("version: 1.6.1"))
		})

		It("packs providers", func() {
			zipOutputDir := GinkgoT().TempDir()
			zipOutputFile := filepath.Join(zipOutputDir, "packdest")

			err := packer.Pack(fakeManifestProviders(), "testmanifest", zipOutputFile, "", false, false)
			Expect(err).ToNot(HaveOccurred())
			reader, err := zippy.Open(zipOutputFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileNames(reader)).To(ContainElements("manifest.yml", "bin/linux/amd64/terraform-provider-random_v3.6.0_x5"))

			data, err := readManifest(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(ContainSubstring("- name: terraform-provider-random"))
			Expect(data).To(ContainSubstring("version: 3.6.0"))
		})

	})

	Context("Packing sources", func() {
		It("packs one tofu binary's sources", func() {
			zipOutputDir := GinkgoT().TempDir()
			zipOutputFile := filepath.Join(zipOutputDir, "packdest")

			err := packer.Pack(fakeManifestTofu(), "testmanifest", zipOutputFile, "", true, false)
			Expect(err).ToNot(HaveOccurred())
			reader, err := zippy.Open(zipOutputFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileNames(reader)).To(ContainElements("manifest.yml", "src/", "src/tofu.zip"))
		})

		// TODO has this ever worked as it should? sources don't seem to be versioned
		It("packs two tofu binary's sources - paks only one", func() {
			zipOutputDir := GinkgoT().TempDir()
			zipOutputFile := filepath.Join(zipOutputDir, "packdest")

			err := packer.Pack(fakeManifestTwoTofus(), "testmanifest", zipOutputFile, "", true, false)
			Expect(err).ToNot(HaveOccurred())
			reader, err := zippy.Open(zipOutputFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(fileNames(reader)).To(ContainElements("manifest.yml", "src/", "src/tofu.zip"))
		})

	})

})

func readManifest(reader zippy.ZipReader) ([]byte, error) {
	binOutput := GinkgoT().TempDir()
	Expect(reader.ExtractFile("manifest.yml", binOutput)).ToNot(HaveOccurred())
	return os.ReadFile(filepath.Join(binOutput, "manifest.yml"))
}

func fileNames(reader zippy.ZipReader) []string {
	var fileNames []string
	for _, f := range reader.List() {
		fileNames = append(fileNames, f.Name)

	}
	return fileNames
}
