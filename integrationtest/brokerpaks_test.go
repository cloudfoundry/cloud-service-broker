package integrationtest_test

import (
	"os"
	"path"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/integrationtest/helper"
	"github.com/cloudfoundry/cloud-service-broker/internal/zippy"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Brokerpaks", func() {
	var testHelper *helper.TestHelper

	BeforeEach(func() {
		testHelper = helper.New(csb)
	})

	AfterEach(func() {
		testHelper.Restore()
	})

	When("duplicate plan IDs", func() {
		It("fails to build", func() {
			testLab := helper.New(csb)
			command := testLab.BuildBrokerpakCommand(testHelper.OriginalDir, "fixtures", "brokerpak-with-duplicate-plan-id")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(10 * time.Minute)

			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services\\[1\\].plans\\[1\\].ID\n"))
		})
	})

	Describe("file inclusion", func() {
		It("includes the files", func() {
			testLab := helper.New(csb)

			err := os.WriteFile(path.Join(testLab.Dir, "extrafile.sh"), []byte("echo hello"), 0777)
			Expect(err).NotTo(HaveOccurred())

			brokerpakPath := path.Join(testLab.Dir, "fake-brokerpak-0.1.0.brokerpak")
			By("building the brokerpak", func() {
				command := testLab.BuildBrokerpakCommand(testHelper.OriginalDir, "fixtures", "brokerpak-file-inclusion")
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(10 * time.Minute)

				Expect(session.ExitCode()).To(BeZero())

				Expect(brokerpakPath).To(BeAnExistingFile())
			})

			extractedPath := GinkgoT().TempDir()
			By("unzipping the brokerpak", func() {
				zr, err := zippy.Open(brokerpakPath)
				Expect(err).NotTo(HaveOccurred())

				err = zr.ExtractDirectory("", extractedPath)
				Expect(err).NotTo(HaveOccurred())
			})

			By("checking that the expected files are there", func() {
				paths := []string{
					"bin/linux/amd64/0.12.21/terraform",
					"bin/linux/amd64/0.13.4/terraform",
					"bin/linux/amd64/cloud-service-broker.linux",
					"bin/linux/amd64/extrafile.sh",
				}
				for _, p := range paths {
					Expect(path.Join(extractedPath, p)).To(BeAnExistingFile())
				}
			})
		})
	})
})
