package manifest_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultTerraformVersion", func() {
	It("returns terraform version", func() {
		m, err := manifest.Parse(fakeManifest())
		Expect(err).NotTo(HaveOccurred())

		actualVersion, err := m.DefaultTerraformVersion()
		Expect(err).NotTo(HaveOccurred())
		Expect(actualVersion).To(Equal(version.Must(version.NewVersion("1.1.4"))))
	})

	It("returns error when it can't parse the terraform version", func() {
		exampleManifest := manifest.Manifest{
			TerraformResources: []manifest.TerraformResource{
				{
					Name:    "terraform",
					Version: "non-semver",
					Source:  "https://github.com/hashicorp/terraform/archive/v0.13.0.zip",
				},
			},
		}

		_, err := exampleManifest.DefaultTerraformVersion()
		Expect(err).To(MatchError("Malformed version: non-semver"))
	})

	It("it returns error when it can't find terraform version", func() {
		exampleManifest := manifest.Manifest{
			TerraformResources: []manifest.TerraformResource{},
		}

		_, err := exampleManifest.DefaultTerraformVersion()
		Expect(err).To(MatchError("terraform not found"))
	})

	When("there is more than one terraform version", func() {
		It("returns the default version", func() {
			m, err := manifest.Parse(fakeManifest(
				withAdditionalEntry("terraform_binaries", map[string]interface{}{
					"name":    "terraform",
					"version": "1.1.5",
					"default": false,
				}),
				withAdditionalEntry("terraform_binaries", map[string]interface{}{
					"name":    "terraform",
					"version": "1.1.6",
					"default": true,
				}),
			))
			Expect(err).NotTo(HaveOccurred())

			actualVersion, err := m.DefaultTerraformVersion()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualVersion).To(Equal(version.Must(version.NewVersion("1.1.6"))))
		})
	})
})
