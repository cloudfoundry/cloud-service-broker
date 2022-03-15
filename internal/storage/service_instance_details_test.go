package storage_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceDetails", func() {
	Describe("StoreServiceInstanceDetails", func() {
		It("creates the right object in the database", func() {
			err := store.StoreServiceInstanceDetails(storage.ServiceInstanceDetails{
				GUID:             "fake-guid",
				Name:             "fake-name",
				Location:         "fake-location",
				URL:              "fake-url",
				Outputs:          map[string]interface{}{"foo": "bar"},
				ServiceGUID:      "fake-service-guid",
				PlanGUID:         "fake-plan-guid",
				SpaceGUID:        "fake-space-guid",
				OrganizationGUID: "fake-org-guid",
				OperationType:    "fake-operation-type",
				OperationGUID:    "fake-operation-guid",
			})
			Expect(err).NotTo(HaveOccurred())

			var receiver models.ServiceInstanceDetails
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ID).To(Equal("fake-guid"))
			Expect(receiver.Name).To(Equal("fake-name"))
			Expect(receiver.Location).To(Equal("fake-location"))
			Expect(receiver.Url).To(Equal("fake-url"))
			Expect(receiver.OtherDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
			Expect(receiver.ServiceId).To(Equal("fake-service-guid"))
			Expect(receiver.PlanId).To(Equal("fake-plan-guid"))
			Expect(receiver.SpaceGuid).To(Equal("fake-space-guid"))
			Expect(receiver.OrganizationGuid).To(Equal("fake-org-guid"))
			Expect(receiver.OperationType).To(Equal("fake-operation-type"))
			Expect(receiver.OperationId).To(Equal("fake-operation-guid"))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreServiceInstanceDetails(storage.ServiceInstanceDetails{})
				Expect(err).To(MatchError("error encoding details: encryption error: bang"))
			})
		})

		When("details for the instance already exist in the database", func() {
			BeforeEach(func() {
				addFakeServiceInstanceDetails()
			})

			It("updates the existing record", func() {
				err := store.StoreServiceInstanceDetails(storage.ServiceInstanceDetails{
					GUID:             "fake-id-1",
					Name:             "fake-name",
					Location:         "fake-location",
					URL:              "fake-url",
					Outputs:          map[string]interface{}{"foo": "bar"},
					ServiceGUID:      "fake-service-guid",
					PlanGUID:         "fake-plan-guid",
					SpaceGUID:        "fake-space-guid",
					OrganizationGUID: "fake-org-guid",
					OperationType:    "fake-operation-type",
					OperationGUID:    "fake-operation-guid",
				})
				Expect(err).NotTo(HaveOccurred())

				var receiver models.ServiceInstanceDetails
				Expect(db.Where(`id = "fake-id-1"`).Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver.ID).To(Equal("fake-id-1"))
				Expect(receiver.Name).To(Equal("fake-name"))
				Expect(receiver.Location).To(Equal("fake-location"))
				Expect(receiver.Url).To(Equal("fake-url"))
				Expect(receiver.OtherDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
				Expect(receiver.ServiceId).To(Equal("fake-service-guid"))
				Expect(receiver.PlanId).To(Equal("fake-plan-guid"))
				Expect(receiver.SpaceGuid).To(Equal("fake-space-guid"))
				Expect(receiver.OrganizationGuid).To(Equal("fake-org-guid"))
				Expect(receiver.OperationType).To(Equal("fake-operation-type"))
				Expect(receiver.OperationId).To(Equal("fake-operation-guid"))
			})
		})
	})

	Describe("GetServiceInstanceDetails", func() {
		BeforeEach(func() {
			addFakeServiceInstanceDetails()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetServiceInstanceDetails("fake-id-2")
			Expect(err).NotTo(HaveOccurred())

			Expect(r.GUID).To(Equal("fake-id-2"))
			Expect(r.Location).To(Equal("fake-location-2"))
			Expect(r.URL).To(Equal("fake-url-2"))
			Expect(r.Outputs).To(Equal(storage.JSONObject{"decrypted": map[string]interface{}{"foo": "bar-2"}}))
			Expect(r.ServiceGUID).To(Equal("fake-service-id-2"))
			Expect(r.PlanGUID).To(Equal("fake-plan-id-2"))
			Expect(r.SpaceGUID).To(Equal("fake-space-guid-2"))
			Expect(r.OrganizationGUID).To(Equal("fake-org-guid-2"))
			Expect(r.OperationType).To(Equal("fake-operation-type-2"))
			Expect(r.OperationGUID).To(Equal("fake-operation-id-2"))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetServiceInstanceDetails("fake-id-1")
				Expect(err).To(MatchError(`error decoding service instance outputs "fake-id-1": decryption error: bang`))
			})
		})

		When("nothing is found", func() {
			It("returns an error", func() {
				_, err := store.GetServiceInstanceDetails("not-there")
				Expect(err).To(MatchError("could not find service instance details for: not-there"))
			})
		})

		When("OtherDetails field is empty", func() {
			It("succeeds with an empty result", func() {
				encryptor.DecryptReturns([]byte(""), nil)

				r, err := store.GetServiceInstanceDetails("fake-id-1")
				Expect(err).NotTo(HaveOccurred())

				Expect(r.Outputs).To(BeEmpty())
			})
		})
	})

	Describe("ExistsServiceInstanceDetails", func() {
		BeforeEach(func() {
			addFakeServiceInstanceDetails()
		})

		It("reads the result from the database", func() {
			Expect(store.ExistsServiceInstanceDetails("not-there")).To(BeFalse())
			Expect(store.ExistsServiceInstanceDetails("also-not-there")).To(BeFalse())
			Expect(store.ExistsServiceInstanceDetails("fake-id-1")).To(BeTrue())
			Expect(store.ExistsServiceInstanceDetails("fake-id-2")).To(BeTrue())
			Expect(store.ExistsServiceInstanceDetails("fake-id-3")).To(BeTrue())
		})
	})

	Describe("DeleteServiceInstanceDetails", func() {
		BeforeEach(func() {
			addFakeServiceInstanceDetails()
		})

		It("deletes from the database", func() {
			Expect(store.ExistsServiceInstanceDetails("fake-id-3")).To(BeTrue())

			Expect(store.DeleteServiceInstanceDetails("fake-id-3")).NotTo(HaveOccurred())

			Expect(store.ExistsServiceInstanceDetails("fake-id-3")).To(BeFalse())
		})

		It("is idempotent", func() {
			Expect(store.DeleteServiceInstanceDetails("not-there")).NotTo(HaveOccurred())
		})
	})
})

