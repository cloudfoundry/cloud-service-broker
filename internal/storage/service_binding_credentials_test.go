package storage_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceBindingCredentials", func() {
	Describe("CreateServiceBindingCredentials", func() {
		It("creates the right object in the database", func() {
			err := store.CreateServiceBindingCredentials(storage.ServiceBindingCredentials{
				ServiceGUID:         "fake-service-id",
				ServiceInstanceGUID: "fake-instance-id",
				BindingGUID:         "fake-binding-id",
				Credentials: storage.JSONObject{
					"fake-cred-1": "fake-val-1",
					"fake-cred-2": "fake-val-2",
				},
			})

			Expect(err).NotTo(HaveOccurred())
			var receiver models.ServiceBindingCredentials
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ServiceID).To(Equal("fake-service-id"))
			Expect(receiver.ServiceInstanceID).To(Equal("fake-instance-id"))
			Expect(receiver.BindingID).To(Equal("fake-binding-id"))
			Expect(receiver.OtherDetails).To(MatchJSON(`{
				"encrypted":{
					"fake-cred-1": "fake-val-1",
					"fake-cred-2": "fake-val-2"
				}
			}`))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.CreateServiceBindingCredentials(storage.ServiceBindingCredentials{})
				Expect(err).To(MatchError("error encoding credentials: encryption error: bang"))
			})
		})
	})

	Describe("GetServiceBindingCredentials", func() {
		BeforeEach(func() {
			addFakeServiceCredentialBindings()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetServiceBindingCredentials("fake-binding-id", "fake-instance-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(r.ServiceGUID).To(Equal("fake-service-id"))
			Expect(r.ServiceInstanceGUID).To(Equal("fake-instance-id"))
			Expect(r.BindingGUID).To(Equal("fake-binding-id"))
			Expect(r.Credentials).To(Equal(storage.JSONObject{
				"decrypted": map[string]interface{}{
					"foo": "baz",
					"bar": "quz",
				},
			}))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetServiceBindingCredentials("fake-binding-id", "fake-instance-id")
				Expect(err).To(MatchError(`error decoding binding credentials "fake-binding-id": decryption error: bang`))
			})
		})

		When("nothing is found", func() {
			It("returns an error", func() {
				_, err := store.GetServiceBindingCredentials("not-there", "also-not-there")
				Expect(err).To(MatchError(`could not find binding credentials for binding "not-there" and service instance "also-not-there"`))
			})
		})

		DescribeTable(
			"JSON parsing",
			func(input []byte) {
				encryptor.DecryptReturns(input, nil)

				r, err := store.GetServiceBindingCredentials("fake-binding-id", "fake-instance-id")
				Expect(err).NotTo(HaveOccurred())
				Expect(r.Credentials).To(Equal(storage.JSONObject(nil)))
			},
			Entry("null", []byte(`null`)),
			Entry("empty", []byte(``)),
			Entry("nil", []byte(nil)),
		)
	})

	Describe("GetServiceBindingIDsForServiceInstance", func() {
		BeforeEach(func() {
			addFakeServiceCredentialBindings()
		})

		It("returns the bindings for the service instance", func() {
			r, err := store.GetServiceBindingIDsForServiceInstance("fake-instance-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(r).To(ConsistOf("fake-binding-id", "fake-other-binding-id"))
		})

		It("returns an empty slice when no bindings are found", func() {
			r, err := store.GetServiceBindingIDsForServiceInstance("instance-with-no-bindings")
			Expect(err).NotTo(HaveOccurred())
			Expect(r).To(BeEmpty())
		})
	})

	Describe("ExistsServiceBindingCredentials", func() {
		BeforeEach(func() {
			addFakeServiceCredentialBindings()
		})

		It("reads the result from the database", func() {
			Expect(store.ExistsServiceBindingCredentials("not-there", "also-not-there")).To(BeFalse())
			Expect(store.ExistsServiceBindingCredentials("fake-binding-id", "also-not-there")).To(BeFalse())
			Expect(store.ExistsServiceBindingCredentials("not-there", "fake-instance-id")).To(BeFalse())
			Expect(store.ExistsServiceBindingCredentials("fake-binding-id", "fake-instance-id")).To(BeTrue())
			Expect(store.ExistsServiceBindingCredentials("fake-binding-id", "fake-other-instance-id")).To(BeTrue())
		})
	})

	Describe("DeleteServiceBindingCredentials", func() {
		BeforeEach(func() {
			addFakeServiceCredentialBindings()
		})

		It("deletes from the database", func() {
			Expect(store.ExistsServiceBindingCredentials("fake-binding-id", "fake-instance-id")).To(BeTrue())

			Expect(store.DeleteServiceBindingCredentials("fake-binding-id", "fake-instance-id")).NotTo(HaveOccurred())

			Expect(store.ExistsServiceBindingCredentials("fake-binding-id", "fake-instance-id")).To(BeFalse())
		})

		It("is idempotent", func() {
			Expect(store.DeleteServiceBindingCredentials("not-there", "also-not-there")).NotTo(HaveOccurred())
		})
	})
})

func addFakeServiceCredentialBindings() {
	Expect(db.Create(&models.ServiceBindingCredentials{
		OtherDetails:      []byte(`{"foo":"bar"}`),
		ServiceID:         "fake-other-service-id",
		ServiceInstanceID: "fake-instance-id",
		BindingID:         "fake-other-binding-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ServiceBindingCredentials{
		OtherDetails:      []byte(`{"foo":"baz","bar":"quz"}`),
		ServiceID:         "fake-service-id",
		ServiceInstanceID: "fake-instance-id",
		BindingID:         "fake-binding-id",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ServiceBindingCredentials{
		OtherDetails:      []byte(`{"foo":"boz"}`),
		ServiceID:         "fake-other-service-id",
		ServiceInstanceID: "fake-other-instance-id",
		BindingID:         "fake-binding-id",
	}).Error).NotTo(HaveOccurred())
}
