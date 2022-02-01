package manifest_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/brokerpak/manifest"
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetTerraformVersion", func() {
	It("returns terraform version", func() {
		m, err := test(testManifest)
		Expect(err).NotTo(HaveOccurred())

		actualVersion, err := m.GetTerraformVersion()
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

		_, err := exampleManifest.GetTerraformVersion()
		Expect(err).To(MatchError("Malformed version: non-semver"))
	})

	It("it returns error when it cant find terraform version", func() {
		exampleManifest := manifest.Manifest{
			TerraformResources: []manifest.TerraformResource{},
		}

		_, err := exampleManifest.GetTerraformVersion()
		Expect(err).To(MatchError("terraform provider not found"))
	})
})
