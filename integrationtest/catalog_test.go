package integrationtest_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	_ "gorm.io/driver/sqlite"
)

var _ = Describe("Catalog", func() {

	var (
		originalDir      string
		fixturesDir      string
		workDir          string
		brokerPort       int
		brokerUsername   string
		brokerPassword   string
		brokerSession    *Session
		databaseFile     string
		runBrokerCommand *exec.Cmd
	)

	BeforeEach(func() {
		var err error
		originalDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = path.Join(originalDir, "fixtures", "brokerpak-for-catalog-test")

		workDir, err = os.MkdirTemp("", "*-csb-test")
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(workDir)
		Expect(err).NotTo(HaveOccurred())

		buildBrokerpakCommand := exec.Command(csb, "pak", "build", fixturesDir)
		session, err := Start(buildBrokerpakCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 10*time.Minute).Should(Exit(0))

		brokerUsername = uuid.New()
		brokerPassword = uuid.New()
		brokerPort = freePort()
		databaseFile = path.Join(workDir, "databaseFile.dat")
		runBrokerCommand = exec.Command(csb, "serve")
	})

	AfterEach(func() {
		brokerSession.Terminate()

		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("a service offering has duplicate plan IDs", func() {
		JustBeforeEach(func() {
			userProvidedPlan := "[{\"name\": \"user-plan\",\"id\":\"8b52a460-b246-11eb-a8f5-d349948e2480\"}]"
			runBrokerCommand.Env = append(
				os.Environ(),
				"CSB_LISTENER_HOST=localhost",
				"DB_TYPE=sqlite3",
				fmt.Sprintf("DB_PATH=%s", databaseFile),
				fmt.Sprintf("PORT=%d", brokerPort),
				fmt.Sprintf("SECURITY_USER_NAME=%s", brokerUsername),
				fmt.Sprintf("SECURITY_USER_PASSWORD=%s", brokerPassword),
				fmt.Sprintf("GSB_SERVICE_ALPHA_SERVICE_PLANS=%s", userProvidedPlan),
			)
		})

		It("fails to start", func() {
			var err error
			brokerSession, err = Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			brokerSession.Wait(time.Minute)

			Expect(brokerSession.ExitCode()).NotTo(BeZero())
			Expect(brokerSession.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: Plans\\[1\\].Id\n"))
		})
	})

	When("two service offerings have duplicate plan IDs", func() {
		JustBeforeEach(func() {
			userProvidedPlan := "[{\"name\": \"user-plan\",\"id\":\"8b52a460-b246-11eb-a8f5-d349948e2480\"}]"
			runBrokerCommand.Env = append(
				os.Environ(),
				"CSB_LISTENER_HOST=localhost",
				"DB_TYPE=sqlite3",
				fmt.Sprintf("DB_PATH=%s", databaseFile),
				fmt.Sprintf("PORT=%d", brokerPort),
				fmt.Sprintf("SECURITY_USER_NAME=%s", brokerUsername),
				fmt.Sprintf("SECURITY_USER_PASSWORD=%s", brokerPassword),
				fmt.Sprintf("GSB_SERVICE_BETA_SERVICE_PLANS=%s", userProvidedPlan),
			)
		})

		It("fails to start", func() {
			var err error
			brokerSession, err = Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			brokerSession.Wait(10 * time.Minute)

			Expect(brokerSession.ExitCode()).NotTo(BeZero())
			Expect(brokerSession.Err).To(Say("duplicated value, must be unique: 8b52a460-b246-11eb-a8f5-d349948e2480: services\\[1\\].Plans\\[1\\].Id\n"))
		})
	})
})
