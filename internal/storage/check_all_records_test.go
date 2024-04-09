package storage_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
)

var _ = Describe("CheckAllRecords", func() {

	BeforeEach(func() {
		encryptor.DecryptStub = func(bytes []byte) ([]byte, error) {
			if string(bytes) == `cannot-be-decrypted` {
				return nil, errors.New("fake decryption error")
			}

			return bytes, nil
		}

		addFakeServiceCredentialBindings()
		addFakeProvisionRequestDetails()
		addFakeBindRequestDetails()
		addFakeServiceInstanceDetails()
		addFakeTerraformDeployments()
	})

	It("does not fail", func() {
		Expect(store.CheckAllRecords()).NotTo(HaveOccurred())
	})

	When("the database contains invalid data", func() {
		BeforeEach(func() {
			Expect(db.Create(&models.ServiceBindingCredentials{
				OtherDetails:      []byte(`binding-not-json-2`),
				ServiceID:         "fake-other-service-id",
				ServiceInstanceID: "fake-bad-instance-id",
				BindingID:         "fake-bad-binding-id-1",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.ServiceBindingCredentials{
				OtherDetails:      []byte(`cannot-be-decrypted`),
				ServiceID:         "fake-other-service-id",
				ServiceInstanceID: "fake-bad-instance-id",
				BindingID:         "fake-bad-binding-id-2",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.ProvisionRequestDetails{
				RequestDetails:    []byte(`cannot-be-decrypted`),
				ServiceInstanceID: "fake-bad-instance-id-1",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.ProvisionRequestDetails{
				RequestDetails:    []byte(`request-details-not-json`),
				ServiceInstanceID: "fake-bad-instance-id-2",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.BindRequestDetails{
				RequestDetails:    []byte(`cannot-be-decrypted`),
				ServiceBindingID:  "fake-bad-binding-id-1",
				ServiceInstanceID: "fake-bad-instance-id",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.BindRequestDetails{
				RequestDetails:    []byte(`request-details-not-json`),
				ServiceBindingID:  "fake-bad-binding-id-2",
				ServiceInstanceID: "fake-bad-instance-id",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.ServiceInstanceDetails{
				ID:           "fake-bad-instance-id-1",
				OtherDetails: []byte(`service-instance-not-json`),
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.ServiceInstanceDetails{
				ID:           "fake-bad-instance-id-2",
				OtherDetails: []byte(`cannot-be-decrypted`),
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.TerraformDeployment{
				ID:                   "fake-bad-id-1",
				Workspace:            []byte("cannot-be-decrypted"),
				LastOperationType:    "create",
				LastOperationState:   "succeeded",
				LastOperationMessage: "amazing",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.TerraformDeployment{
				ID:                   "fake-bad-id-2",
				Workspace:            []byte("workspace-not-json"),
				LastOperationType:    "create",
				LastOperationState:   "succeeded",
				LastOperationMessage: "amazing",
			}).Error).NotTo(HaveOccurred())

			Expect(db.Create(&models.TerraformDeployment{
				ID:                   "fake-bad-id-3",
				Workspace:            []byte(`{"tfstate":42}`),
				LastOperationType:    "create",
				LastOperationState:   "succeeded",
				LastOperationMessage: "amazing",
			}).Error).NotTo(HaveOccurred())
		})

		It("returns all errors", func() {
			Expect(store.CheckAllRecords()).To(MatchError(And(
				ContainSubstring(`decode error for service binding credential "fake-bad-binding-id-1": JSON parse error: invalid character 'b' looking for beginning of value`),
				ContainSubstring(`decode error for service binding credential "fake-bad-binding-id-2": decryption error: fake decryption error`),
				ContainSubstring(`decode error for provision request details "fake-bad-instance-id-1": decryption error: fake decryption error`),
				ContainSubstring(`decode error for provision request details "fake-bad-instance-id-2": JSON parse error: invalid character 'r' looking for beginning of value`),
				ContainSubstring(`decode error for binding request details "fake-bad-binding-id-1": decryption error: fake decryption error`),
				ContainSubstring(`decode error for binding request details "fake-bad-binding-id-2": JSON parse error: invalid character 'r' looking for beginning of value`),
				ContainSubstring(`decode error for service instance details "fake-bad-instance-id-1": JSON parse error: invalid character 's' looking for beginning of value`),
				ContainSubstring(`decode error for service instance details "fake-bad-instance-id-2": decryption error: fake decryption error`),
				ContainSubstring(`decode error for terraform deployment "fake-bad-id-1": decryption error: fake decryption error`),
				ContainSubstring(`decode error for terraform deployment "fake-bad-id-2": JSON parse error: invalid character 'w' looking for beginning of value`),
				ContainSubstring(`decode error for terraform deployment "fake-bad-id-3": JSON parse error: json: cannot unmarshal number into Go struct field TerraformWorkspace.tfstate of type []uint8`),
			)))
		})
	})
})
