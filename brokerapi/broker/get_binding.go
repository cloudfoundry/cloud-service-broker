package broker

import (
	"context"
	"errors"
	"net/http"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v9/domain"
	"github.com/pivotal-cf/brokerapi/v9/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/utils/correlation"
)

// GetBinding fetches an existing service binding.
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string, _ domain.FetchBindingDetails) (domain.GetBindingSpec, error) {
	broker.Logger.Info("GetBinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
	})

	return domain.GetBindingSpec{}, apiresponses.NewFailureResponse(
		errors.New("the service_bindings endpoint is unsupported"),
		http.StatusBadRequest,
		"unsupported")
}
