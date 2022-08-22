package zippy_test

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cloud-service-broker/internal/testmatchers"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
)

var _ = Describe("extraction", func() {
	Describe("ExtractDirectory()", func() {
		var tmpdir string

		BeforeEach(func() {
			tmpdir = GinkgoT().TempDir()
		})

		It("can extract the whole zip", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			err = zr.ExtractDirectory("", tmpdir)
			Expect(err).NotTo(HaveOccurred())

			Expect(tmpdir).To(MatchDirectoryContents("./fixtures/brokerpak"))
		})

		It("can extract a directory within the zip", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			err = zr.ExtractDirectory("src", tmpdir)
			Expect(err).NotTo(HaveOccurred())

			Expect(tmpdir).To(MatchDirectoryContents("./fixtures/brokerpak/src"))
		})

		When("the source does not exist", func() {
			It("extracts nothing", func() {
				zr, err := zippy.Open("./fixtures/brokerpak.zip")
				Expect(err).NotTo(HaveOccurred())

				err = zr.ExtractDirectory("not/there", tmpdir)
				Expect(err).NotTo(HaveOccurred())

				dr, err := os.ReadDir(tmpdir)
				Expect(err).NotTo(HaveOccurred())
				Expect(dr).To(BeEmpty())
			})
		})

		When("the target cannot be written", func() {
			It("returns an appropriate error", func() {
				zr, err := zippy.Open("./fixtures/brokerpak.zip")
				Expect(err).NotTo(HaveOccurred())

				err = zr.ExtractDirectory("", "/dev/zero/cannot/write/here")
				Expect(err).To(MatchError(ContainSubstring("copy couldn't open destination")))
			})
		})

		When("a zip path contains `..`", func() {
			It("returns an appropriate error", func() {
				zipfile := createZipWithDotDot()
				defer func(name string) {
					_ = os.Remove(name)
				}(zipfile)

				zr, err := zippy.Open(zipfile)
				Expect(err).NotTo(HaveOccurred())

				err = zr.ExtractDirectory("", tmpdir)
				Expect(err).To(MatchError(`potential zip slip extracting "foo/../../baz"`))
			})
		})
	})

	Describe("ExtractFile()", func() {
		var tmpdir string

		BeforeEach(func() {
			tmpdir = GinkgoT().TempDir()
		})

		It("can extract a file from the zip", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			err = zr.ExtractFile("foo/bar/quz.txt", tmpdir)
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(tmpdir, "quz.txt")).To(BeAnExistingFile())
			entries, err := os.ReadDir(tmpdir)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
		})

		It("fails when the file does not exist", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			err = zr.ExtractFile("foo/quz.txt", tmpdir)
			Expect(err).To(MatchError(`file "foo/quz.txt" does not exist in the zip`))
		})

		It("fails when the file is a directory", func() {
			zr, err := zippy.Open("./fixtures/brokerpak.zip")
			Expect(err).NotTo(HaveOccurred())

			err = zr.ExtractFile("foo/bar", tmpdir)
			Expect(err).To(MatchError(`file "foo/bar" does not exist in the zip`))
		})

		When("the target cannot be written", func() {
			It("returns an appropriate error", func() {
				zr, err := zippy.Open("./fixtures/brokerpak.zip")
				Expect(err).NotTo(HaveOccurred())

				err = zr.ExtractFile("foo/bar/quz.txt", "/dev/zero/cannot/write/here")
				Expect(err).To(MatchError(ContainSubstring("copy couldn't open destination")))
			})
		})

		When("a zip path contains `..`", func() {
			It("returns an appropriate error", func() {
				zipfile := createZipWithDotDot()
				defer func(name string) {
					_ = os.Remove(name)
				}(zipfile)

				zr, err := zippy.Open(zipfile)
				Expect(err).NotTo(HaveOccurred())

				err = zr.ExtractFile("foo/../../baz", tmpdir)
				Expect(err).To(MatchError(`potential zip slip extracting "foo/../../baz"`))
			})
		})
	})
})

func createZipWithDotDot() string {
	fd, err := os.CreateTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	zipfile := fd.Name()
	defer func(f *os.File) {
		_ = f.Close()
	}(fd)

	w := zip.NewWriter(fd)
	defer func(zw *zip.Writer) {
		_ = zw.Close()
	}(w)

	zd, err := w.CreateHeader(&zip.FileHeader{Name: "foo/../../baz"})
	Expect(err).NotTo(HaveOccurred())
	_, _ = fmt.Fprintf(zd, "hello")

	return zipfile
}
