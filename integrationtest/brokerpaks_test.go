package integrationtest_test

import (
	"os"
	"os/exec"
	"path"
	"time"

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
			session.Wait(time.Minute)

			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services\\[1\\].plans\\[1\\].Id\n"))
		})
	})
})
