package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("Database Encryption", func() {
	const (
		provisionParams           = `{"foo":"bar"}`
		bindParams                = `{"baz":"quz"}`
		updateParams              = `{"update_output": "update output value"}`
		provisionOutput           = `{"provision_output":"provision output value"}`
		provisionOutputStateValue = `value = \"provision output value\"`
		updateOutput              = `{"provision_output":"provision output value","update_output_output":"update output value"}`
		updateOutputStateValue    = `value = \"${var.update_output}\"`
		bindOutput                = `{"bind_output":"provision output value and bind output value"}`
		bindOutputStateValue      = `value = \"${var.provision_output} and bind output value\"`
		tfStateKey                = `"tfstate":`
		serviceOfferingGUID       = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID           = "8b52a460-b246-11eb-a8f5-d349948e2480"
		serviceInstanceFKQuery    = "service_instance_id = ?"
		serviceInstanceIdQuery    = "id = ?"
		tfWorkspaceIdQuery        = "id = ?"
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
	)

	findRecord := func(dest interface{}, query, guid string) {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		result := db.Where(query, guid).First(dest)
		ExpectWithOffset(1, result.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, result.RowsAffected).To(Equal(int64(1)))
	}

	persistedRequestDetails := func(serviceInstanceGUID string) string {
		record := models.ProvisionRequestDetails{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.RequestDetails
	}

	persistedServiceInstanceDetails := func(serviceInstanceGUID string) string {
		record := models.ServiceInstanceDetails{}
		findRecord(&record, serviceInstanceIdQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceInstanceTerraformWorkspace := func(serviceInstanceGUID string) string {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:", serviceInstanceGUID))
		return record.Workspace
	}

	persistedServiceBindingDetails := func(serviceInstanceGUID string) string {
		record := models.ServiceBindingCredentials{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceBindingTerraformWorkspace := func(serviceInstanceGUID, serviceBindingGUID string) string {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		return record.Workspace
	}

	expectServiceBindingDetailsToNotExist := func(serviceInstanceGUID string) {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.ServiceBindingCredentials{}).Where(serviceInstanceFKQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectServiceInstanceDetailsToNotExist := func(serviceInstanceGUID string) {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.ServiceInstanceDetails{}).Where(serviceInstanceIdQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	createBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		bindResponse := brokerClient.Bind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(bindParams))
		Expect(bindResponse.Error).NotTo(HaveOccurred())
	}

	deleteBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		unbindResponse := brokerClient.Unbind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(unbindResponse.Error).NotTo(HaveOccurred())
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
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		waitForAsyncRequest(serviceInstanceGUID)
	}

	deprovisionServiceInstance := func(serviceInstanceGUID string) {
		deprovisionResponse := brokerClient.Deprovision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(deprovisionResponse.Error).NotTo(HaveOccurred())
		Expect(deprovisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
	}

	startBroker := func(encryptionEnabled bool, encryptionPasswords string) *Session {
		runBrokerCommand := exec.Command(csb, "serve")
		os.Unsetenv("CH_CRED_HUB_URL")
		runBrokerCommand.Env = append(
			os.Environ(),
			"CSB_LISTENER_HOST=localhost",
			"DB_TYPE=sqlite3",
			fmt.Sprintf("EXPERIMENTAL_ENCRYPTION_ENABLED=%t", encryptionEnabled),
			fmt.Sprintf("EXPERIMENTAL_ENCRYPTION_PASSWORDS=%s", encryptionPasswords),
			fmt.Sprintf("DB_PATH=%s", databaseFile),
			fmt.Sprintf("PORT=%d", brokerPort),
			fmt.Sprintf("SECURITY_USER_NAME=%s", brokerUsername),
			fmt.Sprintf("SECURITY_USER_PASSWORD=%s", brokerPassword),
		)
		session, err := Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		waitForBrokerToStart(brokerPort)
		return session
	}

	bePlaintextProvisionParams := Equal(provisionParams)
	bePlaintextProvisionOutput := Equal(provisionOutput)
	bePlaintextInstanceTerraformState := SatisfyAll(
		ContainSubstring(provisionOutputStateValue),
		ContainSubstring(tfStateKey),
	)
	haveAnyPlaintextServiceTerraformState := SatisfyAny(
		ContainSubstring(provisionOutputStateValue),
		ContainSubstring(tfStateKey),
	)
	bePlaintextBindingDetails := Equal(bindOutput)
	bePlaintextBindingTerraformState := SatisfyAll(
		ContainSubstring(bindOutputStateValue),
		ContainSubstring(tfStateKey),
	)
	haveAnyPlaintextBindingTerraformState := SatisfyAny(
		ContainSubstring(bindOutputStateValue),
		ContainSubstring(tfStateKey),
	)

	BeforeEach(func() {
		var err error
		originalDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		fixturesDir = path.Join(originalDir, "fixtures", "brokerpak-with-fake-provider")

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

		brokerClient, err = client.New(brokerUsername, brokerPassword, "localhost", brokerPort)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("no encryption key is configured", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
			brokerSession       *Session
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			brokerSession = startBroker(false, "")
			provisionServiceInstance(serviceInstanceGUID)
		})

		AfterEach(func() {
			brokerSession.Terminate()
		})

		It("stores sensitive fields in plaintext", func() {
			By("checking the provision fields")
			Expect(persistedRequestDetails(serviceInstanceGUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(bePlaintextInstanceTerraformState)

			By("checking the binding fields")
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(bePlaintextBindingTerraformState)

			By("checking how update persists service instance fields")
			updateServiceInstance(serviceInstanceGUID)
			Expect(persistedRequestDetails(serviceInstanceGUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(Equal(updateOutput))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)
			expectServiceBindingDetailsToNotExist(serviceInstanceGUID)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(bePlaintextBindingTerraformState)

			By("checking the service instance fields after deprovision")
			deprovisionServiceInstance(serviceInstanceGUID)
			expectServiceInstanceDetailsToNotExist(serviceInstanceGUID)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))
		})
	})

	When("the encryption key is configured", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
			brokerSession       *Session
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			brokerSession = startBroker(true, encryptionPasswords)
			provisionServiceInstance(serviceInstanceGUID)
		})

		AfterEach(func() {
			brokerSession.Terminate()
		})

		It("encrypts sensitive fields", func() {
			By("checking the provision fields")
			Expect(persistedRequestDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(haveAnyPlaintextServiceTerraformState)

			By("checking the binding fields")
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("checking how update persists service instance fields")
			updateServiceInstance(serviceInstanceGUID)
			Expect(persistedRequestDetails(serviceInstanceGUID)).NotTo(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(Equal(updateOutput))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)
			expectServiceBindingDetailsToNotExist(serviceInstanceGUID)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(SatisfyAny(
				ContainSubstring(bindOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("ckecking the service instance fields after deprovision")
			deprovisionServiceInstance(serviceInstanceGUID)
			expectServiceInstanceDetailsToNotExist(serviceInstanceGUID)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))
		})
	})

	When("when encryption is turned on", func() {
		It("it encrypts the database", func() {
			By("starting the broker without a password")
			brokerSession := startBroker(false, "")
			Expect(string(brokerSession.Out.Contents())).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(brokerSession).To(Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`))

			By("creating a service instance and checking fields are stored in plain text")
			serviceInstanceGUID := uuid.New()
			provisionServiceInstance(serviceInstanceGUID)
			Expect(persistedRequestDetails(serviceInstanceGUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(bePlaintextInstanceTerraformState)

			By("creating a binding and checking the fields are stored in plain text")
			serviceBindingGUID := uuid.New()
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(bePlaintextBindingTerraformState)

			By("restarting the broker with a password")
			brokerSession.Terminate()
			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			brokerSession = startBroker(true, encryptionPasswords)
			Expect(brokerSession.Out).To(SatisfyAll(
				Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-password","previous-primary":"none"}}`),
				Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`),
			))

			By("checking that the previous data is now encrypted")
			Expect(persistedRequestDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(haveAnyPlaintextServiceTerraformState)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("restarting the broker with the same password")
			brokerSession.Terminate()
			brokerSession = startBroker(true, encryptionPasswords)
			Expect(string(brokerSession.Out.Contents())).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(brokerSession.Out).To(Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`))

			By("unbinding")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			By("deprovisioning")
			deprovisionServiceInstance(serviceInstanceGUID)

			brokerSession.Terminate()
		})
	})
})
