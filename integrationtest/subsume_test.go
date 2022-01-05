package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pivotal-cf/brokerapi/v8/domain"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	_ "gorm.io/driver/sqlite"
)

var _ = Describe("Subsume", func() {

	var (
		originalDir         string
		fixturesDir         string
		workDir             string
		brokerPort          int
		brokerUsername      string
		brokerPassword      string
		brokerSession       *Session
		brokerClient        *client.Client
		databaseFile        string
		serviceInstanceGUID string
	)

	JustBeforeEach(func() {
		var err error
		originalDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = path.Join(originalDir, "fixtures", "brokerpak-for-subsume-cancel")

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
		runBrokerCommand := exec.Command(csb, "serve")
		runBrokerCommand.Env = append(
			os.Environ(),
			"CSB_LISTENER_HOST=localhost",
			"DB_TYPE=sqlite3",
			fmt.Sprintf("DB_PATH=%s", databaseFile),
			fmt.Sprintf("PORT=%d", brokerPort),
			fmt.Sprintf("SECURITY_USER_NAME=%s", brokerUsername),
			fmt.Sprintf("SECURITY_USER_PASSWORD=%s", brokerPassword),
		)
		brokerSession, err = Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		waitForBrokerToStart(brokerPort)

		brokerClient, err = client.New(brokerUsername, brokerPassword, "localhost", brokerPort)
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		brokerSession.Terminate()

		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can subsume a resource", func() {
		const serviceOfferingGUID = "547cad88-fa93-11eb-9f44-97feefe52547"
		const servicePlanGUID = "59624c68-fa93-11eb-9081-e79b0e1ab5ae"
		serviceInstanceGUID = uuid.New()
		provisionResponse := brokerClient.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(`{"value":"a97fd57a-fa94-11eb-8256-930255607a99"}`))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		Eventually(func() bool {
			lastOperationResponse := brokerClient.LastOperation(serviceInstanceGUID, requestID())
			Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
			Expect(lastOperationResponse.StatusCode).To(Equal(http.StatusOK))
			var receiver domain.LastOperation
			err := json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
			Expect(err).NotTo(HaveOccurred())
			Expect(receiver.State).NotTo(Equal("failed"))
			return receiver.State == "succeeded"
		}, time.Minute*2, time.Second*10).Should(BeTrue())
	})

	It("cancels a subsume operation when a resource would be deleted", func() {
		// This test relies on a behaviour in the random string resource where it gets re-created after being imported
		const serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		const servicePlanGUID = "8b52a460-b246-11eb-a8f5-d349948e2481"
		serviceInstanceGUID = uuid.New()
		provisionResponse := brokerClient.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(`{"value":"thisisnotrandomatall"}`))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		var receiver domain.LastOperation
		Eventually(func() bool {
			lastOperationResponse := brokerClient.LastOperation(serviceInstanceGUID, requestID())
			Expect(lastOperationResponse.Error).NotTo(HaveOccurred())
			Expect(lastOperationResponse.StatusCode).To(Equal(http.StatusOK))
			err := json.Unmarshal(lastOperationResponse.ResponseBody, &receiver)
			Expect(err).NotTo(HaveOccurred())
			Expect(receiver.State).NotTo(Equal("succeeded"))
			return receiver.State == "failed"
		}, time.Minute*2, time.Second*10).Should(BeTrue())
		Expect(receiver.Description).To(Equal("terraform plan shows that resources would be destroyed - cancelling subsume"))
	})
})
