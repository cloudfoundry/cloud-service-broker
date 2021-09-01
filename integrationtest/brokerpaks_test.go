package integrationtest_test

import (
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/zippy"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Brokerpaks", func() {
	var (
		originalDir string
		fixturesDir string
		workDir     string
	)

	BeforeEach(func() {
		var err error
		originalDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = path.Join(originalDir, "fixtures")

		workDir, err = os.MkdirTemp("", "*-csb-test")
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("duplicate plan IDs", func() {
		It("fails to build", func() {
			command := exec.Command(csb, "pak", "build", path.Join(fixturesDir, "brokerpak-with-duplicate-plan-id"))
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			session.Wait(10 * time.Minute)

			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services\\[1\\].plans\\[1\\].Id\n"))
		})
	})

	Describe("file inclusion", func() {
		BeforeEach(func() {
			err := os.WriteFile(path.Join(workDir, "extrafile.sh"), []byte("echo hello"), 0777)
			Expect(err).NotTo(HaveOccurred())
		})

		It("includes the files", func() {
			brokerpakPath := path.Join(workDir, "fake-brokerpak-0.1.0.brokerpak")
			By("building the brokerpak", func() {
				command := exec.Command(csb, "pak", "build", path.Join(fixturesDir, "brokerpak-file-inclusion"))
				session, err := Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				session.Wait(10 * time.Minute)

				Expect(session.ExitCode()).To(BeZero())

				Expect(brokerpakPath).To(BeAnExistingFile())
			})

			extractedPath := path.Join(workDir, "extracted")
			By("unzipping the brokerpak", func() {
				zr, err := zippy.Open(brokerpakPath)
				Expect(err).NotTo(HaveOccurred())

				err = zr.Extract("", extractedPath)
				Expect(err).NotTo(HaveOccurred())
			})

			By("checking that the expected files are there", func() {
				paths := []string{
					"bin/linux/amd64/terraform",
					"bin/linux/amd64/cloud-service-broker.linux",
					"bin/linux/amd64/extrafile.sh",
					"src/terraform.zip",
				}
				for _, p := range paths {
					Expect(path.Join(extractedPath, p)).To(BeAnExistingFile())
				}
			})
		})
	})
})
