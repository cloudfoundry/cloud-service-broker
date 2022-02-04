package integrationtest_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/integrationtest/helper"
	. "github.com/onsi/ginkgo/v2"
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
		provisionParams           = `{"provision_input":"bar"}`
		bindParams                = `{"bind_input":"quz"}`
		updateParams              = `{"update_input": "update output value"}`
		mergedParams              = `{"provision_input":"bar","update_input":"update output value"}`
		provisionOutput           = `{"provision_output":"provision output value"}`
		provisionOutputStateValue = `value = \"provision output value\"`
		updateOutput              = `{"provision_output":"provision output value","update_output":"update output value"}`
		updateOutputStateValue    = `value = \"${var.update_input}\"`
		bindOutput                = `{"bind_output":"provision output value and bind output value"}`
		bindOutputStateValue      = `value = \"${var.provision_output} and bind output value\"`
		tfStateKey                = `"tfstate":`
		serviceOfferingGUID       = "76c5725c-b246-11eb-871f-ffc97563fbd0"
		servicePlanGUID           = "8b52a460-b246-11eb-a8f5-d349948e2480"
		serviceInstanceFKQuery    = "service_instance_id = ?"
		serviceBindingIdQuery     = "service_binding_id = ?"
		serviceInstanceIdQuery    = "id = ?"
		tfWorkspaceIdQuery        = "id = ?"
		passwordMetadataQuery     = "label = ?"
	)

	var (
		originalDir helper.Original
		testLab     *helper.TestLab
		session     *Session
	)

	findRecord := func(dest interface{}, query, guid string) {
		db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		result := db.Where(query, guid).First(dest)
		ExpectWithOffset(1, result.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, result.RowsAffected).To(Equal(int64(1)))
	}

	persistedProvisionRequestDetails := func(serviceInstanceGUID string) []byte {
		record := models.ProvisionRequestDetails{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.RequestDetails
	}

	persistedServiceInstanceDetails := func(serviceInstanceGUID string) []byte {
		record := models.ServiceInstanceDetails{}
		findRecord(&record, serviceInstanceIdQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceInstanceTerraformWorkspace := func(serviceInstanceGUID string) []byte {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:", serviceInstanceGUID))
		return record.Workspace
	}

	persistedBindRequestDetails := func(serviceBindingGUID string) []byte {
		record := models.BindRequestDetails{}
		findRecord(&record, serviceBindingIdQuery, serviceBindingGUID)
		return record.RequestDetails
	}

	persistedServiceBindingDetails := func(serviceInstanceGUID string) []byte {
		record := models.ServiceBindingCredentials{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceBindingTerraformWorkspace := func(serviceInstanceGUID, serviceBindingGUID string) []byte {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIdQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		return record.Workspace
	}

	persistedPasswordMetadata := func(label string) models.PasswordMetadata {
		record := models.PasswordMetadata{}
		findRecord(&record, passwordMetadataQuery, label)
		return record
	}

	expectNoPasswordMetadataToExist := func() {
		db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.PasswordMetadata{}).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectPasswordMetadataToNotExist := func(label string) {
		db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.PasswordMetadata{}).Where(passwordMetadataQuery, label).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectServiceBindingDetailsToNotExist := func(serviceInstanceGUID string) {
		db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.ServiceBindingCredentials{}).Where(serviceInstanceFKQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectBindRequestDetailsToNotExist := func(serviceBindingGUID string) {
		db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.BindRequestDetails{}).Where(serviceBindingIdQuery, serviceBindingGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectServiceInstanceDetailsToNotExist := func(serviceInstanceGUID string) {
		db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		var count int64
		Expect(db.Model(&models.ServiceInstanceDetails{}).Where(serviceInstanceIdQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	createBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		bindResponse := testLab.Client().Bind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(bindParams))
		Expect(bindResponse.Error).NotTo(HaveOccurred())
		Expect(bindResponse.StatusCode).To(Equal(http.StatusCreated))
	}

	deleteBinding := func(serviceInstanceGUID, serviceBindingGUID string) {
		unbindResponse := testLab.Client().Unbind(serviceInstanceGUID, serviceBindingGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(unbindResponse.Error).NotTo(HaveOccurred())
		Expect(unbindResponse.StatusCode).To(Equal(http.StatusOK))
	}

	waitForAsyncRequest := func(serviceInstanceGUID string) {
		Eventually(func() bool {
			lastOperationResponse := testLab.Client().LastOperation(serviceInstanceGUID, requestID())
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
		updateResponse := testLab.Client().Update(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID(), []byte(updateParams))
		Expect(updateResponse.Error).NotTo(HaveOccurred())
		Expect(updateResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
	}

	provisionServiceInstance := func(serviceInstanceGUID string) {
		provisionResponse := testLab.Client().Provision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, uuid.New(), []byte(provisionParams))
		ExpectWithOffset(1, provisionResponse.Error).NotTo(HaveOccurred())
		ExpectWithOffset(1, provisionResponse.StatusCode).To(Equal(http.StatusAccepted), string(provisionResponse.ResponseBody))

		waitForAsyncRequest(serviceInstanceGUID)
	}

	deprovisionServiceInstance := func(serviceInstanceGUID string) {
		deprovisionResponse := testLab.Client().Deprovision(serviceInstanceGUID, serviceOfferingGUID, servicePlanGUID, requestID())
		Expect(deprovisionResponse.Error).NotTo(HaveOccurred())
		Expect(deprovisionResponse.StatusCode).To(Equal(http.StatusAccepted))
		waitForAsyncRequest(serviceInstanceGUID)
	}

	startBrokerSession := func(encryptionEnabled bool, encryptionPasswords string) *Session {
		return testLab.StartBrokerSession(
			fmt.Sprintf("ENCRYPTION_ENABLED=%t", encryptionEnabled),
			fmt.Sprintf("ENCRYPTION_PASSWORDS=%s", encryptionPasswords),
		)
	}

	startBroker := func(encryptionEnabled bool, encryptionPasswords string) *Session {
		return testLab.StartBroker(
			fmt.Sprintf("ENCRYPTION_ENABLED=%t", encryptionEnabled),
			fmt.Sprintf("ENCRYPTION_PASSWORDS=%s", encryptionPasswords),
		)
	}

	bePlaintextProvisionParams := Equal([]byte(provisionParams))
	bePlaintextMergedParams := Equal([]byte(mergedParams))
	bePlaintextProvisionOutput := Equal([]byte(provisionOutput))
	bePlaintextInstanceTerraformState := SatisfyAll(
		ContainSubstring(provisionOutputStateValue),
		ContainSubstring(tfStateKey),
	)
	haveAnyPlaintextServiceTerraformState := SatisfyAny(
		ContainSubstring(provisionOutputStateValue),
		ContainSubstring(tfStateKey),
	)
	bePlaintextBindParams := Equal([]byte(bindParams))
	bePlaintextBindingDetails := Equal([]byte(bindOutput))
	bePlaintextBindingTerraformState := SatisfyAll(
		ContainSubstring(bindOutputStateValue),
		ContainSubstring(tfStateKey),
	)
	haveAnyPlaintextBindingTerraformState := SatisfyAny(
		ContainSubstring(bindOutputStateValue),
		ContainSubstring(tfStateKey),
	)

	BeforeEach(func() {
		originalDir = helper.OriginalDir()
		testLab = helper.NewTestLab(csb)
		testLab.BuildBrokerpak(string(originalDir), "fixtures", "brokerpak-with-fake-provider")
	})

	AfterEach(func() {
		session.Terminate()
		originalDir.Return()
	})

	When("encryption is turned off", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			session = startBroker(false, "")
			provisionServiceInstance(serviceInstanceGUID)
		})

		It("stores sensitive fields in plaintext", func() {
			By("checking the provision fields")
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(bePlaintextInstanceTerraformState)

			By("checking the binding fields")
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).To(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(haveAnyPlaintextBindingTerraformState)

			By("checking how update persists service instance fields")
			updateServiceInstance(serviceInstanceGUID)
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).To(bePlaintextMergedParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(Equal([]byte(updateOutput)))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)
			expectServiceBindingDetailsToNotExist(serviceInstanceGUID)
			expectBindRequestDetailsToNotExist(serviceBindingGUID)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(haveAnyPlaintextBindingTerraformState)

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

	When("encryption is turned on", func() {
		var (
			serviceInstanceGUID string
			serviceBindingGUID  string
		)

		BeforeEach(func() {
			serviceInstanceGUID = uuid.New()
			serviceBindingGUID = uuid.New()

			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			session = startBroker(true, encryptionPasswords)
			provisionServiceInstance(serviceInstanceGUID)
		})

		It("encrypts sensitive fields", func() {
			By("checking the provision fields")
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(haveAnyPlaintextServiceTerraformState)

			By("checking the binding fields")
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("checking how update persists service instance fields")
			updateServiceInstance(serviceInstanceGUID)
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).NotTo(Equal(mergedParams))
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(Equal(updateOutput))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)
			expectServiceBindingDetailsToNotExist(serviceInstanceGUID)
			expectBindRequestDetailsToNotExist(serviceBindingGUID)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

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

	When("encryption is turned on after it was previously off", func() {
		It("it encrypts the database", func() {
			By("starting the broker without a password")
			session = startBroker(false, "")
			Expect(string(session.Out.Contents())).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(session).To(Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`))

			By("creating a service instance and checking fields are stored in plain text")
			serviceInstanceGUID := uuid.New()
			provisionServiceInstance(serviceInstanceGUID)
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(bePlaintextInstanceTerraformState)

			By("creating a binding and checking the fields are stored in plain text")
			serviceBindingGUID := uuid.New()
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).To(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(bePlaintextBindingTerraformState)

			By("restarting the broker with a password")
			session.Terminate()
			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			session = startBroker(true, encryptionPasswords)
			Expect(session.Out).To(SatisfyAll(
				Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-password","previous-primary":"none"}}`),
				Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`),
			))

			By("checking that the password metadata are still stored")
			Expect(persistedPasswordMetadata("my-password").Primary).To(BeTrue())

			By("checking that the previous data is now encrypted")
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(haveAnyPlaintextServiceTerraformState)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("restarting the broker with the same password")
			session.Terminate()
			session = startBroker(true, encryptionPasswords)
			Expect(string(session.Out.Contents())).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(session.Out).To(Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`))

			By("checking that the password metadata are still stored")
			Expect(persistedPasswordMetadata("my-password").Primary).To(BeTrue())

			By("unbinding")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			By("deprovisioning")
			deprovisionServiceInstance(serviceInstanceGUID)
		})
	})

	When("encryption is turned off after it was previously on", func() {
		It("decrypts the database", func() {
			By("starting the broker with a password")
			encryptionPasswords := `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			session = startBroker(true, encryptionPasswords)
			Expect(session.Out).To(SatisfyAll(
				Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-password","previous-primary":"none"}}`),
				Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`),
			))

			By("creating a service instance and checking fields are stored encrypted")
			serviceInstanceGUID := uuid.New()
			provisionServiceInstance(serviceInstanceGUID)
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(haveAnyPlaintextServiceTerraformState)

			By("creating a binding and checking the fields are stored encrypted")
			serviceBindingGUID := uuid.New()
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("restarting the broker with encryption turned off")
			session.Terminate()
			encryptionPasswords = `[{"primary":false,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			session = startBroker(false, encryptionPasswords)
			Expect(session.Out).To(SatisfyAll(
				Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"none","previous-primary":"my-password"}}`),
				Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`),
			))

			By("checking that the password metadata are still stored")
			Expect(persistedPasswordMetadata("my-password").Primary).To(BeFalse())

			By("checking that the previous data is now decrypted")
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).To(bePlaintextInstanceTerraformState)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).To(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).To(bePlaintextBindingTerraformState)

			By("restarting the broker with encryption turned off again")
			session.Terminate()
			session = startBroker(false, "")
			Expect(string(session.Out.Contents())).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(session.Out).To(Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`))

			By("checking that the password metadata are removed")
			expectNoPasswordMetadataToExist()

			By("unbinding")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			By("deprovisioning")
			deprovisionServiceInstance(serviceInstanceGUID)
		})
	})

	When("encryption is turned on and passwords are rotated", func() {
		It("it re-encrypts the database using new password", func() {
			By("starting the broker with a password")
			firstEncryptionPassword := `{"primary":true,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
			encryptionPasswords := fmt.Sprintf("[%s]", firstEncryptionPassword)
			session = startBroker(true, encryptionPasswords)
			Expect(session.Out).To(SatisfyAll(
				Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-first-password","previous-primary":"none"}}`),
				Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-first-password"}}`),
			))

			By("checking that the password metadata are stored")
			Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeTrue())

			By("creating a service instance and checking fields are stored encrypted")
			serviceInstanceGUID := uuid.New()
			provisionServiceInstance(serviceInstanceGUID)
			firstEncryptionPersistedRequestDetails := persistedProvisionRequestDetails(serviceInstanceGUID)
			firstEncryptionPersistedServiceInstanceDetails := persistedServiceInstanceDetails(serviceInstanceGUID)
			firstEncryptionpersistedServiceInstanceTerraformWorkspace := persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)

			By("creating a binding and checking the fields are stored in plain text")
			serviceBindingGUID := uuid.New()
			createBinding(serviceInstanceGUID, serviceBindingGUID)
			firstEncryptionPersistedServiceBindingParams := persistedBindRequestDetails(serviceBindingGUID)
			firstEncryptionPersistedServiceBindingDetails := persistedServiceBindingDetails(serviceInstanceGUID)
			firstEncryptionPersistedServiceBindingTerraformWorkspace := persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)

			By("restarting the broker with a different primary password")
			session.Terminate()
			firstEncryptionPassword = `{"primary":false,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
			const secondEncryptionPassword = `{"primary":true,"label":"my-second-password","password":{"secret":"verysecretcoolpassword"}}`
			encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
			session = startBroker(true, encryptionPasswords)
			Expect(session.Out).To(SatisfyAll(
				Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`),
				Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`),
			))

			By("checking that the password metadata are stored")
			Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeFalse())
			Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeTrue())

			By("checking that the previous data is still encrypted")
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(haveAnyPlaintextServiceTerraformState)
			Expect(persistedBindRequestDetails(serviceBindingGUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("checking that the previous data is encrypted differently")
			Expect(persistedProvisionRequestDetails(serviceInstanceGUID)).NotTo(Equal(firstEncryptionPersistedRequestDetails))
			Expect(persistedServiceInstanceDetails(serviceInstanceGUID)).NotTo(Equal(firstEncryptionPersistedServiceInstanceDetails))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstanceGUID)).NotTo(Equal(firstEncryptionpersistedServiceInstanceTerraformWorkspace))
			Expect(persistedBindRequestDetails(serviceBindingGUID)).NotTo(Equal(firstEncryptionPersistedServiceBindingParams))
			Expect(persistedServiceBindingDetails(serviceInstanceGUID)).NotTo(Equal(firstEncryptionPersistedServiceBindingDetails))
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstanceGUID, serviceBindingGUID)).NotTo(Equal(firstEncryptionPersistedServiceBindingTerraformWorkspace))

			By("restarting the broker with the new password only")
			session.Terminate()
			session = startBroker(true, fmt.Sprintf("[%s]", secondEncryptionPassword))
			Expect(string(session.Out.Contents())).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(session.Out).To(Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`))

			By("checking password metadata are cleaned up")
			Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeTrue())
			expectPasswordMetadataToNotExist("my-first-password")

			By("unbinding")
			deleteBinding(serviceInstanceGUID, serviceBindingGUID)

			By("deprovisioning")
			deprovisionServiceInstance(serviceInstanceGUID)
		})

		When("previous password is not provided", func() {
			It("fails to start up", func() {
				By("starting the broker with a password")
				firstEncryptionPassword := `{"primary":true,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				encryptionPasswords := fmt.Sprintf("[%s]", firstEncryptionPassword)
				session = startBroker(true, encryptionPasswords)
				Expect(session.Out).To(SatisfyAll(
					Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-first-password","previous-primary":"none"}}`),
					Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-first-password"}}`),
				))

				By("restarting the broker with a different primary password and without the initial password")
				session.Terminate()
				const secondEncryptionPassword = `{"primary":true,"label":"my-second-password","password":{"secret":"verysecretcoolpassword"}}`
				encryptionPasswords = fmt.Sprintf("[%s]", secondEncryptionPassword)
				session = startBrokerSession(true, encryptionPasswords)
				session.Wait(time.Minute)

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session.Err).To(Say(`the password labelled "my-first-password" must be supplied to decrypt the database`))

				By("restarting the broker with a different primary password and with the initial password")
				firstEncryptionPassword = `{"primary":false,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
				session = startBroker(true, encryptionPasswords)
				Expect(session.Out).To(SatisfyAll(
					Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`),
					Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`),
				))
			})
		})

		When("database re-encryption fails", func() {
			It("can restart re-encrypting", func() {
				By("starting the broker with a password")
				firstEncryptionPassword := `{"primary":true,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				encryptionPasswords := fmt.Sprintf("[%s]", firstEncryptionPassword)
				session = startBroker(true, encryptionPasswords)
				Expect(session.Out).To(SatisfyAll(
					Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-first-password","previous-primary":"none"}}`),
					Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-first-password"}}`),
				))

				By("creating a service instance and checking fields are stored encrypted")
				serviceInstanceGUID1 := uuid.New()
				provisionServiceInstance(serviceInstanceGUID1)

				serviceInstanceGUID2 := uuid.New()
				provisionServiceInstance(serviceInstanceGUID2)

				By("corrupting the DB")
				db, err := gorm.Open(sqlite.Open(testLab.DatabaseFile), &gorm.Config{})
				Expect(err).NotTo(HaveOccurred())
				record := models.ServiceInstanceDetails{}
				findRecord(&record, serviceInstanceIdQuery, serviceInstanceGUID2)
				recordToRecover := record
				record.OtherDetails = []byte("something-that-cannot-be-decrypted")
				Expect(db.Save(&record).Error).NotTo(HaveOccurred())

				By("restarting the broker with a different primary password")
				session.Terminate()
				firstEncryptionPassword = `{"primary":false,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				const secondEncryptionPassword = `{"primary":true,"label":"my-second-password","password":{"secret":"verysecretcoolpassword"}}`
				encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
				session = startBrokerSession(true, encryptionPasswords)
				session.Wait(time.Minute)

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session.Out).To(Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`))
				Expect(session.Err).To(Say(`Error rotating database encryption`))

				By("checking password metadata are not removed")
				Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeTrue())
				Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeFalse())

				By("fixing the corrupted value")
				Expect(db.Save(&recordToRecover).Error).NotTo(HaveOccurred())

				By("restarting the broker with same config")
				encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
				session = startBroker(true, encryptionPasswords)
				Expect(session.Out).To(SatisfyAll(
					Say(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`),
					Say(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`),
				))

				By("checking password metadata are updated")
				Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeFalse())
				Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeTrue())
			})
		})
	})
})
