// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package brokers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/credstore"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/request"
	"github.com/pivotal-cf/brokerapi/v8"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

var (
	invalidUserInputMsg        = "User supplied parameters must be in the form of a valid JSON map."
	ErrInvalidUserInput        = apiresponses.NewFailureResponse(errors.New(invalidUserInputMsg), http.StatusBadRequest, "parsing-user-request")
	ErrGetInstancesUnsupported = apiresponses.NewFailureResponse(errors.New("the service_instances endpoint is unsupported"), http.StatusBadRequest, "unsupported")
	ErrGetBindingsUnsupported  = apiresponses.NewFailureResponse(errors.New("the service_bindings endpoint is unsupported"), http.StatusBadRequest, "unsupported")
	ErrNonUpdatableParameter   = apiresponses.NewFailureResponse(errors.New("attempt to update parameter that may result in service instance re-creation and data loss"), http.StatusBadRequest, "prohibited")
)

const credhubClientIdentifier = "csb"

// ServiceBroker is a brokerapi.ServiceBroker that can be used to generate an OSB compatible service broker.
type ServiceBroker struct {
	registry  broker.BrokerRegistry
	Credstore credstore.CredStore

	Logger lager.Logger
}

// New creates a ServiceBroker.
// Exactly one of ServiceBroker or error will be nil when returned.
func New(cfg *BrokerConfig, logger lager.Logger) (*ServiceBroker, error) {
	return &ServiceBroker{
		registry:  cfg.Registry,
		Credstore: cfg.Credstore,
		Logger:    logger,
	}, nil
}

// Services lists services in the broker's catalog.
// It is called through the `GET /v2/catalog` endpoint or the `cf marketplace` command.
func (broker *ServiceBroker) Services(ctx context.Context) ([]domain.Service, error) {
	var svcs []domain.Service

	enabledServices, err := broker.registry.GetEnabledServices()
	if err != nil {
		return nil, err
	}

	for _, service := range enabledServices {
		entry, err := service.CatalogEntry()
		if err != nil {
			return svcs, err
		}
		svcs = append(svcs, entry.ToPlain())
	}

	return svcs, nil
}

func (broker *ServiceBroker) getDefinitionAndProvider(serviceId string) (*broker.ServiceDefinition, broker.ServiceProvider, error) {
	defn, err := broker.registry.GetServiceById(serviceId)
	if err != nil {
		return nil, nil, err
	}

	providerBuilder := defn.ProviderBuilder(broker.Logger)
	return defn, providerBuilder, nil
}

// Provision creates a new instance of a service.
// It is bound to the `PUT /v2/service_instances/:instance_id` endpoint and can be called using the `cf create-service` command.
func (broker *ServiceBroker) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, clientSupportsAsync bool) (domain.ProvisionedServiceSpec, error) {
	broker.Logger.Info("Provisioning", correlation.ID(ctx), lager.Data{
		"instanceId":         instanceID,
		"accepts_incomplete": clientSupportsAsync,
		"details":            details,
	})

	// make sure that instance hasn't already been provisioned
	exists, err := db_service.ExistsServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("database error checking for existing instance: %s", err)
	}
	if exists {
		return domain.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
	}

	brokerService, serviceHelper, err := broker.getDefinitionAndProvider(details.ServiceID)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// verify the service exists and the plan exists
	plan, err := brokerService.GetPlanById(details.PlanID)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// verify async provisioning is allowed if it is required
	shouldProvisionAsync := serviceHelper.ProvisionsAsync()
	if shouldProvisionAsync && !clientSupportsAsync {
		return domain.ProvisionedServiceSpec{}, apiresponses.ErrAsyncRequired
	}

	// Give the user a better error message if they give us a bad request
	if !isValidOrEmptyJSON(details.GetRawParameters()) {
		return domain.ProvisionedServiceSpec{}, ErrInvalidUserInput
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	vars, err := brokerService.ProvisionVariables(instanceID, details, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// get instance details
	instanceDetails, err := serviceHelper.Provision(ctx, vars)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// save instance details
	instanceDetails.ServiceId = details.ServiceID
	instanceDetails.ID = instanceID
	instanceDetails.PlanId = details.PlanID
	instanceDetails.SpaceGuid = details.SpaceGUID
	instanceDetails.OrganizationGuid = details.OrganizationGUID

	err = db_service.CreateServiceInstanceDetails(ctx, &instanceDetails)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
	}

	// save provision request details
	pr := models.ProvisionRequestDetails{
		ServiceInstanceId: instanceID,
	}
	pr.SetRequestDetails(details.RawParameters)

	if err = db_service.CreateProvisionRequestDetails(ctx, &pr); err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	}

	return domain.ProvisionedServiceSpec{IsAsync: shouldProvisionAsync, DashboardURL: "", OperationData: instanceDetails.OperationId}, nil
}

