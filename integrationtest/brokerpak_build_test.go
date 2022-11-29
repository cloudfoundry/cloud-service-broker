package integrationtest_test

import (
	"fmt"
	"path"
	"runtime"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/packer"
	"github.com/cloudfoundry/cloud-service-broker/internal/brokerpak/platform"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Brokerpak Build", func() {
	When("duplicate plan IDs", func() {
		It("fails to build", func() {
			_, err := packer.BuildBrokerpak(csb, fixtures("brokerpak-build-duplicate-plan-id"))
			Expect(err).To(MatchError(ContainSubstring(`duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services[1].plans[1].ID`)), err.Error())
		})
	})

	Describe("file inclusion", func() {
		It("includes the files", func() {

			By("building the brokerpak")
			brokerpak := must(packer.BuildBrokerpak(
				csb,
				fixtures("brokerpak-build-file-inclusion"),
				packer.WithExtraFile("extrafile.sh", "echo hello"),
			))
			DeferCleanup(func() {
				cleanup(brokerpak)
			})

			brokerpakPath := path.Join(string(brokerpak), "fake-brokerpak-0.1.0.brokerpak")
			Expect(brokerpakPath).To(BeAnExistingFile())

			extractionPath := GinkgoT().TempDir()
			By("unzipping the brokerpak")
			zr, err := zippy.Open(brokerpakPath)
			Expect(err).NotTo(HaveOccurred())

			err = zr.ExtractDirectory("", extractionPath)
			Expect(err).NotTo(HaveOccurred())

			By("checking that the expected files are there")
			paths := []string{
				fmt.Sprintf("bin/%s/0.12.21/terraform", platform.CurrentPlatform()),
				fmt.Sprintf("bin/%s/0.13.4/terraform", platform.CurrentPlatform()),
				fmt.Sprintf("bin/%s/cloud-service-broker.%s", platform.CurrentPlatform(), runtime.GOOS),
				fmt.Sprintf("bin/%s/extrafile.sh", platform.CurrentPlatform()),
			}
			for _, p := range paths {
				Expect(path.Join(extractionPath, p)).To(BeAnExistingFile())
			}
		})
	})
})
