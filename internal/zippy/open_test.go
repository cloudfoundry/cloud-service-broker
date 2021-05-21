package zippy_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