// Deprovision destroys an existing instance of a service.
// It is bound to the `DELETE /v2/service_instances/:instance_id` endpoint and can be called using the `cf delete-service` command.
// If a deprovision is asynchronous, the returned DeprovisionServiceSpec will contain the operation ID for tracking its progress.
func (broker *ServiceBroker) Deprovision(ctx context.Context, instanceID string, details domain.DeprovisionDetails, clientSupportsAsync bool) (response domain.DeprovisionServiceSpec, err error) {
	broker.Logger.Info("Deprovisioning", correlation.ID(ctx), lager.Data{
		"instance_id":        instanceID,
		"accepts_incomplete": clientSupportsAsync,
		"details":            details,
	})

	// make sure that instance actually exists
	instance, err := db_service.GetServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return response, brokerapi.ErrInstanceDoesNotExist
	}

	brokerService, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceId)
	if err != nil {
		return response, err
	}

	// verify the service exists and the plan exists
	plan, err := brokerService.GetPlanById(details.PlanID)
	if err != nil {
		return response, err
	}

	// if async deprovisioning isn't allowed but this service needs it, throw an error
	if serviceProvider.DeprovisionsAsync() && !clientSupportsAsync {
		return response, brokerapi.ErrAsyncRequired
	}

	pr, err := db_service.GetProvisionRequestDetailsByInstanceId(ctx, instanceID)
	if err != nil {
		return response, fmt.Errorf("updating non-existent instanceid: %v", instanceID)
	}

	provisionDetails := domain.ProvisionDetails{
		ServiceID:     details.ServiceID,
		PlanID:        details.PlanID,
		RawParameters: pr.GetRequestDetails(),
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	vars, err := brokerService.ProvisionVariables(instanceID, provisionDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return response, err
	}

	operationId, err := serviceProvider.Deprovision(ctx, *instance, details, vars)
	if err != nil {
		return response, err
	}

	if operationId == nil {
		// soft-delete instance details from the db if this is a synchronous operation
		// if it's an async operation we can't delete from the db until we're sure delete succeeded, so this is
		// handled internally to LastOperation
		if err := db_service.DeleteServiceInstanceDetailsById(ctx, instanceID); err != nil {
			return response, fmt.Errorf("error deleting instance details from database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
		}
		return response, nil
	} else {
		response.IsAsync = true
		response.OperationData = *operationId

		instance.OperationType = models.DeprovisionOperationType
		instance.OperationId = *operationId
		if err := db_service.SaveServiceInstanceDetails(ctx, instance); err != nil {
			return response, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
		}
		return response, nil
	}
}

// Bind creates an account with credentials to access an instance of a service.
// It is bound to the `PUT /v2/service_instances/:instance_id/service_bindings/:binding_id` endpoint and can be called using the `cf bind-service` command.
func (broker *ServiceBroker) Bind(ctx context.Context, instanceID, bindingID string, details domain.BindDetails, clientSupportsAsync bool) (domain.Binding, error) {
	broker.Logger.Info("Binding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
		"details":     details,
	})

	// check for existing binding
	exists, err := db_service.ExistsServiceBindingCredentialsByServiceInstanceIdAndBindingId(ctx, instanceID, bindingID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error checking for existing binding: %w", err)
	}
	if exists {
		return domain.Binding{}, apiresponses.ErrBindingAlreadyExists
	}

	// get existing service instance details
	instanceRecord, err := db_service.GetServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error retrieving service instance details: %w", err)
	}

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(instanceRecord.ServiceId)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error retrieving service definition: %w", err)
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanById(details.PlanID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error getting service plan: %w", err)
	}

	// Give the user a better error message if they give us a bad request
	if !isValidOrEmptyJSON(details.GetRawParameters()) {
		return domain.Binding{}, ErrInvalidUserInput
	}

	// validate parameters meet the service's schema and merge the plan's vars with
	// the user's
	vars, err := serviceDefinition.BindVariables(*instanceRecord, bindingID, details, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error generating bind variabled: %w", err)
	}

	// create binding
	credsDetails, err := serviceProvider.Bind(ctx, vars)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error performing bind: %w", err)
	}

	// save binding to database
	newCreds := models.ServiceBindingCredentials{
		ServiceInstanceId: instanceID,
		BindingId:         bindingID,
		ServiceId:         details.ServiceID,
	}

	if err := newCreds.SetOtherDetails(credsDetails); err != nil {
		return domain.Binding{}, fmt.Errorf("error serializing credentials: %w. WARNING: these credentials cannot be unbound through cf. Please contact your operator for cleanup", err)
	}

	if err := db_service.CreateServiceBindingCredentials(ctx, &newCreds); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving credentials to database: %w. WARNING: these credentials cannot be unbound through cf. Please contact your operator for cleanup",
			err)
	}

	binding, err := serviceProvider.BuildInstanceCredentials(ctx, newCreds, *instanceRecord)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error building credentials: %w", err)
	}

	if broker.Credstore != nil {
		credentialName := getCredentialName(broker.getServiceName(serviceDefinition), bindingID)

		_, err := broker.Credstore.Put(credentialName, binding.Credentials)
		if err != nil {
			return domain.Binding{}, fmt.Errorf("bind failure: unable to put credentials in Credstore: %w", err)
		}

		_, err = broker.Credstore.AddPermission(credentialName, "mtls-app:"+details.AppGUID, []string{"read"})
		if err != nil {
			return domain.Binding{}, fmt.Errorf("bind failure: unable to add Credstore permissions to app: %w", err)
		}

		binding.Credentials = map[string]interface{}{
			"credhub-ref": credentialName,
		}
	}

	return *binding, nil
}

