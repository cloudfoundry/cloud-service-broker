package zippy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/zippy"
)

var _ = Describe("Open", func() {
	It("returns a ZipReader", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())
		Expect(zr).To(BeAssignableToTypeOf(zippy.ZipReader{}))
	})

	When("the zip file does not exist", func() {
		It("returns an appropriate error", func() {
			_, err := zippy.Open("/this/does/not/exist")
			Expect(err).To(MatchError("open /this/does/not/exist: no such file or directory"))
		})
	})
})
