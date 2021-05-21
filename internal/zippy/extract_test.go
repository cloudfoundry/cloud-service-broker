package zippy_test

import (
	"archive/zip"
	"fmt"
	"os"

	. "github.com/cloudfoundry-incubator/cloud-service-broker/internal/testmatchers"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extract", func() {
	var tmpdir string

	BeforeEach(func() {
		var err error
		tmpdir, err = os.MkdirTemp("", "test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpdir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can extract the whole zip", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())

		zr.Extract("", tmpdir)

		Expect(tmpdir).To(MatchDirectoryContents("./fixtures/brokerpak"))
	})

	It("can extract a directory within the zip", func() {
		zr, err := zippy.Open("./fixtures/brokerpak.zip")
		Expect(err).NotTo(HaveOccurred())

		zr.Extract("src", tmpdir)

		Expect(tmpdir).To(MatchDirectoryContents("./fixtures/brokerpak/src"))
	})

	When("the source does not exist", func() {
		It("extracts nothing", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			zr.Extract("not/there", tmpdir)

			dr, err := os.ReadDir(tmpdir)
			Expect(err).NotTo(HaveOccurred())
			Expect(dr).To(BeEmpty())
		})
	})

	When("the target cannot be written", func() {
		It("returns an appropriate error", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			err = zr.Extract("", "/dev/zero/cannot/write/here")
			Expect(err).To(MatchError(ContainSubstring("copy couldn't open destination")))
		})
	})

	When("a zip path contains `..`", func() {
		var zipfile string

		BeforeEach(func() {
			fd, err := os.CreateTemp("", "")
			Expect(err).NotTo(HaveOccurred())
			zipfile = fd.Name()
			defer fd.Close()

			w := zip.NewWriter(fd)
			defer w.Close()

			zd, err := w.CreateHeader(&zip.FileHeader{Name: "foo/../../baz"})
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(zd, "hello")
		})

		AfterEach(func() {
			os.Remove(zipfile)
		})

		It("returns an appropriate error", func() {
			zr, err := zippy.Open(zipfile)
			Expect(err).NotTo(HaveOccurred())

			err = zr.Extract("", tmpdir)
			Expect(err).To(MatchError(`potential zip slip extracting "foo/../../baz"`))
		})
	})
})
