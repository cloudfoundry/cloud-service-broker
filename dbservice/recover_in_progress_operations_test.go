package dbservice

import (
	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/cloudfoundry/cloud-service-broker/v2/dbservice/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var _ = Describe("RecoverInProgressOperations()", func() {
	It("recovers the expected operations", func() {
		const (
			recoverID = "fake-id-to-recover"
			okID      = "fake-id-that-does-not-need-to-be-recovered"
		)

		// Setup
		db, err := gorm.Open(sqlite.Open(":memory:"), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(db.Migrator().CreateTable(&models.TerraformDeployment{})).To(Succeed())

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

		// Call the function
		logger := lagertest.NewTestLogger("test")
		recoverInProgressOperations(db, logger)

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
