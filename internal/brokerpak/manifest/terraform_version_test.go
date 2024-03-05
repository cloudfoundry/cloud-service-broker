package manifest_test

import (
	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/manifest"
)

var _ = Describe("DefaultTerraformVersion", func() {
	It("returns terraform version", func() {
		m, err := manifest.Parse(fakeManifest())
		Expect(err).NotTo(HaveOccurred())

		actualVersion, err := m.DefaultTerraformVersion()
		Expect(err).NotTo(HaveOccurred())
		Expect(actualVersion).To(Equal(version.Must(version.NewVersion("1.6.0"))))
	})

	It("it returns error when it can't find terraform version", func() {
		var exampleManifest manifest.Manifest

		_, err := exampleManifest.DefaultTerraformVersion()
		Expect(err).To(MatchError("terraform not found"))
	})

	When("there is more than one terraform version", func() {
		It("returns the default version", func() {
			m, err := manifest.Parse(fakeManifest(
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.6.1",
					"default": false,
				}),
				withAdditionalEntry("terraform_binaries", map[string]any{
					"name":    "tofu",
					"version": "1.6.2",
					"default": true,
				}),
			))
			Expect(err).NotTo(HaveOccurred())

			actualVersion, err := m.DefaultTerraformVersion()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualVersion).To(Equal(version.Must(version.NewVersion("1.6.2"))))
		})
	})
})
