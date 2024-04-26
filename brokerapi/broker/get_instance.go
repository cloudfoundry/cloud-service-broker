package broker

import (
	"context"
	"errors"
	"net/http"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/v3/utils/correlation"
)

// GetInstance fetches information about a service instance
// GET /v2/service_instances/{instance_id}
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) GetInstance(ctx context.Context, instanceID string, _ domain.FetchInstanceDetails) (domain.GetInstanceDetailsSpec, error) {
	broker.Logger.Info("GetInstance", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
	})

	return domain.GetInstanceDetailsSpec{}, apiresponses.NewFailureResponse(
		errors.New("the service_instances endpoint is unsupported"),
		http.StatusBadRequest,
		"unsupported")
}
