package zippy_test

import (
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("List", func() {
	It("returns the list of contents", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, fd := range zr.List() {
			names = append(names, fd.Name)
		}
		Expect(names).To(ConsistOf("bin/", "bin/hello", "foo/", "foo/bar/", "foo/bar/baz.txt", "manifest.yml", "src/", "src/bye.sh"))
	})
})
