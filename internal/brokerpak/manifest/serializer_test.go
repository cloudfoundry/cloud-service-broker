package manifest_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/brokerpak/manifest"
)

var _ = Describe("TerraformResource", func() {
	It("can serialize what was parsed", func() {
		p, err := manifest.Parse(testManifest)
		Expect(err).NotTo(HaveOccurred())

		s, err := p.Serialize()
		Expect(err).NotTo(HaveOccurred())

		Expect(s).To(MatchYAML(testManifest))
	})
})
