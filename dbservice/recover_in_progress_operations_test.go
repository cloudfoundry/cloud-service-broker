package dbservice

import (
	"strings"
	"testing"

	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/cloudfoundry/cloud-service-broker/dbservice/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRecoverInProgressOperations(t *testing.T) {
	// Setup
	db, err := gorm.Open(sqlite.Open(":memory:"), nil)
	if err != nil {
		t.Errorf("failed to create test database: %s", err)
	}

	if err = db.Migrator().CreateTable(&models.TerraformDeployment{}); err != nil {
		t.Errorf("failed to create test table: %s", err)
	}

	const recoverID = "fake-id-to-recover"
	err = db.Create(&models.TerraformDeployment{
		ID:                   recoverID,
		LastOperationType:    "fake-type",
		LastOperationState:   "in progress",
		LastOperationMessage: "fake-type in progress",
	}).Error
	if err != nil {
		t.Errorf("failed to create test database data: %s", err)
	}
	const okID = "fake-id-that-does-not-need-to-be-recovered"
	err = db.Create(&models.TerraformDeployment{
		ID:                   okID,
		LastOperationType:    "fake-type",
		LastOperationState:   "succeeded",
		LastOperationMessage: "fake-type succeeded",
	}).Error
	if err != nil {
		t.Errorf("failed to create test database data: %s", err)
	}

	// Call the function
	logger := lagertest.NewTestLogger("test")
	recoverInProgressOperations(db, logger)

	// It marks the in-progress operation as failed
	var r1 models.TerraformDeployment
	err = db.Where("id = ?", recoverID).First(&r1).Error
	if err != nil {
		t.Errorf("failed to load updated test data: %s", err)
	}

	const expState = "failed"
	if r1.LastOperationState != expState {
		t.Errorf("LastOperationState, expected %q, got %q", expState, r1.LastOperationState)
	}

	const expMessage = "the broker restarted while the operation was in progress"
	if r1.LastOperationMessage != expMessage {
		t.Errorf("LastOperationMessage, expected %q, got %q", expMessage, r1.LastOperationMessage)
	}

	// It does not update other operations
	var r2 models.TerraformDeployment
	err = db.Where("id = ?", okID).First(&r2).Error
	if err != nil {
		t.Errorf("failed to load updated test data: %s", err)
	}
	if r2.LastOperationState != "succeeded" || r2.LastOperationMessage != "fake-type succeeded" {
		t.Error("row corruption")
	}

	// It logs the expected message
	const expLog1 = `"message":"test.recover-in-progress-operations.mark-as-failed"`
	const expLog2 = `"workspace_id":"fake-id-to-recover"`
	logMessage := string(logger.Buffer().Contents())
	if !strings.Contains(logMessage, expLog1) || !strings.Contains(logMessage, expLog2) {
		t.Errorf("log, expected to contain %q and %q, got %q", expLog1, expLog2, logMessage)
	}
}