func addFakeServiceInstanceDetails() {
	Expect(db.Create(&models.ServiceInstanceDetails{
		ID:               "fake-id-1",
		Name:             "fake-name-1",
		Location:         "fake-location-1",
		Url:              "fake-url-1",
		OtherDetails:     []byte(`{"foo":"bar-1"}`),
		ServiceId:        "fake-service-id-1",
		PlanId:           "fake-plan-id-1",
		SpaceGuid:        "fake-space-guid-1",
		OrganizationGuid: "fake-org-guid-1",
		OperationType:    "fake-operation-type-1",
		OperationId:      "fake-operation-id-1",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ServiceInstanceDetails{
		ID:               "fake-id-2",
		Name:             "fake-name-2",
		Location:         "fake-location-2",
		Url:              "fake-url-2",
		OtherDetails:     []byte(`{"foo":"bar-2"}`),
		ServiceId:        "fake-service-id-2",
		PlanId:           "fake-plan-id-2",
		SpaceGuid:        "fake-space-guid-2",
		OrganizationGuid: "fake-org-guid-2",
		OperationType:    "fake-operation-type-2",
		OperationId:      "fake-operation-id-2",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ServiceInstanceDetails{
		ID:               "fake-id-3",
		Name:             "fake-name-3",
		Location:         "fake-location-3",
		Url:              "fake-url-3",
		OtherDetails:     []byte(`{"foo":"bar-3"}`),
		ServiceId:        "fake-service-id-3",
		PlanId:           "fake-plan-id-3",
		SpaceGuid:        "fake-space-guid-3",
		OrganizationGuid: "fake-org-guid-3",
		OperationType:    "fake-operation-type-3",
		OperationId:      "fake-operation-id-3",
	}).Error).NotTo(HaveOccurred())
}
