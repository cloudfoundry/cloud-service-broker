package zippy_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/zippy"
)

var _ = Describe("List", func() {
	It("returns the list of contents", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())

		var names []string
		for _, fd := range zr.List() {
			names = append(names, fd.Name)
		}
		Expect(names).To(ConsistOf("bin/", "bin/hello", "foo/", "foo/bar/", "foo/bar/baz.txt", "foo/bar/quz.txt", "manifest.yml", "src/", "src/bye.sh"))
	})
})
