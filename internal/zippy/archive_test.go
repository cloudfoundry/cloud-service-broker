package zippy_test

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cloud-service-broker/internal/testmatchers"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
)

var _ = Describe("Archive", func() {
	var tmpdir string

	BeforeEach(func() {
		tmpdir = GinkgoT().TempDir()
	})

	It("creates a zip", func() {
		target := path.Join(tmpdir, "test.zip")

		err := zippy.Archive("./fixtures/brokerpak", target)
		Expect(err).NotTo(HaveOccurred())

		// Although we have a brokerpak.zip fixture that we could
		// theoretically compare against, in practice this is too
		// fragile as the contents vary in a subtle way depending
		// on the system used to run the test
		zr, err := zippy.Open(target)
		Expect(err).NotTo(HaveOccurred())

		extracted := path.Join(tmpdir, "extracted")
		err = zr.ExtractDirectory("", extracted)
		Expect(err).NotTo(HaveOccurred())

		Expect(extracted).To(MatchDirectoryContents("./fixtures/brokerpak"))
	})

	When("the source does not exist", func() {
		It("returns an appropriate error", func() {
			target := path.Join(tmpdir, "test.zip")

			err := zippy.Archive("/this/does/not/exist", target)
			Expect(err).To(MatchError("lstat /this/does/not/exist: no such file or directory"))
		})
	})

	When("the target cannot be written", func() {
		It("returns an appropriate error", func() {
			err := zippy.Archive("./fixtures/brokerpak", "/this/does/not/exist")
			Expect(err).To(MatchError(`couldn't create archive "/this/does/not/exist": open /this/does/not/exist: no such file or directory`))
		})
	})
})
