package db_service

import (
	"context"

	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
)

func GetProvisionRequestDetailsByInstanceId(ctx context.Context, instanceId string) (*models.ProvisionRequestDetails, error) {
	return defaultDatastore().GetProvisionRequestDetailsByInstanceId(ctx, instanceId)
}
func (ds *SqlDatastore) GetProvisionRequestDetailsByInstanceId(ctx context.Context, instanceId string) (*models.ProvisionRequestDetails, error) {
	record := models.ProvisionRequestDetails{}
	if err := ds.db.Where("service_instance_id = ?", instanceId).First(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}
