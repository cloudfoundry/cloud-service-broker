package integrationtest_test

import (
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		serviceBindingIDQuery     = "service_binding_id = ?"
		serviceInstanceIDQuery    = "id = ?"
		tfWorkspaceIDQuery        = "id = ?"
		passwordMetadataQuery     = "label = ?"
	)

	var (
		brokerpak string
		broker    *testdrive.Broker
	)

	findRecord := func(dest any, query, guid string) {
		result := dbConn.Where(query, guid).First(dest)
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
		findRecord(&record, serviceInstanceIDQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceInstanceTerraformWorkspace := func(serviceInstanceGUID string) []byte {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIDQuery, fmt.Sprintf("tf:%s:", serviceInstanceGUID))
		return record.Workspace
	}

	persistedBindRequestDetails := func(serviceBindingGUID string) []byte {
		record := models.BindRequestDetails{}
		findRecord(&record, serviceBindingIDQuery, serviceBindingGUID)
		return record.RequestDetails
	}

	persistedServiceBindingDetails := func(serviceInstanceGUID string) []byte {
		record := models.ServiceBindingCredentials{}
		findRecord(&record, serviceInstanceFKQuery, serviceInstanceGUID)
		return record.OtherDetails
	}

	persistedServiceBindingTerraformWorkspace := func(serviceInstanceGUID, serviceBindingGUID string) []byte {
		record := models.TerraformDeployment{}
		findRecord(&record, tfWorkspaceIDQuery, fmt.Sprintf("tf:%s:%s", serviceInstanceGUID, serviceBindingGUID))
		return record.Workspace
	}

	persistedPasswordMetadata := func(label string) models.PasswordMetadata {
		record := models.PasswordMetadata{}
		findRecord(&record, passwordMetadataQuery, label)
		return record
	}

	expectNoPasswordMetadataToExist := func() {
		var count int64
		Expect(dbConn.Model(&models.PasswordMetadata{}).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectPasswordMetadataToNotExist := func(label string) {
		var count int64
		Expect(dbConn.Model(&models.PasswordMetadata{}).Where(passwordMetadataQuery, label).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectServiceBindingDetailsToNotExist := func(serviceInstanceGUID string) {
		var count int64
		Expect(dbConn.Model(&models.ServiceBindingCredentials{}).Where(serviceInstanceFKQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectBindRequestDetailsToNotExist := func(serviceBindingGUID string) {
		var count int64
		Expect(dbConn.Model(&models.BindRequestDetails{}).Where(serviceBindingIDQuery, serviceBindingGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	expectServiceInstanceDetailsToNotExist := func(serviceInstanceGUID string) {
		var count int64
		Expect(dbConn.Model(&models.ServiceInstanceDetails{}).Where(serviceInstanceIDQuery, serviceInstanceGUID).Count(&count).Error).NotTo(HaveOccurred())
		Expect(count).To(BeZero())
	}

	tryStartBroker := func(encryptionEnabled bool, encryptionPasswords string) error {
		_, err := testdrive.StartBroker(
			csb,
			brokerpak,
			database,
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
			testdrive.WithEnv(
				fmt.Sprintf("ENCRYPTION_ENABLED=%t", encryptionEnabled),
				fmt.Sprintf("ENCRYPTION_PASSWORDS=%s", encryptionPasswords),
			),
		)
		Expect(err).To(HaveOccurred())
		return err
	}

	startBroker := func(encryptionEnabled bool, encryptionPasswords string) {
		broker = must(testdrive.StartBroker(
			csb,
			brokerpak,
			database,
			testdrive.WithOutputs(GinkgoWriter, GinkgoWriter),
			testdrive.WithEnv(
				fmt.Sprintf("ENCRYPTION_ENABLED=%t", encryptionEnabled),
				fmt.Sprintf("ENCRYPTION_PASSWORDS=%s", encryptionPasswords),
			),
		))
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
		brokerpak = must(testdrive.BuildBrokerpak(csb, fixtures("database-encryption")))

		DeferCleanup(func() {
			Expect(broker.Stop()).To(Succeed())
			cleanup(brokerpak)
		})
	})

	When("encryption is turned off", func() {
		BeforeEach(func() {
			startBroker(false, "")
		})

		It("stores sensitive fields in plaintext", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))

			By("checking the provision fields")
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).To(bePlaintextInstanceTerraformState)

			By("checking the binding fields")
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).To(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).To(haveAnyPlaintextBindingTerraformState)

			By("checking how update persists service instance fields")
			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(updateParams))).To(Succeed())
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).To(bePlaintextMergedParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).To(Equal([]byte(updateOutput)))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			broker.DeleteBinding(serviceInstance, serviceBinding.GUID)
			expectServiceBindingDetailsToNotExist(serviceInstance.GUID)
			expectBindRequestDetailsToNotExist(serviceBinding.GUID)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).To(haveAnyPlaintextBindingTerraformState)

			By("checking the service instance fields after deprovision")
			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
			expectServiceInstanceDetailsToNotExist(serviceInstance.GUID)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).To(SatisfyAll(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))
		})
	})

	When("encryption is turned on", func() {
		BeforeEach(func() {
			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			startBroker(true, encryptionPasswords)
		})

		It("encrypts sensitive fields", func() {
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			By("checking the provision fields")
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(haveAnyPlaintextServiceTerraformState)

			By("checking the binding fields")
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("checking how update persists service instance fields")
			Expect(broker.UpdateService(serviceInstance, testdrive.WithUpdateParams(updateParams))).To(Succeed())
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).NotTo(Equal(mergedParams))
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).NotTo(Equal(updateOutput))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))

			By("checking the binding fields after unbind")
			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())
			expectServiceBindingDetailsToNotExist(serviceInstance.GUID)
			expectBindRequestDetailsToNotExist(serviceBinding.GUID)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("ckecking the service instance fields after deprovision")
			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
			expectServiceInstanceDetailsToNotExist(serviceInstance.GUID)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(SatisfyAny(
				ContainSubstring(provisionOutputStateValue),
				ContainSubstring(updateOutputStateValue),
				ContainSubstring(tfStateKey),
			))
		})
	})

	When("encryption is turned on after it was previously off", func() {
		It("it encrypts the database", func() {
			By("starting the broker without a password")
			startBroker(false, "")
			Expect(broker.Stdout.String()).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(broker.Stdout.String()).To(MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`))

			By("creating a service instance and checking fields are stored in plain text")
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).To(bePlaintextInstanceTerraformState)

			By("creating a binding and checking the fields are stored in plain text")
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).To(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).To(bePlaintextBindingTerraformState)

			By("restarting the broker with a password")
			Expect(broker.Stop()).To(Succeed())
			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			startBroker(true, encryptionPasswords)
			Expect(broker.Stdout.String()).To(SatisfyAll(
				MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-password","previous-primary":"none"}}`),
				MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`),
			))

			By("checking that the password metadata are still stored")
			Expect(persistedPasswordMetadata("my-password").Primary).To(BeTrue())

			By("checking that the previous data is now encrypted")
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(haveAnyPlaintextServiceTerraformState)
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("restarting the broker with the same password")
			Expect(broker.Stop()).To(Succeed())
			startBroker(true, encryptionPasswords)
			Expect(broker.Stdout.String()).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(broker.Stdout.String()).To(MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`))

			By("checking that the password metadata are still stored")
			Expect(persistedPasswordMetadata("my-password").Primary).To(BeTrue())

			By("unbinding")
			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())

			By("deprovisioning")
			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
		})
	})

	When("encryption is turned off after it was previously on", func() {
		It("decrypts the database", func() {
			By("starting the broker with a password")
			encryptionPasswords := `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			startBroker(true, encryptionPasswords)
			Expect(broker.Stdout.String()).To(SatisfyAll(
				MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-password","previous-primary":"none"}}`),
				MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-password"}}`),
			))

			By("creating a service instance and checking fields are stored encrypted")
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(haveAnyPlaintextServiceTerraformState)

			By("creating a binding and checking the fields are stored encrypted")
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("restarting the broker with encryption turned off")
			Expect(broker.Stop()).To(Succeed())
			encryptionPasswords = `[{"primary":false,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			startBroker(false, encryptionPasswords)
			Expect(broker.Stdout.String()).To(SatisfyAll(
				MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"none","previous-primary":"my-password"}}`),
				MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`),
			))

			By("checking that the password metadata are still stored")
			Expect(persistedPasswordMetadata("my-password").Primary).To(BeFalse())

			By("checking that the previous data is now decrypted")
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).To(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).To(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).To(bePlaintextInstanceTerraformState)
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).To(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).To(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).To(bePlaintextBindingTerraformState)

			By("restarting the broker with encryption turned off again")
			Expect(broker.Stop()).To(Succeed())
			startBroker(false, "")
			Expect(broker.Stdout.String()).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(broker.Stdout.String()).To(MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"none"}}`))

			By("checking that the password metadata are removed")
			expectNoPasswordMetadataToExist()

			By("unbinding")
			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())

			By("deprovisioning")
			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
		})
	})

	When("encryption is turned on and passwords are rotated", func() {
		It("it re-encrypts the database using new password", func() {
			By("starting the broker with a password")
			firstEncryptionPassword := `{"primary":true,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
			encryptionPasswords := fmt.Sprintf("[%s]", firstEncryptionPassword)
			startBroker(true, encryptionPasswords)
			Expect(broker.Stdout.String()).To(SatisfyAll(
				MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-first-password","previous-primary":"none"}}`),
				MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-first-password"}}`),
			))

			By("checking that the password metadata are stored")
			Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeTrue())

			By("creating a service instance and checking fields are stored encrypted")
			serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))
			firstEncryptionPersistedRequestDetails := persistedProvisionRequestDetails(serviceInstance.GUID)
			firstEncryptionPersistedServiceInstanceDetails := persistedServiceInstanceDetails(serviceInstance.GUID)
			firstEncryptionpersistedServiceInstanceTerraformWorkspace := persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)

			By("creating a binding and checking the fields are stored in plain text")
			serviceBinding := must(broker.CreateBinding(serviceInstance, testdrive.WithBindingParams(bindParams)))
			firstEncryptionPersistedServiceBindingParams := persistedBindRequestDetails(serviceBinding.GUID)
			firstEncryptionPersistedServiceBindingDetails := persistedServiceBindingDetails(serviceInstance.GUID)
			firstEncryptionPersistedServiceBindingTerraformWorkspace := persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)

			By("restarting the broker with a different primary password")
			Expect(broker.Stop()).To(Succeed())
			firstEncryptionPassword = `{"primary":false,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
			const secondEncryptionPassword = `{"primary":true,"label":"my-second-password","password":{"secret":"verysecretcoolpassword"}}`
			encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
			startBroker(true, encryptionPasswords)
			Expect(broker.Stdout.String()).To(SatisfyAll(
				MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`),
				MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`),
			))

			By("checking that the password metadata are stored")
			Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeFalse())
			Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeTrue())

			By("checking that the previous data is still encrypted")
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionParams)
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).NotTo(bePlaintextProvisionOutput)
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(haveAnyPlaintextServiceTerraformState)
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).NotTo(bePlaintextBindParams)
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).NotTo(bePlaintextBindingDetails)
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).NotTo(haveAnyPlaintextBindingTerraformState)

			By("checking that the previous data is encrypted differently")
			Expect(persistedProvisionRequestDetails(serviceInstance.GUID)).NotTo(Equal(firstEncryptionPersistedRequestDetails))
			Expect(persistedServiceInstanceDetails(serviceInstance.GUID)).NotTo(Equal(firstEncryptionPersistedServiceInstanceDetails))
			Expect(persistedServiceInstanceTerraformWorkspace(serviceInstance.GUID)).NotTo(Equal(firstEncryptionpersistedServiceInstanceTerraformWorkspace))
			Expect(persistedBindRequestDetails(serviceBinding.GUID)).NotTo(Equal(firstEncryptionPersistedServiceBindingParams))
			Expect(persistedServiceBindingDetails(serviceInstance.GUID)).NotTo(Equal(firstEncryptionPersistedServiceBindingDetails))
			Expect(persistedServiceBindingTerraformWorkspace(serviceInstance.GUID, serviceBinding.GUID)).NotTo(Equal(firstEncryptionPersistedServiceBindingTerraformWorkspace))

			By("restarting the broker with the new password only")
			Expect(broker.Stop()).To(Succeed())
			startBroker(true, fmt.Sprintf("[%s]", secondEncryptionPassword))
			Expect(broker.Stdout.String()).NotTo(ContainSubstring(`cloud-service-broker.rotating-database-encryption`))
			Expect(broker.Stdout.String()).To(MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`))

			By("checking password metadata are cleaned up")
			Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeTrue())
			expectPasswordMetadataToNotExist("my-first-password")

			By("unbinding")
			Expect(broker.DeleteBinding(serviceInstance, serviceBinding.GUID)).To(Succeed())

			By("deprovisioning")
			Expect(broker.Deprovision(serviceInstance)).To(Succeed())
		})

		When("previous password is not provided", func() {
			It("fails to start up", func() {
				By("starting the broker with a password")
				firstEncryptionPassword := `{"primary":true,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				encryptionPasswords := fmt.Sprintf("[%s]", firstEncryptionPassword)
				startBroker(true, encryptionPasswords)
				Expect(broker.Stdout.String()).To(SatisfyAll(
					MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-first-password","previous-primary":"none"}}`),
					MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-first-password"}}`),
				))

				By("restarting the broker with a different primary password and without the initial password")
				Expect(broker.Stop()).To(Succeed())
				const secondEncryptionPassword = `{"primary":true,"label":"my-second-password","password":{"secret":"verysecretcoolpassword"}}`
				encryptionPasswords = fmt.Sprintf("[%s]", secondEncryptionPassword)
				Expect(tryStartBroker(true, encryptionPasswords)).To(MatchError(ContainSubstring(`the password labelled "my-first-password" must be supplied to decrypt the database`)))

				By("restarting the broker with a different primary password and with the initial password")
				firstEncryptionPassword = `{"primary":false,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
				startBroker(true, encryptionPasswords)
				Expect(broker.Stdout.String()).To(SatisfyAll(
					MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`),
					MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`),
				))
			})
		})

		When("database re-encryption fails", func() {
			It("can restart re-encrypting", func() {
				By("starting the broker with a password")
				firstEncryptionPassword := `{"primary":true,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				encryptionPasswords := fmt.Sprintf("[%s]", firstEncryptionPassword)
				startBroker(true, encryptionPasswords)
				Expect(broker.Stdout.String()).To(SatisfyAll(
					MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-first-password","previous-primary":"none"}}`),
					MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-first-password"}}`),
				))

				By("creating a service instance and checking fields are stored encrypted")
				serviceInstance := must(broker.Provision(serviceOfferingGUID, servicePlanGUID, testdrive.WithProvisionParams(provisionParams)))

				By("corrupting the DB")
				var record models.ServiceInstanceDetails
				findRecord(&record, serviceInstanceIDQuery, serviceInstance.GUID)
				recordToRecover := record
				record.OtherDetails = []byte("something-that-cannot-be-decrypted")
				Expect(dbConn.Save(&record).Error).NotTo(HaveOccurred())

				By("restarting the broker with a different primary password")
				Expect(broker.Stop()).To(Succeed())
				firstEncryptionPassword = `{"primary":false,"label":"my-first-password","password":{"secret":"supersecretcoolpassword"}}`
				const secondEncryptionPassword = `{"primary":true,"label":"my-second-password","password":{"secret":"verysecretcoolpassword"}}`
				encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
				Expect(tryStartBroker(true, encryptionPasswords)).To(MatchError(SatisfyAll(
					MatchRegexp(`cloud-service-broker.refusing to encrypt the database as some fields cannot be successfully read`),
					MatchRegexp(`decode error for service instance details "\S+": decryption error: cipher: message authentication failed`),
				)))

				By("checking password metadata are not removed")
				Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeTrue())
				Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeFalse())

				By("fixing the corrupted value")
				Expect(dbConn.Save(&recordToRecover).Error).NotTo(HaveOccurred())

				By("restarting the broker with same config")
				encryptionPasswords = fmt.Sprintf("[%s, %s]", firstEncryptionPassword, secondEncryptionPassword)
				startBroker(true, encryptionPasswords)
				Expect(broker.Stdout.String()).To(SatisfyAll(
					MatchRegexp(`cloud-service-broker.rotating-database-encryption\S*"data":{"new-primary":"my-second-password","previous-primary":"my-first-password"}}`),
					MatchRegexp(`cloud-service-broker.database-encryption\S*"data":{"primary":"my-second-password"}}`),
				))

				By("checking password metadata are updated")
				Expect(persistedPasswordMetadata("my-first-password").Primary).To(BeFalse())
				Expect(persistedPasswordMetadata("my-second-password").Primary).To(BeTrue())
			})
		})
	})

	When("encryption is turned on for a corrupted database", func() {
		BeforeEach(func() {
			Expect(dbConn.Migrator().CreateTable(&models.ServiceInstanceDetails{})).NotTo(HaveOccurred())
			Expect(dbConn.Create(&models.ServiceInstanceDetails{
				ID:           "fake-id-1",
				OtherDetails: []byte(`{"valid":"json"}`),
			}).Error).NotTo(HaveOccurred())
			Expect(dbConn.Create(&models.ServiceInstanceDetails{
				ID:           "fake-id-2",
				OtherDetails: []byte("not-json"),
			}).Error).NotTo(HaveOccurred())
		})

		It("fails without encrypting anything", func() {
			const encryptionPasswords = `[{"primary":true,"label":"my-password","password":{"secret":"supersecretcoolpassword"}}]`
			Expect(tryStartBroker(true, encryptionPasswords)).To(MatchError(SatisfyAll(
				MatchRegexp(`cloud-service-broker.refusing to encrypt the database as some fields cannot be successfully read`),
				MatchRegexp(`decode error for service instance details "fake-id-2": JSON parse error: invalid character 'o' in literal null \(expecting 'u'\)`),
			)))

			var receiver models.ServiceInstanceDetails
			Expect(dbConn.Where("id= ? ", "fake-id-1").First(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.OtherDetails).To(Equal([]byte(`{"valid":"json"}`)))
		})
	})

	When("a password is removed for a corrupt database", func() {
		It("does not clean up the password metadata", func() {
			By("registering the password")
			const encryptionPasswords = `[{"label":"obsolete","password":{"secret":"supersecretcoolpassword"}}]`
			startBroker(false, encryptionPasswords)
			Expect(broker.Stop()).To(Succeed())

			var receiver models.PasswordMetadata
			Expect(dbConn.Where("label=?", "obsolete").First(&receiver).Error).NotTo(HaveOccurred())

			By("corrupting the database")
			Expect(dbConn.Create(&models.ServiceInstanceDetails{
				ID:           "fake-id-2",
				OtherDetails: []byte("not-json"),
			}).Error).NotTo(HaveOccurred())

			By("starting the broker and checking that the password is not removed")
			startBroker(false, "")

			Expect(dbConn.Where("label=?", "obsolete").First(&receiver).Error).NotTo(HaveOccurred())
			Expect(broker.Stdout.String()).To(SatisfyAll(
				ContainSubstring(`cloud-service-broker.database-field-error`),
				ContainSubstring(`decode error for service instance details \"fake-id-2\": JSON parse error: invalid character 'o' in literal null (expecting 'u')`),
			))
		})
	})
})
