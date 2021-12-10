package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

var _ = Describe("Multiple Updates", func() {
	const (
		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
	)

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

	BeforeEach(func() {
		var err error
		originalDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = path.Join(originalDir, "fixtures")

		workDir, err = os.MkdirTemp("", "*-csb-test")
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(workDir)
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(csb, "pak", "build", path.Join(fixturesDir, "brokerpak-for-multiple-updates"))
		session, err := Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		session.Wait(10 * time.Minute)

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

	checkBindingOutput := func(expected string) {
		bindResponse := brokerClient.Bind(serviceInstanceGUID, uuid.New(), serviceOfferingGUID, servicePlanGUID, requestID(), nil)
		ExpectWithOffset(1, bindResponse.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, string(bindResponse.ResponseBody)).To(ContainSubstring(expected))
	}

	waitForCompletion := func() {
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
	}

	// This test was added for issue https://www.pivotaltracker.com/story/show/178213626 where a parameter that was
	// updated would be reverted to the default value in subsequent updates
	It("persists updated parameters in subsequent updates", func() {
		By("provisioning with parameters")
		const provisionParams = `{"alpha_input":"foo","beta_input":"bar"}`
		serviceInstanceGUID = uuid.New()
		provisionResponse := brokerClient.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(provisionParams))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForCompletion()

		By("checking that the parameter values are in a binding")
		checkBindingOutput(`"bind_output":"foo;bar"`)

		By("updating a parameter")
		const updateOneParams = `{"beta_input":"baz"}`
		updateOneResponse := brokerClient.Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateOneParams))
		Expect(updateOneResponse.Error).NotTo(HaveOccurred())
		Expect(updateOneResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForCompletion()

		By("checking the value is updated in a binding")
		checkBindingOutput(`"bind_output":"foo;baz"`)

		By("updating another parameter")
		const updateTwoParams = `{"alpha_input":"quz"}`
		updateTwoResponse := brokerClient.Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateTwoParams))
		Expect(updateTwoResponse.Error).NotTo(HaveOccurred())
		Expect(updateTwoResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForCompletion()

		By("checking that both parameters remain updated in a binding")
		checkBindingOutput(`"bind_output":"quz;baz"`)
	})
})