func (broker *ServiceBroker) getServiceName(def *broker.ServiceDefinition) string {
	return def.Name
}

func getCredentialName(serviceName, bindingID string) string {
	return fmt.Sprintf("/c/%s/%s/%s/secrets-and-services", credhubClientIdentifier, serviceName, bindingID)
}

// GetBinding fetches an existing service binding.
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) GetBinding(ctx context.Context, instanceID, bindingID string, details domain.FetchBindingDetails) (domain.GetBindingSpec, error) {
	broker.Logger.Info("GetBinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
	})

	return domain.GetBindingSpec{}, ErrGetBindingsUnsupported
}

// GetInstance fetches information about a service instance
// GET /v2/service_instances/{instance_id}
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) GetInstance(ctx context.Context, instanceID string, details domain.FetchInstanceDetails) (domain.GetInstanceDetailsSpec, error) {
	broker.Logger.Info("GetInstance", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
	})

	return domain.GetInstanceDetailsSpec{}, ErrGetInstancesUnsupported
}

// LastBindingOperation fetches last operation state for a service binding.
// GET /v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation
//
// NOTE: This functionality is not implemented.
func (broker *ServiceBroker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	broker.Logger.Info("LastBindingOperation", correlation.ID(ctx), lager.Data{
		"instance_id":    instanceID,
		"binding_id":     bindingID,
		"plan_id":        details.PlanID,
		"service_id":     details.ServiceID,
		"operation_data": details.OperationData,
	})

	return domain.LastOperation{}, apiresponses.ErrAsyncRequired
}

