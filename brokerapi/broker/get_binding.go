package broker

import (
	"context"
	"errors"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

// GetBinding fetches an existing service binding.
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string, details domain.FetchBindingDetails) (domain.GetBindingSpec, error) {
	broker.Logger.Info("GetBinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
	})

	return domain.GetBindingSpec{}, apiresponses.NewFailureResponse(
		errors.New("the service_bindings endpoint is unsupported"),
		http.StatusBadRequest,
		"unsupported")
}
