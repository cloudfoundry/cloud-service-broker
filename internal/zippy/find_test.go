package zippy_test

import (
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Find", func() {
	It("finds a file in the zip", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())

		zf := zr.Find("foo/bar/baz.txt")
		Expect(zf).NotTo(BeNil())

		r, err := zf.Open()
		Expect(err).NotTo(HaveOccurred())
		defer r.Close()

		Eventually(BufferReader(r)).Should(Say("some stuff"))
	})

	When("the file does not exist", func() {
		It("returns nil", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			zf := zr.Find("not/there")
			Expect(zf).To(BeNil())
		})
	})
})
