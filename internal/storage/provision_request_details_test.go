package storage_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProvisionRequestDetails", func() {
	Describe("StoreProvisionRequestDetails", func() {
		It("creates the right object in the database", func() {
			err := store.StoreProvisionRequestDetails("fake-instance-id", map[string]interface{}{"foo": "bar"})
			Expect(err).NotTo(HaveOccurred())

			var receiver models.ProvisionRequestDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ServiceInstanceId).To(Equal("fake-instance-id"))
			Expect(receiver.RequestDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreProvisionRequestDetails("fake-instance-id", map[string]interface{}{"foo": "bar"})
				Expect(err).To(MatchError("error encoding details: encryption error: bang"))
			})
		})

		When("details for the instance already exist in the database", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.ProvisionRequestDetails{
					ServiceInstanceId: "fake-instance-id",
					RequestDetails:    []byte(`{"foo":"bar"}`),
				}).Error).NotTo(HaveOccurred())
			})

			It("updates the existing record", func() {
				err := store.StoreProvisionRequestDetails("fake-instance-id", map[string]interface{}{"foo": "qux"})
				Expect(err).NotTo(HaveOccurred())

				var receiver []models.ProvisionRequestDetails
				Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver).To(HaveLen(1))
				Expect(receiver[0].ServiceInstanceId).To(Equal("fake-instance-id"))
				Expect(receiver[0].RequestDetails).To(Equal([]byte(`{"encrypted":{"foo":"qux"}}`)))
			})
		})
	})

	Describe("GetProvisionRequestDetails", func() {
		BeforeEach(func() {
			addFakeProvisionRequestDetails()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetProvisionRequestDetails("fake-instance-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(r).To(Equal(storage.JSONObject{"decrypted": map[string]interface{}{"foo": "bar"}}))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetProvisionRequestDetails("fake-instance-id")
				Expect(err).To(MatchError(`error decoding provision request details "fake-instance-id": decryption error: bang`))
			})
		})

		When("JSON parsing fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns([]byte("not-json"), nil)

				_, err := store.GetProvisionRequestDetails("fake-instance-id")
				Expect(err).To(MatchError(`error decoding provision request details "fake-instance-id": JSON parse error: invalid character 'o' in literal null (expecting 'u')`))
			})
		})

		When("nothing is found", func() {
			It("returns an error", func() {
				_, err := store.GetProvisionRequestDetails("not-there")
				Expect(err).To(MatchError("could not find provision request details for service instance: not-there"))
			})
		})
	})

	Describe("DeleteProvisionRequestDetails", func() {
		BeforeEach(func() {
			addFakeProvisionRequestDetails()
		})

		It("deletes from the database", func() {
			exists := func() bool {
				var count int64
				Expect(db.Model(&models.ProvisionRequestDetails{}).Where(`service_instance_id="fake-instance-id"`).Count(&count).Error).NotTo(HaveOccurred())
				return count != 0
			}
			Expect(exists()).To(BeTrue())

			Expect(store.DeleteProvisionRequestDetails("fake-instance-id")).NotTo(HaveOccurred())

			Expect(exists()).To(BeFalse())
		})

		It("is idempotent", func() {
			Expect(store.DeleteProvisionRequestDetails("not-there")).NotTo(HaveOccurred())
		})
	})
})

func addFakeProvisionRequestDetails() {
	Expect(db.Create(&models.ProvisionRequestDetails{
		RequestDetails:    []byte(`{"foo":"bar"}`),
		ServiceInstanceId: "fake-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ProvisionRequestDetails{
		RequestDetails:    []byte(`{"foo":"baz","bar":"quz"}`),
		ServiceInstanceId: "fake-other-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ProvisionRequestDetails{
		RequestDetails:    []byte(`{"foo":"boz"}`),
		ServiceInstanceId: "fake-yet-another-instance-id",
	}).Error).NotTo(HaveOccurred())
}
