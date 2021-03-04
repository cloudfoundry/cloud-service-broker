package db_service

import (
	"context"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

func TestSqlDatastore_GetsProvisionRequestDetailsByInstanceId(t *testing.T) {
	ds := newInMemoryDatastore(t)
	_, instance := createProvisionRequestDetailsInstance()
	testCtx := context.Background()

	if _, err := ds.GetProvisionRequestDetailsById(testCtx, instance.ID); err != gorm.ErrRecordNotFound {
		t.Errorf("Expected an ErrRecordNotFound trying to get non-existing record got %v", err)
	}

	beforeCreation := time.Now()
	if err := ds.CreateProvisionRequestDetails(testCtx, &instance); err != nil {
		t.Errorf("Expected to be able to create the item %#v, got error: %s", instance, err)
	}
	afterCreation := time.Now()

	// after creation we should be able to get the item
	ret, err := ds.GetProvisionRequestDetailsByInstanceId(testCtx, instance.ServiceInstanceId)
	if err != nil {
		t.Errorf("Expected no error trying to get saved item, got: %v", err)
	}

	if ret.CreatedAt.Before(beforeCreation) || ret.CreatedAt.After(afterCreation) {
		t.Errorf("Expected creation time to be between  %v and %v got %v", beforeCreation, afterCreation, ret.CreatedAt)
	}

	if !ret.UpdatedAt.Equal(ret.CreatedAt) {
		t.Errorf("Expected initial update time to equal creation time, but got update: %v, create: %v", ret.UpdatedAt, ret.CreatedAt)
	}

	// Ensure non-gorm fields were deserialized correctly
	ensureProvisionRequestDetailsFieldsMatch(t, &instance, ret)
}
