package freeport_test

import (
	"github.com/cloudfoundry/cloud-service-broker/utils/freeport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("freeport", func() {
	beAPort := SatisfyAll(
		BeNumerically(">", 2<<10),
		BeNumerically("<", 2<<16),
	)

	Describe("Port", func() {
		It("returns a port", func() {
			port, err := freeport.Port()
			Expect(err).NotTo(HaveOccurred())
			Expect(port).To(beAPort)
		})

		It("returns a different port", func() {
			p1, err := freeport.Port()
			Expect(err).NotTo(HaveOccurred())
			p2, err := freeport.Port()
			Expect(err).NotTo(HaveOccurred())
			Expect(p1).NotTo(Equal(p2))
		})
	})

	Describe("Must", func() {
		It("returns a port", func() {
			Expect(freeport.Must()).To(beAPort)
		})

		It("returns a different port", func() {
			Expect(freeport.Must()).NotTo(Equal(freeport.Must()))
		})
	})
})
