package manifest_test

import (
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/manifest"
)

var _ = Describe("DefaultTerraformVersion", func() {
	It("returns tofu version", func() {
		m, err := manifest.Parse(fakeManifest())
		Expect(err).NotTo(HaveOccurred())

		actualVersion, err := m.DefaultTerraformVersion()
		Expect(err).NotTo(HaveOccurred())
		Expect(actualVersion).To(Equal(version.Must(version.NewVersion("1.1.4"))))
	})

	It("it returns error when it can't find tofu version", func() {
		var exampleManifest manifest.Manifest

		_, err := exampleManifest.DefaultTerraformVersion()
		Expect(err).To(MatchError("tofu not found"))
	})

	When("there is more than one tofu version", func() {
		It("returns the default version", func() {
			m, err := manifest.Parse(fakeManifest(
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.1.5",
					"default": false,
				}),
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
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
