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
		originalDir         string
		fixturesDir         string
		workDir             string
		brokerPort          int
		brokerUsername      string
		brokerPassword      string
		brokerSession       *Session
		brokerClient        *client.Client
		databaseFile        string
		encryptionPasswords string
		encryptionEnabled   bool
		serviceInstanceGUID string
		serviceBindingGUID  string
	)

	findRecord := func(dest interface{}, query, guid string) {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		err = db.Where(query, guid).First(dest).Error
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}

	persistedRequestDetails := func() string {
		record := models.ProvisionRequestDetails{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.RequestDetails
	}

	persistedServiceInstanceDetails := func() string {
		record := models.ServiceInstanceDetails{}
		findRecord(&record, serviceInstanceIdQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceInstanceTerraformWorkspace := func() string {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:", serviceInstanceGUID))
		return record.Workspace
	}

	persistedServiceBindingDetails := func() string {
		record := models.ServiceBindingCredentials{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceBindingTerraformWorkspace := func() string {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		return record.Workspace
	}

	expectServiceBindingDetailsToNotExist := func() {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.ServiceBindingCredentials{}).Where(serviceInstanceFKQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectServiceInstanceDetailsToNotExist := func() {
		db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.ServiceInstanceDetails{}).Where(serviceInstanceIdQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	createBinding := func() {
		serviceBindingGUID = uuid.New()
		bindResponse := brokerClient.Bind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(bindParams))
		Expect(bindResponse.Error).NotTo(HaveOccurred())
	}

	deleteBinding := func() {
		unbindResponse := brokerClient.Unbind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(unbindResponse.Error).NotTo(HaveOccurred())
	}

	waitForAsyncRequest := func() {
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

	updateServiceInstance := func() {
		updateResponse := brokerClient.Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateParams))
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest()
	}

	deprovisionServiceInstance := func() {
		deprovisionResponse := brokerClient.Deprovision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(deprovisionResponse.Error).NotTo(HaveOccurred())
		Expect(deprovisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest()
	}

	JustBeforeEach(func() {
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
		brokerSession, err = Start(runBrokerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		waitForBrokerToStart(brokerPort)

		brokerClient, err = client.New(brokerUsername, brokerPassword, "localhost", brokerPort)
		Expect(err).NotTo(HaveOccurred())

		serviceInstanceGUID = uuid.New()
		provisionResponse := brokerClient.Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(provisionParams))
		Expect(provisionResponse.Error).NotTo(HaveOccurred())
		Expect(provisionResponse.StatusCode).To(Equal(http.StatusAccepted))

		waitForAsyncRequest()
	})

	AfterEach(func() {
		brokerSession.Terminate()

		err := os.Chdir(originalDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(workDir)
		Expect(err).NotTo(HaveOccurred())
	})

	When("no encryption key is configured", func() {
		BeforeEach(func() {
			encryptionEnabled = false
			encryptionPasswords = ""
		})

		It("stores sensitive fields in plaintext", func() {
			By("checking the provision fields")
			Expect(persistedRequestDetails()).To(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails()).To(Equal(provisionOutput))
			Expect(persistedServiceInstanceTerraformWorkspace()).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields")
			createBinding()
			Expect(persistedServiceBindingDetails()).To(Equal(bindOutput))
			Expect(persistedServiceBindingTerraformWorkspace()).To(SatisfyAll(
				ContainSubstring(bindOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking how update persists service instance fields")
			updateServiceInstance()
			Expect(persistedRequestDetails()).To(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails()).To(Equal(updateOutput))
			Expect(persistedServiceInstanceTerraformWorkspace()).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			deleteBinding()
			expectServiceBindingDetailsToNotExist()
			Expect(persistedServiceBindingTerraformWorkspace()).To(SatisfyAll(
				ContainSubstring(bindOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("ckecking the service instance fields after deprovision", func() {
				deprovisionServiceInstance()
				expectServiceInstanceDetailsToNotExist()
				Expect(persistedServiceInstanceTerraformWorkspace()).To(SatisfyAll(
					ContainSubstring(provisionOutputStateValue),
					ContainSubstring(updateOutputStateValue),
					ContainSubstring(tfStateKey),
				))
			})
		})
	})

	When("the encryption key is configured", func() {
		BeforeEach(func() {
			encryptionEnabled = true
			encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
		})

		It("encrypts sensitive fields", func() {
			By("checking the provision fields")
			Expect(persistedRequestDetails()).NotTo(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails()).NotTo(Equal(provisionOutput))
			Expect(persistedServiceInstanceTerraformWorkspace()).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields")
			createBinding()
			Expect(persistedServiceBindingDetails()).NotTo(Equal(bindOutput))
			Expect(persistedServiceBindingTerraformWorkspace()).NotTo(SatisfyAny(
				ContainSubstring(bindOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking how update persists service instance fields")
			updateServiceInstance()
			Expect(persistedRequestDetails()).NotTo(Equal(provisionParams))
			Expect(persistedServiceInstanceDetails()).NotTo(Equal(updateOutput))
			Expect(persistedServiceInstanceTerraformWorkspace()).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			deleteBinding()
			expectServiceBindingDetailsToNotExist()
			Expect(persistedServiceBindingTerraformWorkspace()).NotTo(SatisfyAny(
				ContainSubstring(bindOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("ckecking the service instance fields after deprovision", func() {
				deprovisionServiceInstance()
				expectServiceInstanceDetailsToNotExist()
				Expect(persistedServiceInstanceTerraformWorkspace()).NotTo(SatisfyAny(
					ContainSubstring(provisionOutputStateValue),
					ContainSubstring(updateOutputStateValue),
					ContainSubstring(tfStateKey),
				))
			})
		})
	})
})
