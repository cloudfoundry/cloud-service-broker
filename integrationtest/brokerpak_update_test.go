package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/providers/tf/wrapper"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("Brokerpak Update", func() {
	const (
		provisionParams             = `{"provision_input":"bar"}`
		bindParams                  = `{"bind_input":"quz"}`
		updateParams                = `{"update_input": "update output value"}`
		updateOutputStateKey        = `"update_output"`
		updatedUpdateOutputStateKey = `"update_output_updated"`
		updatedOutputHCL            = `output update_output_updated { value = "${var.update_input}" }`
		initialOutputHCL            = `output update_output { value = "${var.update_input}" }`
		initialBindingOutputHCL     = `bind_output { value = "${var.provision_output} and bind output value" }`
		updatedBindingOutputHCL     = `bind_output_updated { value = "${var.provision_output} and bind output value" }`
		bindingOutputStateKey       = `"bind_output"`

		serviceOfferingGUID = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID     = "8b52a460-b246-11eb-a8f5-d349948e2480"
		tfWorkspaceIdQuery  = "id = ?"
		initialBrokerpak    = "brokerpak-with-fake-provider"
		updatedBrokerpak    = "brokerpak-with-fake-provider-updated"
	)

	var (
		originalDir    string
		fixturesDir    string
		workDir        string
		brokerPort     int
		brokerUsername string
		brokerPassword string
		brokerClient   *client.Client
		databaseFile   string
		brokerSession  *Session
	)

	findRecord := func(dest interface{}, query, guid string) {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		result := db.Where(query, guid).First(dest)
		ExpectWithOffset(3, result.Error).NotTo(HaveOccurred())
		ExpectWithOffset(3, result.RowsAffected).To(Equal(int64(1)))
	}

	persistedTerraformWorkspace := func(serviceInstanceGUID, serviceBindingGUID string) *wrapper.TerraformWorkspace {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		ws, _ := wrapper.DeserializeWorkspace(string(record.Workspace))
		return ws
	}

	createBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		bindResponse := brokerClient.Bind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(bindParams))
		Expect(bindResponse.Error).NotTo(HaveOccurred())
		Expect(bindResponse.StatusCode).To(Equal(http.StatusCreated))
	}

	deleteBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		unbindResponse := brokerClient.Unbind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(unbindResponse.Error).NotTo(HaveOccurred())
		Expect(unbindResponse.StatusCode).To(Equal(http.StatusOK))
	}

	waitForAsyncRequest := func(serviceInstanceGUID string) {
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

	updateServiceInstance := func(serviceInstanceGUID string) {
		updateResponse := brokerClient.Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateParams))
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
	}

	provisionServiceInstance := func(serviceInstanceGUID string) {
		provisionResponse := brokerClient.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(provisionParams))
		ExpectWithOffset(1, provisionResponse.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))

		waitForAsyncRequest(serviceInstanceGUID)
	}

	deprovisionServiceInstance := func(serviceInstanceGUID string) {
		deprovisionResponse := brokerClient.Deprovision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(deprovisionResponse.Error).NotTo(HaveOccurred())
		Expect(deprovisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
	}

	beInitialTerraformHCL := SatisfyAll(
		ContainSubstring(initialOutputHCL),
		Not(ContainSubstring(updatedOutputHCL)),
	)

	beInitialBindingTerraformHCL := SatisfyAll(
		ContainSubstring(initialBindingOutputHCL),
		Not(ContainSubstring(updatedBindingOutputHCL)),
	)

	haveInitialOutputs := SatisfyAll(
		ContainSubstring(updateOutputStateKey),
		Not(ContainSubstring(updatedUpdateOutputStateKey)),
	)

	haveInitialBindingOutputs := ContainSubstring(bindingOutputStateKey)

	haveEmtpyState := SatisfyAll(
		ContainSubstring(`"outputs": {}`),
		ContainSubstring(`"resources": []`),
	)

	beUpdatedTerraformHCL := SatisfyAll(
		ContainSubstring(updatedOutputHCL),
		Not(ContainSubstring(initialOutputHCL)),
	)

	beUpdatedBindingTerraformHCL := SatisfyAll(
		ContainSubstring(updatedBindingOutputHCL),
		Not(ContainSubstring(initialBindingOutputHCL)),
	)

	haveUpdatedOutputs := SatisfyAll(
		ContainSubstring(updatedUpdateOutputStateKey),
		Not(ContainSubstring(updateOutputStateKey)),
	)

	startBrokerSession := func(updatesEnabled bool) *Session {
		runBrokerCommand := exec.Command(csb, "serve")
		os.Unsetenv("CH_CRED_HUB_URL")
		runBrokerCommand.Env = append(
			os.Environ(),
			"CSB_LISTENER_HOST=localhost",
			"DB_TYPE=sqlite3",
			fmt.Sprintf("BROKERPAK_UPDATES_ENABLED=%t", updatesEnabled),
			fmt.Sprintf("DB_PATH=%s", databaseFile),
			fmt.Sprintf("PORT=%d", brokerPort),
			fmt.Sprintf("SECURITY_USER_NAME=%s", brokerUsername),
			fmt.Sprintf("SECURITY_USER_PASSWORD=%s", brokerPassword),
		)
		session, err := Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

	startBroker := func(updatesEnabled bool) *Session {
		session := startBrokerSession(updatesEnabled)
		waitForBrokerToStart(brokerPort)
		return session
	}

	persistedTerraformModuleDefinition := func(serviceInstanceGUID, serviceBindingGUID string) string {
		return persistedTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID).Modules[0].Definitions["main"]
	}

	persistedTerraformState := func(serviceInstanceGUID, serviceBindingGUID string) string {
		return string(persistedTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID).State)
	}

	buildBrokerpakFor := func(fixtureName string) {
		var err error

		if workDir == "" {
			originalDir, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			workDir, err = os.MkdirTemp("", "*-csb-test")
			Expect(err).NotTo(HaveOccurred())
		}
		err = os.Chdir(workDir)
		Expect(err).NotTo(HaveOccurred())

		fixturesDir = path.Join(originalDir, "fixtures", fixtureName)
		buildBrokerpakCommand := exec.Command(csb, "pak", "build", fixturesDir)
		session, err := Start(buildBrokerpakCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		EventuallyWithOffset(1, session, 10*time.Minute).Should(Exit(0))
	}

	pushUpdatedBrokerpak := func(updatesEnabled bool) {
		brokerSession.Terminate()
		buildBrokerpakFor(updatedBrokerpak)
		brokerSession = startBroker(updatesEnabled)
	}

	BeforeEach(func() {
		var err error

		buildBrokerpakFor(initialBrokerpak)

		brokerUsername = uuid.New()
		brokerPassword = uuid.New()
		brokerPort = freePort()
		databaseFile = path.Join(workDir, "databaseFile.dat")

		brokerClient, err = client.New(brokerUsername, brokerPassword, "localhost", brokerPort)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
		workDir = ""
	})

	When("brokerpak updates are disabled", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			brokerSession = startBroker(false)
			provisionServiceInstance(serviceInstanceGUID)
		})

		AfterEach(func() {
			brokerSession.Terminate()
		})

		It("uses the old HCL for service instances operations", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(false)

			updateServiceInstance(serviceInstanceGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveInitialOutputs)

			deprovisionServiceInstance(serviceInstanceGUID)
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveEmtpyState)
		})

		It("uses the initial HCL for binding operations", func() {
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(false)

			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveEmtpyState)
		})
	})

	When("brokerpak updates are enabled", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			brokerSession = startBroker(true)
			provisionServiceInstance(serviceInstanceGUID)
		})

		AfterEach(func() {
			brokerSession.Terminate()
		})

		It("uses the updated HCL for an update", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			updateServiceInstance(serviceInstanceGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveUpdatedOutputs)
		})

		It("uses the updated HCL for a deprovision", func() {
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beInitialTerraformHCL)

			pushUpdatedBrokerpak(true)

			deprovisionServiceInstance(serviceInstanceGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, "")).To(beUpdatedTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, "")).To(haveEmtpyState)
		})

		It("uses the updated HCL for unbind", func() {
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beInitialBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveInitialBindingOutputs)

			pushUpdatedBrokerpak(true)

			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			Expect(persistedTerraformModuleDefinition(serviceInstanceGUID, serviceBindingGUID)).To(beUpdatedBindingTerraformHCL)
			Expect(persistedTerraformState(serviceInstanceGUID, serviceBindingGUID)).To(haveEmtpyState)
		})
	})
})
