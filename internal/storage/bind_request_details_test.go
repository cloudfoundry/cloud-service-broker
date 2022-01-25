package storage_test

import (
	"encoding/json"
	"errors"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BindRequestDetails", func() {
	Describe("StoreBindRequestDetails", func() {
		It("creates the right object in the database", func() {
			err := store.StoreBindRequestDetails("fake-binding-id", json.RawMessage(`{"foo":"bar"}`))
			Expect(err).NotTo(HaveOccurred())

			var receiver models.BindRequestDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ServiceBindingId).To(Equal("fake-binding-id"))
			Expect(receiver.RequestDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
		})

		It("does not store when params are nil", func() {
			err := store.StoreBindRequestDetails("fake-binding-id", nil)
			Expect(err).NotTo(HaveOccurred())

			var receiver models.BindRequestDetails
			Expect(db.First(&receiver).Error).To(MatchError("record not found"))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreBindRequestDetails("fake-binding-id", json.RawMessage(`{"foo":"bar"}`))
				Expect(err).To(MatchError("error encoding details: bang"))
			})
		})

		When("details for the binding already exist in the database", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.BindRequestDetails{
					ServiceBindingId: "fake-binding-id",
					RequestDetails:   []byte(`{"foo":"bar"}`),
				}).Error).NotTo(HaveOccurred())
			})

			It("updates the existing record", func() {
				err := store.StoreBindRequestDetails("fake-binding-id", json.RawMessage(`{"foo":"qux"}`))
				Expect(err).NotTo(HaveOccurred())

				var receiver []models.BindRequestDetails
				Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver).To(HaveLen(1))
				Expect(receiver[0].ServiceBindingId).To(Equal("fake-binding-id"))
				Expect(receiver[0].RequestDetails).To(Equal([]byte(`{"encrypted":{"foo":"qux"}}`)))
			})
		})
	})

	Describe("GetBindRequestDetails", func() {
		BeforeEach(func() {
			addFakeBindRequestDetails()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetBindRequestDetails("fake-binding-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(r).To(Equal(json.RawMessage(`{"decrypted":{"foo":"bar"}}`)))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetBindRequestDetails("fake-binding-id")
				Expect(err).To(MatchError("error decoding bind request details: bang"))
			})
		})

		When("nothing is found", func() {
			It("returns nil details", func() {
				details, err := store.GetBindRequestDetails("not-there")
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

			Expect(store.DeleteBindRequestDetails("fake-binding-id")).NotTo(HaveOccurred())

			Expect(exists()).To(BeFalse())
		})

		It("is idempotent", func() {
			Expect(store.DeleteBindRequestDetails("not-there")).NotTo(HaveOccurred())
		})
	})
})

func addFakeBindRequestDetails() {
	Expect(db.Create(&models.BindRequestDetails{
		RequestDetails:   []byte(`{"foo":"bar"}`),
		ServiceBindingId: "fake-binding-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		RequestDetails:   []byte(`{"foo":"baz","bar":"quz"}`),
		ServiceBindingId: "fake-other-binding-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		RequestDetails:   []byte(`{"foo":"boz"}`),
		ServiceBindingId: "fake-yet-another-binding-id",
	}).Error).NotTo(HaveOccurred())
}
