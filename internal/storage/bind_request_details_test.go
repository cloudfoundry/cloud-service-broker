package storage_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
)

var _ = Describe("BindRequestDetails", func() {
	Describe("StoreBindRequestDetails", func() {
		const serviceInstanceID = "fake-instance-id"
		const serviceBindingID = "fake-binding-id"

		It("creates the right object in the database", func() {
			err := store.StoreBindRequestDetails(
				serviceBindingID,
				serviceInstanceID,
				storage.JSONObject{"bar": "baz"},
				storage.JSONObject{"foo": "bar"},
			)
			Expect(err).NotTo(HaveOccurred())

			var receiver models.BindRequestDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ServiceInstanceID).To(Equal(serviceInstanceID))
			Expect(receiver.ServiceBindingID).To(Equal(serviceBindingID))
			Expect(receiver.BindResource).To(Equal([]byte(`{"encrypted":{"bar":"baz"}}`)))
			Expect(receiver.Parameters).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
		})

		When("bind resource is nil", func() {
			It("stores JSON null value", func() {
				err := store.StoreBindRequestDetails(
					serviceBindingID,
					serviceInstanceID,
					nil,
					storage.JSONObject{"foo": "bar"},
				)
				Expect(err).NotTo(HaveOccurred())

				var receiver models.BindRequestDetails
				Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver.ServiceInstanceID).To(Equal(serviceInstanceID))
				Expect(receiver.ServiceBindingID).To(Equal(serviceBindingID))
				Expect(receiver.BindResource).To(Equal([]byte(`{"encrypted":null}`)))
				Expect(receiver.Parameters).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
			})
		})

		When("parameters are nil", func() {
			It("stores JSON null value", func() {
				err := store.StoreBindRequestDetails(
					serviceBindingID,
					serviceInstanceID,
					storage.JSONObject{"foo": "bar"},
					nil,
				)
				Expect(err).NotTo(HaveOccurred())

				var receiver models.BindRequestDetails
				Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver.ServiceInstanceID).To(Equal(serviceInstanceID))
				Expect(receiver.ServiceBindingID).To(Equal(serviceBindingID))
				Expect(receiver.BindResource).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
				Expect(receiver.Parameters).To(Equal([]byte(`{"encrypted":null}`)))
			})
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreBindRequestDetails(
					serviceBindingID,
					serviceInstanceID,
					nil,
					storage.JSONObject{"foo": "bar"},
				)
				Expect(err).To(MatchError(MatchRegexp(`error encoding bind request details \w+: encryption error: bang`)))
			})
		})

		When("details for the binding already exist in the database", func() {
			BeforeEach(func() {
				Expect(db.Create(&models.BindRequestDetails{
					ServiceBindingID: serviceBindingID,
					Parameters:       []byte(`{"foo":"bar"}`),
				}).Error).NotTo(HaveOccurred())
			})

			It("errors", func() {
				err := store.StoreBindRequestDetails(
					serviceBindingID,
					serviceInstanceID,
					nil,
					storage.JSONObject{"foo": "qux"},
				)
				Expect(err).To(MatchError(ContainSubstring("error saving bind request details: Binding already exists")))
			})
		})
	})

	Describe("GetBindRequestDetails", func() {
		BeforeEach(func() {
			addFakeBindRequestDetails()
		})

		It("reads the right object from the database", func() {
			actual, err := store.GetBindRequestDetails("fake-binding-id", "fake-instance-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(actual.ServiceBindingGUID).To(Equal("fake-binding-id"))
			Expect(actual.ServiceInstanceGUID).To(Equal("fake-instance-id"))
			Expect(actual.BindResource).To(Equal(storage.JSONObject{"decrypted": map[string]any{"bar": "baz"}}))
			Expect(actual.Parameters).To(Equal(storage.JSONObject{"decrypted": map[string]any{"foo": "bar"}}))
		})

		When("bindResource is null", func() {
			It("returns empty JSONObject", func() {
				actual, err := store.GetBindRequestDetails("empty-bind-resource-binding-id", "fake-instance-id-4")
				Expect(err).NotTo(HaveOccurred())
				Expect(actual.BindResource).To(Equal(storage.JSONObject{"decrypted": nil}))
				Expect(actual.Parameters).To(Equal(storage.JSONObject{"decrypted": map[string]any{"foo": "bar"}}))
			})
		})

		When("parameters is null", func() {
			It("returns empty JSONObject", func() {
				actual, err := store.GetBindRequestDetails("empty-parameters-binding-id", "fake-instance-id-5")
				Expect(err).NotTo(HaveOccurred())
				Expect(actual.BindResource).To(Equal(storage.JSONObject{"decrypted": map[string]any{"foo": "bar"}}))
				Expect(actual.Parameters).To(Equal(storage.JSONObject{"decrypted": nil}))
			})
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetBindRequestDetails("fake-binding-id", "fake-instance-id")
				Expect(err).To(MatchError(MatchRegexp(`error decoding bind request detail \w+ for "fake-binding-id": decryption error: bang`)))
			})
		})

		When("nothing is found", func() {
			It("returns empty struct and nil parameters when binding not found", func() {
				actual, err := store.GetBindRequestDetails("not-there", "fake-instance-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(BeZero())
				Expect(actual.Parameters).To(BeNil())
			})

			It("returns empty struct and nil parameters when instance not found", func() {
				actual, err := store.GetBindRequestDetails("fake-binding-id", "not-there")
				Expect(err).NotTo(HaveOccurred())
				Expect(actual).To(BeZero())
				Expect(actual.Parameters).To(BeNil())
			})
		})

		DescribeTable(
			"JSON parsing",
			func(input []byte) {
				encryptor.DecryptReturns(input, nil)

				actual, err := store.GetBindRequestDetails("fake-binding-id", "fake-instance-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(actual.BindResource).To(Equal(storage.JSONObject(nil)))
				Expect(actual.Parameters).To(Equal(storage.JSONObject(nil)))
			},
			Entry("null", []byte(`null`)),
			Entry("empty", []byte(``)),
			Entry("nil", []byte(nil)),
		)
	})

	Describe("DeleteBindRequestDetails", func() {
		BeforeEach(func() {
			addFakeBindRequestDetails()
		})

		It("deletes physically from the database", func() {
			exists := func() bool {
				var count int64
				Expect(db.Model(&models.BindRequestDetails{}).Where(`service_binding_id="fake-binding-id"`).Unscoped().Count(&count).Error).NotTo(HaveOccurred())
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
		BindResource:      []byte(`{"bar":"baz"}`),
		Parameters:        []byte(`{"foo":"bar"}`),
		ServiceBindingID:  "fake-binding-id",
		ServiceInstanceID: "fake-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		BindResource:      []byte(`{"foo":"baz","bar":"quz"}`),
		Parameters:        []byte(`{"foo":"baz","bar":"quz"}`),
		ServiceBindingID:  "fake-other-binding-id",
		ServiceInstanceID: "fake-other-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		BindResource:      []byte(`{"foo":"boz"}`),
		Parameters:        []byte(`{"foo":"boz"}`),
		ServiceBindingID:  "fake-yet-another-binding-id",
		ServiceInstanceID: "fake-yet-another-instance-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		BindResource:      []byte(`null`),
		Parameters:        []byte(`{"foo":"bar"}`),
		ServiceBindingID:  "empty-bind-resource-binding-id",
		ServiceInstanceID: "fake-instance-id-4",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.BindRequestDetails{
		BindResource:      []byte(`{"foo":"bar"}`),
		Parameters:        []byte(`null`),
		ServiceBindingID:  "empty-parameters-binding-id",
		ServiceInstanceID: "fake-instance-id-5",
	}).Error).NotTo(HaveOccurred())
}
