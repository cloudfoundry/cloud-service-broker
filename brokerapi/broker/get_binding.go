package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v11/domain"

	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

// GetBinding fetches an existing service binding.
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
func (broker *ServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string, details domain.FetchBindingDetails) (domain.GetBindingSpec, error) {
	broker.Logger.Info("GetBinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
		"service_id":  details.ServiceID,
		"plan_id":     details.PlanID,
	})

	// check whether instance exists
	instanceExists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.GetBindingSpec{}, fmt.Errorf("error checking for existing instance: %w", err)
	}
	if !instanceExists {
		return domain.GetBindingSpec{}, ErrNotFound
	}

	// get instance details
	instanceRecord, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.GetBindingSpec{}, fmt.Errorf("error retrieving service instance details: %w", err)
	}

	// check whether request parameters (if not empty) match instance details
	if len(details.ServiceID) > 0 && details.ServiceID != instanceRecord.ServiceGUID {
		return domain.GetBindingSpec{}, ErrNotFound
	}
	if len(details.PlanID) > 0 && details.PlanID != instanceRecord.PlanGUID {
		return domain.GetBindingSpec{}, ErrNotFound
	}

	// check whether service plan is bindable
	serviceDefinition, _, err := broker.getDefinitionAndProvider(instanceRecord.ServiceGUID)
	if err != nil {
		return domain.GetBindingSpec{}, fmt.Errorf("error retrieving service definition: %w", err)
	}
	if !serviceDefinition.Bindable {
		return domain.GetBindingSpec{}, ErrBadRequest
	}

	// check whether binding exists
	//	 with the current implementation, bind is a synchroneous operation which waits for all resources to be created before binding credentials are stored
	//   therefore, we can assume the binding operation is completed if it exists at the store
	bindingExists, err := broker.store.ExistsServiceBindingCredentials(bindingID, instanceID)
	if err != nil {
		return domain.GetBindingSpec{}, fmt.Errorf("error checking for existing binding: %w", err)
	}
	if !bindingExists {
		return domain.GetBindingSpec{}, ErrNotFound
	}

	// get binding parameters
	params, err := broker.store.GetBindRequestDetails(bindingID, instanceID)
	if err != nil {
		return domain.GetBindingSpec{}, fmt.Errorf("error retrieving bind request details: %w", err)
	}

	// broker does not support Log Drain, Route Services, or Volume Mounts
	// broker does not support binding metadata
	// credentials are returned with synchronous bind request
	return domain.GetBindingSpec{
		Credentials:     nil,
		SyslogDrainURL:  "",
		RouteServiceURL: "",
		VolumeMounts:    nil,
		Parameters:      params,
		Endpoints:       nil,
		Metadata:        domain.BindingMetadata{},
	}, nil
}