// Unbind destroys an account and credentials with access to an instance of a service.
// It is bound to the `DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id` endpoint and can be called using the `cf unbind-service` command.
func (broker *ServiceBroker) Unbind(ctx context.Context, instanceID, bindingID string, details domain.UnbindDetails, asyncSupported bool) (domain.UnbindSpec, error) {
	broker.Logger.Info("Unbinding", correlation.ID(ctx), lager.Data{
		"instance_id": instanceID,
		"binding_id":  bindingID,
		"details":     details,
	})

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(details.ServiceID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	// validate existence of binding
	existingBinding, err := db_service.GetServiceBindingCredentialsByServiceInstanceIdAndBindingId(ctx, instanceID, bindingID)
	if err != nil {
		return domain.UnbindSpec{}, apiresponses.ErrBindingDoesNotExist
	}

	// get existing service instance details
	instance, err := db_service.GetServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error retrieving service instance details: %s", err)
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanById(details.PlanID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	pr, err := db_service.GetProvisionRequestDetailsByInstanceId(ctx, instanceID)
	if err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("updating non-existent instanceid: %v", instanceID)
	}

	// validate parameters meet the service's schema and merge the plan's vars with
	// the user's
	bindDetails := domain.BindDetails{
		PlanID:        details.PlanID,
		ServiceID:     details.ServiceID,
		RawParameters: pr.GetRequestDetails(),
	}

	vars, err := serviceDefinition.BindVariables(*instance, bindingID, bindDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	// remove binding from service provider
	if err := serviceProvider.Unbind(ctx, *instance, *existingBinding, vars); err != nil {
		return domain.UnbindSpec{}, err
	}

	if broker.Credstore != nil {
		credentialName := getCredentialName(broker.getServiceName(serviceDefinition), bindingID)

		err = broker.Credstore.DeletePermission(credentialName)
		if err != nil {
			broker.Logger.Error(fmt.Sprintf("fail to delete permissions on the key %s", credentialName), err)
		}

		err := broker.Credstore.Delete(credentialName)
		if err != nil {
			return domain.UnbindSpec{}, err
		}
	}

	// remove binding from database
	if err := db_service.DeleteServiceBindingCredentials(ctx, existingBinding); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup", err)
	}

	return domain.UnbindSpec{}, nil
}

// LastOperation fetches last operation state for a service instance.
// It is bound to the `GET /v2/service_instances/:instance_id/last_operation` endpoint.
// It is called by `cf create-service` or `cf delete-service` if the operation was asynchronous.
func (broker *ServiceBroker) LastOperation(ctx context.Context, instanceID string, details domain.PollDetails) (domain.LastOperation, error) {
	broker.Logger.Info("Last Operation", correlation.ID(ctx), lager.Data{
		"instance_id":    instanceID,
		"plan_id":        details.PlanID,
		"service_id":     details.ServiceID,
		"operation_data": details.OperationData,
	})

	instance, err := db_service.GetServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return domain.LastOperation{}, apiresponses.ErrInstanceDoesNotExist
	}

	_, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceId)
	if err != nil {
		return domain.LastOperation{}, err
	}

	isAsyncService := serviceProvider.ProvisionsAsync() || serviceProvider.DeprovisionsAsync()
	if !isAsyncService {
		return domain.LastOperation{}, apiresponses.ErrAsyncRequired
	}

	lastOperationType := instance.OperationType

	done, message, err := serviceProvider.PollInstance(ctx, *instance)

	if err != nil {
		return domain.LastOperation{State: domain.Failed, Description: err.Error()}, nil
	}

	if !done {
		return domain.LastOperation{State: domain.InProgress, Description: message}, nil
	}

	// the instance may have been invalidated, so we pass its primary key rather than the
	// instance directly.
	updateErr := broker.updateStateOnOperationCompletion(ctx, serviceProvider, lastOperationType, instanceID)
	return domain.LastOperation{State: domain.Succeeded, Description: message}, updateErr
}

