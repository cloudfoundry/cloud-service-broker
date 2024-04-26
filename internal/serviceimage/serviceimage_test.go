package serviceimage_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cloud-service-broker/v3/internal/serviceimage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceImage", func() {
	It("returns a url unmodified", func() {
		testImage := serviceimage.ServiceImage("https://test-image-url.com")

		Expect(testImage.Encode("testBase")).To(Succeed())
		Expect(testImage).To(Equal(serviceimage.ServiceImage("https://test-image-url.com")))
	})

	Describe("local file", func() {
		var tmpFile string
		var testImage serviceimage.ServiceImage

		BeforeEach(func() {
			tmpFile = filepath.Join(GinkgoT().TempDir(), "test-image.png")
			Expect(os.WriteFile(tmpFile, []byte("abcd"), 0644)).To(Succeed())
			testImage = serviceimage.ServiceImage(fmt.Sprintf("file://%s", tmpFile))
		})

		It("encodes a local image", func() {
			Expect(testImage.Encode("")).To(Succeed())
			Expect(testImage).To(Equal(serviceimage.ServiceImage("YWJjZA==")))
		})

		It("returns an error if unable to read file", func() {
			Expect(testImage.Encode("bad-path")).To(MatchError(fmt.Sprintf("unable to read service image file bad-path%s", tmpFile)))
		})
	})
})
