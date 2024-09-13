package storage_test

import (
	"os"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	recoverID = "fake-id-to-recover"
	okID      = "fake-id-that-does-not-need-to-be-recovered"
)

var _ = Describe("RecoverInProgressOperations()", func() {
	BeforeEach(func() {
		// Setup
		Expect(db.Create(&models.TerraformDeployment{
			ID:                   recoverID,
			LastOperationType:    "fake-type",
			LastOperationState:   "in progress",
			LastOperationMessage: "fake-type in progress",
		}).Error).To(Succeed())
		Expect(db.Create(&models.TerraformDeployment{
			ID:                   okID,
			LastOperationType:    "fake-type",
			LastOperationState:   "succeeded",
			LastOperationMessage: "fake-type succeeded",
		}).Error).To(Succeed())
		var rowCount int64
		db.Model(&models.TerraformDeployment{}).Count(&rowCount)
		Expect(rowCount).To(BeNumerically("==", 2))

		logger = lagertest.NewTestLogger("test")
		store = storage.New(db, encryptor)
	})

	When("running as a cf app", func() {
		It("recovers the expected operations", func() {
			os.Setenv("CF_INSTANCE_GUID", "something") // The presence of this variable means we are running as an App
			defer os.Unsetenv("CF_INSTANCE_GUID")

			// Call the function
			Expect(store.RecoverInProgressOperations(logger)).To(Succeed())

			// Behaviors
			By("marking the in-progress operation as failed")
			var r1 models.TerraformDeployment
			Expect(db.Where("id = ?", recoverID).First(&r1).Error).To(Succeed())
			Expect(r1.LastOperationState).To(Equal("failed"))
			Expect(r1.LastOperationMessage).To(Equal("the broker restarted while the operation was in progress"))

			By("no updating other operations")
			var r2 models.TerraformDeployment
			Expect(db.Where("id = ?", okID).First(&r2).Error).To(Succeed())
			Expect(r2.LastOperationState).To(Equal("succeeded"))
			Expect(r2.LastOperationMessage).To(Equal("fake-type succeeded"))

			By("logging the expected message")
			Expect(logger.Buffer().Contents()).To(SatisfyAll(
				ContainSubstring(`"message":"test.recover-in-progress-operations.mark-as-failed"`),
				ContainSubstring(`"workspace_id":"fake-id-to-recover"`),
			))
		})
	})

	When("running on a VM", func() {
		It("recovers the expected operations", func() {
			// When running on a VM there will be a lockfile and record in the db
			Expect(store.WriteLockFile(recoverID)).To(Succeed())

			// Call the function
			Expect(store.RecoverInProgressOperations(logger)).To(Succeed())

			// Behaviors
			By("marking the in-progress operation as failed")
			var r1 models.TerraformDeployment
			Expect(db.Where("id = ?", recoverID).First(&r1).Error).To(Succeed())
			Expect(r1.LastOperationState).To(Equal("failed"))
			Expect(r1.LastOperationMessage).To(Equal("the broker restarted while the operation was in progress"))

			By("no updating other operations")
			var r2 models.TerraformDeployment
			Expect(db.Where("id = ?", okID).First(&r2).Error).To(Succeed())
			Expect(r2.LastOperationState).To(Equal("succeeded"))
			Expect(r2.LastOperationMessage).To(Equal("fake-type succeeded"))

			By("logging the expected message")
			Expect(logger.Buffer().Contents()).To(SatisfyAll(
				ContainSubstring(`"message":"test.recover-in-progress-operations.mark-as-failed"`),
				ContainSubstring(`"workspace_id":"fake-id-to-recover"`),
			))
		})
	})
})
