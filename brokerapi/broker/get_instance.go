package broker

import (
	"context"
	"errors"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

// GetInstance fetches information about a service instance
// GET /v2/service_instances/{instance_id}
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) GetInstance(ctx context.Context, instanceID string, details domain.FetchInstanceDetails) (domain.GetInstanceDetailsSpec, error) {
	broker.Logger.Info("GetInstance", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
	})

	return domain.GetInstanceDetailsSpec{}, apiresponses.NewFailureResponse(
		errors.New("the service_instances endpoint is unsupported"),
		http.StatusBadRequest,
		"unsupported")
}
