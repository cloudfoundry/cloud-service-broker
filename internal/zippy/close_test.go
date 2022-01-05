package zippy_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Close", func() {
	It("can close a ZipReader", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())

		zr.Close()
		zr.Close() // idempotent
	})
})
