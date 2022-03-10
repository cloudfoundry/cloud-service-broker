package storage_test

import (
	"encoding/json"
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BindRequestDetails", func() {
	Describe("StoreBindRequestDetails", func() {
		serviceInstanceId := "fake-instance-id"
		serviceBindingId := "fake-binding-id"

		It("creates the right object in the database", func() {
			err := store.StoreBindRequestDetails(storage.BindRequestDetails{
				ServiceInstanceGUID: serviceInstanceId,
				ServiceBindingGUID:  serviceBindingId,
				RequestDetails:      json.RawMessage(`{"foo":"bar"}`),
			})
			Expect(err).NotTo(HaveOccurred())

			var receiver models.BindRequestDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ServiceInstanceId).To(Equal(serviceInstanceId))
			Expect(receiver.ServiceBindingId).To(Equal(serviceBindingId))
			Expect(receiver.RequestDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
		})

		It("does not store when params are nil", func() {
			err := store.StoreBindRequestDetails(storage.BindRequestDetails{
				ServiceInstanceGUID: serviceInstanceId,
				ServiceBindingGUID:  serviceBindingId,
				RequestDetails:      nil,
			})
			Expect(err).NotTo(HaveOccurred())

			var receiver models.BindRequestDetails
			Expect(db.First(&receiver).Error).To(MatchError("record not found"))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreBindRequestDetails(storage.BindRequestDetails{
					ServiceInstanceGUID: serviceInstanceId,
					ServiceBindingGUID:  serviceBindingId,
					RequestDetails:      json.RawMessage(`{"foo":"bar"}`),
				})
				Expect(err).To(MatchError("error encoding details: encryption error: bang"))
			})
		})

		When("details for the binding already exist in the database", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.BindRequestDetails{
					ServiceBindingId: serviceBindingId,
					RequestDetails:   []byte(`{"foo":"bar"}`),
				}).Error).NotTo(HaveOccurred())
			})

			It("errors", func() {
				err := store.StoreBindRequestDetails(storage.BindRequestDetails{
					ServiceInstanceGUID: serviceInstanceId,
					ServiceBindingGUID:  serviceBindingId,
					RequestDetails:      json.RawMessage(`{"foo":"qux"}`),
				})
				Expect(err).To(MatchError(ContainSubstring("error saving bind request details: Binding already exists")))
			})
		})
	})

	Describe("GetBindRequestDetails", func() {
		BeforeEach(func() {
			addFakeBindRequestDetails()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetBindRequestDetails("fake-binding-id", "fake-instance-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(r).To(Equal(json.RawMessage(`{"decrypted":{"foo":"bar"}}`)))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetBindRequestDetails("fake-binding-id", "fake-instance-id")
				Expect(err).To(MatchError(`error decoding bind request details "fake-binding-id": decryption error: bang`))
			})
		})

		When("nothing is found", func() {
			It("returns nil details when binding not found", func() {
				details, err := store.GetBindRequestDetails("not-there", "fake-instance-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(details).To(BeNil())
			})

			It("returns nil details when instance not found", func() {
				details, err := store.GetBindRequestDetails("fake-binding-id", "not-there")
				Expect(err).NotTo(HaveOccurred())
				Expect(details).To(BeNil())
			})
		})
	})

	Describe("DeleteBindRequestDetails", func() {
		BeforeEach(func() {
			addFakeBindRequestDetails()
		})

		It("deletes from the database", func() {
			exists := func() bool {
				var count int64
				Expect(db.Model(&models.BindRequestDetails{}).Where(`service_binding_id="fake-binding-id"`).Count(&count).Error).NotTo(HaveOccurred())
				return count != 0
			}
			Expect(exists()).To(BeTrue())

			Expect(store.DeleteBindRequestDetails("fake-binding-id", "fake-instance-id")).NotTo(HaveOccurred())

			Expect(exists()).To(BeFalse())
		})

		When("binding doesnt exist", func() {
			It("is idempotent when binding not found", func() {
				Expect(store.DeleteBindRequestDetails("not-there", "fake-instance-id")).NotTo(HaveOccurred())
			})

			It("is idempotent when instance not found", func() {
				Expect(store.DeleteBindRequestDetails("fake-binding-id", "not-there")).NotTo(HaveOccurred())
			})
		})
	})
})

func addFakeBindRequestDetails() {
	Expect(db.Create(&models.BindRequestDetails{
		RequestDetails:    []byte(`{"foo":"bar"}`),
		ServiceBindingId:  "fake-binding-id",
		ServiceInstanceId: "fake-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		RequestDetails:    []byte(`{"foo":"baz","bar":"quz"}`),
		ServiceBindingId:  "fake-other-binding-id",
		ServiceInstanceId: "fake-other-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		RequestDetails:    []byte(`{"foo":"boz"}`),
		ServiceBindingId:  "fake-yet-another-binding-id",
		ServiceInstanceId: "fake-yet-another-instance-id",
	}).Error).NotTo(HaveOccurred())
}
