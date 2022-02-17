package storage_test

import (
	"errors"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"

	"github.com/cloudfoundry/cloud-service-broker/db_service/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TerraformDeployments", func() {
	Describe("StoreTerraformDeployments", func() {
		It("creates the right object in the database", func() {
			err := store.StoreTerraformDeployment(storage.TerraformDeployment{
				ID:                   "fake-id",
				Workspace:            []byte("fake-workspace-stuff"),
				LastOperationType:    "create",
				LastOperationState:   "succeeded",
				LastOperationMessage: "yes!!",
			})
			Expect(err).NotTo(HaveOccurred())

			var receiver models.TerraformDeployment
			Expect(db.Find(&receiver).Error).NotTo(HaveOccurred())
			Expect(receiver.ID).To(Equal("fake-id"))
			Expect(receiver.Workspace).To(Equal([]byte(`{"encrypted":fake-workspace-stuff}`)))
			Expect(receiver.LastOperationType).To(Equal("create"))
			Expect(receiver.LastOperationState).To(Equal("succeeded"))
			Expect(receiver.LastOperationMessage).To(Equal("yes!!"))
		})

		When("encoding fails", func() {
			It("returns an error", func() {
				encryptor.EncryptReturns(nil, errors.New("bang"))

				err := store.StoreTerraformDeployment(storage.TerraformDeployment{})
				Expect(err).To(MatchError("error encoding workspace: bang"))
			})
		})

		When("details for the instance already exist in the database", func() {
			BeforeEach(func() {
				addFakeTerraformDeployments()
			})

			It("updates the existing record", func() {
				err := store.StoreTerraformDeployment(storage.TerraformDeployment{
					ID:                   "fake-id-2",
					Workspace:            []byte("fake-workspace-details"),
					LastOperationType:    "create",
					LastOperationState:   "succeeded",
					LastOperationMessage: "yes!!",
				})
				Expect(err).NotTo(HaveOccurred())

				var receiver models.TerraformDeployment
				Expect(db.Where(`id = "fake-id-2"`).Find(&receiver).Error).NotTo(HaveOccurred())
				Expect(receiver.ID).To(Equal("fake-id-2"))
				Expect(receiver.Workspace).To(Equal([]byte(`{"encrypted":fake-workspace-details}`)))
				Expect(receiver.LastOperationType).To(Equal("create"))
				Expect(receiver.LastOperationState).To(Equal("succeeded"))
				Expect(receiver.LastOperationMessage).To(Equal("yes!!"))
			})
		})
	})

	Describe("GetTerraformDeployments", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("reads the right object from the database", func() {
			r, err := store.GetTerraformDeployment("fake-id-2")
			Expect(err).NotTo(HaveOccurred())

			Expect(r.ID).To(Equal("fake-id-2"))
			Expect(r.Workspace).To(Equal([]byte(`{"decrypted":fake-workspace-2}`)))
			Expect(r.LastOperationType).To(Equal("update"))
			Expect(r.LastOperationState).To(Equal("failed"))
			Expect(r.LastOperationMessage).To(Equal("too bad"))
		})

		When("decoding fails", func() {
			It("returns an error", func() {
				encryptor.DecryptReturns(nil, errors.New("bang"))

				_, err := store.GetTerraformDeployment("fake-id-1")
				Expect(err).To(MatchError("error decoding workspace: bang"))
			})
		})

		When("nothing is found", func() {
			It("returns an error", func() {
				_, err := store.GetTerraformDeployment("not-there")
				Expect(err).To(MatchError("could not find terraform deployment: not-there"))
			})
		})
	})

	Describe("ExistsTerraformDeployments", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("reads the result from the database", func() {
			Expect(store.ExistsTerraformDeployment("not-there")).To(BeFalse())
			Expect(store.ExistsTerraformDeployment("also-not-there")).To(BeFalse())
			Expect(store.ExistsTerraformDeployment("fake-id-1")).To(BeTrue())
			Expect(store.ExistsTerraformDeployment("fake-id-2")).To(BeTrue())
			Expect(store.ExistsTerraformDeployment("fake-id-3")).To(BeTrue())
		})
	})

	Describe("DeleteTerraformDeployments", func() {
		BeforeEach(func() {
			addFakeTerraformDeployments()
		})

		It("deletes from the database", func() {
			Expect(store.ExistsTerraformDeployment("fake-id-3")).To(BeTrue())

			Expect(store.DeleteTerraformDeployment("fake-id-3")).NotTo(HaveOccurred())

			Expect(store.ExistsTerraformDeployment("fake-id-3")).To(BeFalse())
		})

		It("is idempotent", func() {
			Expect(store.DeleteTerraformDeployment("not-there")).NotTo(HaveOccurred())
		})
	})
})

func addFakeTerraformDeployments() {
	Expect(db.Create(&models.TerraformDeployment{
		ID:                   "fake-id-1",
		Workspace:            []byte("fake-workspace-1"),
		LastOperationType:    "create",
		LastOperationState:   "succeeded",
		LastOperationMessage: "amazing",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.TerraformDeployment{
		ID:                   "fake-id-2",
		Workspace:            []byte("fake-workspace-2"),
		LastOperationType:    "update",
		LastOperationState:   "failed",
		LastOperationMessage: "too bad",
	}).Error).NotTo(HaveOccurred())
	Expect(db.Create(&models.TerraformDeployment{
		ID:                   "fake-id-3",
		Workspace:            []byte("fake-workspace-3"),
		LastOperationType:    "update",
		LastOperationState:   "succeeded",
		LastOperationMessage: "great",
	}).Error).NotTo(HaveOccurred())
}