// updateStateOnOperationCompletion handles updating/cleaning-up resources that need to be changed
// once lastOperation finishes successfully.
func (broker *ServiceBroker) updateStateOnOperationCompletion(ctx context.Context, service broker.ServiceProvider, lastOperationType, instanceID string) error {
	if lastOperationType == models.DeprovisionOperationType {
		if err := db_service.DeleteServiceInstanceDetailsById(ctx, instanceID); err != nil {
			return fmt.Errorf("error deleting instance details from database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
		}

		return nil
	}

	// If the operation was not a delete, clear out the ID and type and update
	// any changed (or finalized) state like IP addresses, selflinks, etc.
	details, err := db_service.GetServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("error getting instance details from database %v", err)
	}

	if err := service.UpdateInstanceDetails(ctx, details); err != nil {
		return fmt.Errorf("error getting new instance details from GCP: %v", err)
	}

	details.OperationId = ""
	details.OperationType = models.ClearOperationType
	if err := db_service.SaveServiceInstanceDetails(ctx, details); err != nil {
		return fmt.Errorf("error saving instance details to database %v", err)
	}

	return nil
}

// Update a service instance plan.
// This functionality is not implemented and will return an error indicating that plan changes are not supported.
func (broker *ServiceBroker) Update(ctx context.Context, instanceID string, details domain.UpdateDetails, asyncAllowed bool) (response domain.UpdateServiceSpec, err error) {
	broker.Logger.Info("Updating", correlation.ID(ctx), lager.Data{
		"instance_id":        instanceID,
		"accepts_incomplete": asyncAllowed,
		"details":            details,
	})

	// make sure that instance actually exists
	instance, err := db_service.GetServiceInstanceDetailsById(ctx, instanceID)
	if err != nil {
		return response, brokerapi.ErrInstanceDoesNotExist
	}

	brokerService, serviceHelper, err := broker.getDefinitionAndProvider(instance.ServiceId)
	if err != nil {
		return response, err
	}

	// verify the service exists and the plan exists
	plan, err := brokerService.GetPlanById(details.PlanID)
	if err != nil {
		return response, err
	}

	// verify async provisioning is allowed if it is required
	shouldProvisionAsync := serviceHelper.ProvisionsAsync()
	if shouldProvisionAsync && !asyncAllowed {
		return response, brokerapi.ErrAsyncRequired
	}

	// Give the user a better error message if they give us a bad request
	if !isValidOrEmptyJSON(details.GetRawParameters()) {
		return response, ErrInvalidUserInput
	}

	allowUpdate, err := brokerService.AllowedUpdate(details)

	if err != nil {
		return response, err
	}

	if !allowUpdate {
		return response, ErrNonUpdatableParameter
	}

	pr, err := db_service.GetProvisionRequestDetailsByInstanceId(ctx, instanceID)
	if err != nil {
		return response, fmt.Errorf("updating non-existent instanceid: %v", instanceID)
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	vars, err := brokerService.UpdateVariables(instanceID, details, pr.GetRequestDetails(), *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return response, err
	}

	// get instance details
	newInstanceDetails, err := serviceHelper.Update(ctx, vars)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	// save instance details

	instance.PlanId = newInstanceDetails.PlanId

	err = db_service.SaveServiceInstanceDetails(ctx, instance)
	if err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
	}

	// save provision request details
	// pr := models.ProvisionRequestDetails{
	// 	ServiceInstanceId: instanceID,
	// 	RequestDetails:    string(details.RawParameters),
	// }
	// if err = db_service.SaveProvisionRequestDetails(ctx, &pr); err != nil {
	// 	return brokerapi.UpdateServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	// }

	response.IsAsync = shouldProvisionAsync
	response.DashboardURL = ""
	response.OperationData = newInstanceDetails.OperationId

	return response, nil
}

func isValidOrEmptyJSON(msg json.RawMessage) bool {
	return msg == nil || len(msg) == 0 || json.Valid(msg)
}
