package storage_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
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
				Outputs:          map[string]any{"foo": "bar"},
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
			Expect(receiver.URL).To(Equal("fake-url"))
			Expect(receiver.OtherDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
			Expect(receiver.ServiceID).To(Equal("fake-service-guid"))
			Expect(receiver.PlanID).To(Equal("fake-plan-guid"))
			Expect(receiver.SpaceGUID).To(Equal("fake-space-guid"))
			Expect(receiver.OrganizationGUID).To(Equal("fake-org-guid"))
			Expect(receiver.OperationType).To(Equal("fake-operation-type"))
			Expect(receiver.OperationID).To(Equal("fake-operation-guid"))
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
					Outputs:          map[string]any{"foo": "bar"},
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
				Expect(receiver.URL).To(Equal("fake-url"))
				Expect(receiver.OtherDetails).To(Equal([]byte(`{"encrypted":{"foo":"bar"}}`)))
				Expect(receiver.ServiceID).To(Equal("fake-service-guid"))
				Expect(receiver.PlanID).To(Equal("fake-plan-guid"))
				Expect(receiver.SpaceGUID).To(Equal("fake-space-guid"))
				Expect(receiver.OrganizationGUID).To(Equal("fake-org-guid"))
				Expect(receiver.OperationType).To(Equal("fake-operation-type"))
				Expect(receiver.OperationID).To(Equal("fake-operation-guid"))
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
			Expect(r.Outputs).To(Equal(storage.JSONObject{"decrypted": map[string]any{"foo": "bar-2"}}))
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

		DescribeTable(
			"JSON parsing",
			func(input []byte) {
				encryptor.DecryptReturns(input, nil)

				r, err := store.GetServiceInstanceDetails("fake-id-1")
				Expect(err).NotTo(HaveOccurred())
				Expect(r.Outputs).To(Equal(storage.JSONObject(nil)))
			},
			Entry("null", []byte(`null`)),
			Entry("empty", []byte(``)),
			Entry("nil", []byte(nil)),
		)
	})

	Describe("GetServiceInstanceIDs", func() {
		When("database has service instances", func() {
			BeforeEach(func() {
				addFakeServiceInstanceDetails()
			})

			It("reads the ids database", func() {
				r, err := store.GetServiceInstancesIDs()
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(ConsistOf("fake-id-1", "fake-id-2", "fake-id-3"))
			})
		})

		When("nothing is found", func() {
			It("returns an empty slice", func() {
				r, err := store.GetServiceInstancesIDs()
				Expect(err).NotTo(HaveOccurred())
				Expect(r).To(BeEmpty())
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
		URL:              "fake-url-1",
		OtherDetails:     []byte(`{"foo":"bar-1"}`),
		ServiceID:        "fake-service-id-1",
		PlanID:           "fake-plan-id-1",
		SpaceGUID:        "fake-space-guid-1",
		OrganizationGUID: "fake-org-guid-1",
		OperationType:    "fake-operation-type-1",
		OperationID:      "fake-operation-id-1",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ServiceInstanceDetails{
		ID:               "fake-id-2",
		Name:             "fake-name-2",
		Location:         "fake-location-2",
		URL:              "fake-url-2",
		OtherDetails:     []byte(`{"foo":"bar-2"}`),
		ServiceID:        "fake-service-id-2",
		PlanID:           "fake-plan-id-2",
		SpaceGUID:        "fake-space-guid-2",
		OrganizationGUID: "fake-org-guid-2",
		OperationType:    "fake-operation-type-2",
		OperationID:      "fake-operation-id-2",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.ServiceInstanceDetails{
		ID:               "fake-id-3",
		Name:             "fake-name-3",
		Location:         "fake-location-3",
		URL:              "fake-url-3",
		OtherDetails:     []byte(`{"foo":"bar-3"}`),
		ServiceID:        "fake-service-id-3",
		PlanID:           "fake-plan-id-3",
		SpaceGUID:        "fake-space-guid-3",
		OrganizationGUID: "fake-org-guid-3",
		OperationType:    "fake-operation-type-3",
		OperationID:      "fake-operation-id-3",
	}).Error).NotTo(HaveOccurred())
}
