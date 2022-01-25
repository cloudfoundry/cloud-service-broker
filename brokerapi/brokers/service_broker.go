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
	"sort"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/cloud-service-broker/db_service/models"
	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/credstore"
	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/varcontext"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/correlation"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils/request"
	"github.com/pivotal-cf/brokerapi/v8"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
	"github.com/spf13/viper"
)

const (
	credhubClientIdentifier          = "csb"
	DisableRequestPropertyValidation = "request.property.validation.disabled"
)

var (
	invalidUserInputMsg        = "User supplied parameters must be in the form of a valid JSON map."
	ErrInvalidUserInput        = apiresponses.NewFailureResponse(errors.New(invalidUserInputMsg), http.StatusBadRequest, "parsing-user-request")
	ErrGetInstancesUnsupported = apiresponses.NewFailureResponse(errors.New("the service_instances endpoint is unsupported"), http.StatusBadRequest, "unsupported")
	ErrGetBindingsUnsupported  = apiresponses.NewFailureResponse(errors.New("the service_bindings endpoint is unsupported"), http.StatusBadRequest, "unsupported")
	ErrNonUpdatableParameter   = apiresponses.NewFailureResponse(errors.New("attempt to update parameter that may result in service instance re-creation and data loss"), http.StatusBadRequest, "prohibited")
)

func init() {
	viper.BindEnv(DisableRequestPropertyValidation, "CSB_DISABLE_REQUEST_PROPERTY_VALIDATION")
}

// ServiceBroker is a brokerapi.ServiceBroker that can be used to generate an OSB compatible service broker.
type ServiceBroker struct {
	registry  broker.BrokerRegistry
	Credstore credstore.CredStore

	Logger lager.Logger
	store  Storage
}

// New creates a ServiceBroker.
// Exactly one of ServiceBroker or error will be nil when returned.
func New(cfg *BrokerConfig, logger lager.Logger, store Storage) (*ServiceBroker, error) {
	return &ServiceBroker{
		registry:  cfg.Registry,
		Credstore: cfg.Credstore,
		Logger:    logger,
		store:     store,
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
		entry := service.CatalogEntry()
		svcs = append(svcs, entry.ToPlain())
	}

	return svcs, nil
}

func (broker *ServiceBroker) getDefinitionAndProvider(serviceId string) (*broker.ServiceDefinition, broker.ServiceProvider, error) {
	defn, err := broker.registry.GetServiceById(serviceId)
	if err != nil {
		return nil, nil, err
	}

	providerBuilder := defn.ProviderBuilder(broker.Logger, broker.store)
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
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("database error checking for existing instance: %s", err)
	case exists:
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
	if err := validateParameters(details.GetRawParameters(), brokerService.ProvisionInputVariables, brokerService.ImportInputVariables); err != nil {
		return domain.ProvisionedServiceSpec{}, err
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
	instanceDetails.ServiceGUID = details.ServiceID
	instanceDetails.GUID = instanceID
	instanceDetails.PlanGUID = details.PlanID
	instanceDetails.SpaceGUID = details.SpaceGUID
	instanceDetails.OrganizationGUID = details.OrganizationGUID

	if err := broker.store.StoreServiceInstanceDetails(instanceDetails); err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
	}

	// save provision request details
	if err := broker.store.StoreProvisionRequestDetails(instanceID, details.RawParameters); err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	}

	return domain.ProvisionedServiceSpec{IsAsync: shouldProvisionAsync, DashboardURL: "", OperationData: instanceDetails.OperationGUID}, nil
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
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return response, err
	case !exists:
		return response, apiresponses.ErrInstanceDoesNotExist
	}

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return response, err
	}

	brokerService, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
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

	rawParameters, err := broker.store.GetProvisionRequestDetails(instanceID)
	if err != nil {
		return response, fmt.Errorf("error retrieving provision request details for %q: %w", instanceID, err)
	}

	provisionDetails := domain.ProvisionDetails{
		ServiceID:     details.ServiceID,
		PlanID:        details.PlanID,
		RawParameters: rawParameters,
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	vars, err := brokerService.ProvisionVariables(instanceID, provisionDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return response, err
	}

	operationId, err := serviceProvider.Deprovision(ctx, instance.GUID, details, vars)
	if err != nil {
		return response, err
	}

	if operationId == nil {
		// soft-delete instance details from the db if this is a synchronous operation
		// if it's an async operation we can't delete from the db until we're sure delete succeeded, so this is
		// handled internally to LastOperation
		if err := broker.store.DeleteServiceInstanceDetails(instanceID); err != nil {
			return response, fmt.Errorf("error deleting instance details from database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
		}
		if err := broker.store.DeleteProvisionRequestDetails(instanceID); err != nil {
			return response, fmt.Errorf("error deleting provision request details from the database: %w", err)
		}
		return response, nil
	}

	response.IsAsync = true
	response.OperationData = *operationId

	instance.OperationType = models.DeprovisionOperationType
	instance.OperationGUID = *operationId
	if err := broker.store.StoreServiceInstanceDetails(instance); err != nil {
		return response, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
	}
	return response, nil

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
	exists, err := broker.store.ExistsServiceBindingCredentials(bindingID, instanceID)
	switch {
	case err != nil:
		return domain.Binding{}, fmt.Errorf("error checking for existing binding: %w", err)
	case exists:
		return domain.Binding{}, apiresponses.ErrBindingAlreadyExists
	}

	// get existing service instance details
	instanceRecord, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error retrieving service instance details: %w", err)
	}

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(instanceRecord.ServiceGUID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error retrieving service definition: %w", err)
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanById(details.PlanID)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error getting service plan: %w", err)
	}

	// Give the user a better error message if they give us a bad request
	if err := validateParameters(details.GetRawParameters(), serviceDefinition.BindInputVariables, nil); err != nil {
		return domain.Binding{}, err
	}

	// validate parameters meet the service's schema and merge the plan's vars with
	// the user's
	vars, err := serviceDefinition.BindVariables(instanceRecord, bindingID, details, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error generating bind variables: %w", err)
	}

	// create binding
	credsDetails, err := serviceProvider.Bind(ctx, vars)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("error performing bind: %w", err)
	}

	// save binding to database
	newCreds := storage.ServiceBindingCredentials{
		ServiceInstanceGUID: instanceID,
		BindingGUID:         bindingID,
		ServiceGUID:         details.ServiceID,
		Credentials:         credsDetails,
	}
	if err := broker.store.CreateServiceBindingCredentials(newCreds); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving credentials to database: %w. WARNING: these credentials cannot be unbound through cf. Please contact your operator for cleanup",
			err)
	}

	bindRequest := storage.BindRequestDetails{
		ServiceInstanceGUID: instanceID,
		ServiceBindingGUID:  bindingID,
		RequestDetails:      details.RawParameters,
	}

	if err := broker.store.StoreBindRequestDetails(bindRequest); err != nil {
		return domain.Binding{}, fmt.Errorf("error saving bind request details to database: %s. Unbind operations will not be able to complete", err)
	}

	binding, err := serviceProvider.BuildInstanceCredentials(ctx, newCreds.Credentials, instanceRecord.Outputs)
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
	exists, err := broker.store.ExistsServiceBindingCredentials(bindingID, instanceID)
	switch {
	case err != nil:
		return domain.UnbindSpec{}, fmt.Errorf("error locating service binding: %w", err)
	case !exists:
		return domain.UnbindSpec{}, apiresponses.ErrBindingDoesNotExist
	}

	// get existing service instance details
	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error retrieving service instance details: %s", err)
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanById(details.PlanID)
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	rawParameters, err := broker.store.GetBindRequestDetails(bindingID, instanceID)
	if err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error retrieving bind request details for %q: %w", instanceID, err)
	}

	bindDetails := domain.BindDetails{
		PlanID:        details.PlanID,
		ServiceID:     details.ServiceID,
		RawParameters: rawParameters,
	}

	vars, err := serviceDefinition.BindVariables(instance, bindingID, bindDetails, plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.UnbindSpec{}, err
	}

	// remove binding from service provider
	if err := serviceProvider.Unbind(ctx, instanceID, bindingID, vars); err != nil {
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
	if err := broker.store.DeleteServiceBindingCredentials(bindingID, instanceID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting credentials from database: %s. WARNING: these credentials will remain visible in cf. Contact your operator for cleanup", err)
	}
	if err := broker.store.DeleteBindRequestDetails(bindingID, instanceID); err != nil {
		return domain.UnbindSpec{}, fmt.Errorf("error soft-deleting bind request details from database: %s", err)
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

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return domain.LastOperation{}, apiresponses.ErrInstanceDoesNotExist
	}

	_, serviceProvider, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
	if err != nil {
		return domain.LastOperation{}, err
	}

	isAsyncService := serviceProvider.ProvisionsAsync() || serviceProvider.DeprovisionsAsync()
	if !isAsyncService {
		return domain.LastOperation{}, apiresponses.ErrAsyncRequired
	}

	lastOperationType := instance.OperationType

	done, message, err := serviceProvider.PollInstance(ctx, instance.GUID)

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
		if err := broker.store.DeleteServiceInstanceDetails(instanceID); err != nil {
			return fmt.Errorf("error deleting instance details from database: %s. WARNING: this instance will remain visible in cf. Contact your operator for cleanup", err)
		}
		if err := broker.store.DeleteProvisionRequestDetails(instanceID); err != nil {
			return fmt.Errorf("error deleting provision request details from the database: %w", err)
		}

		return nil
	}

	// If the operation was not a delete, clear out the ID and type and update
	// any changed (or finalized) state like IP addresses, selflinks, etc.
	details, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return fmt.Errorf("error getting instance details from database %v", err)
	}

	outs, err := service.GetTerraformOutputs(ctx, details.GUID)
	if err != nil {
		return fmt.Errorf("error getting new instance details from GCP: %v", err)
	}

	details.Outputs = outs
	details.OperationGUID = ""
	details.OperationType = models.ClearOperationType
	if err := broker.store.StoreServiceInstanceDetails(details); err != nil {
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
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return response, err
	case !exists:
		return response, apiresponses.ErrInstanceDoesNotExist
	}

	instance, err := broker.store.GetServiceInstanceDetails(instanceID)
	if err != nil {
		return response, err
	}

	brokerService, serviceHelper, err := broker.getDefinitionAndProvider(instance.ServiceGUID)
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
		return response, apiresponses.ErrAsyncRequired
	}

	// Give the user a better error message if they give us a bad request
	if err := validateParameters(details.GetRawParameters(), brokerService.ProvisionInputVariables, nil); err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	allowUpdate, err := brokerService.AllowedUpdate(details)

	if err != nil {
		return response, err
	}

	if !allowUpdate {
		return response, ErrNonUpdatableParameter
	}

	provisionDetails, err := broker.store.GetProvisionRequestDetails(instanceID)
	if err != nil {
		return response, fmt.Errorf("error retrieving provision request details for %q: %w", instanceID, err)
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	mergedDetails, err := mergeJSON(provisionDetails, details.GetRawParameters())
	if err != nil {
		return response, fmt.Errorf("error merging update and provision details: %w", err)
	}

	vars, err := brokerService.UpdateVariables(instanceID, details, mergedDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return response, err
	}

	// get instance details
	newInstanceDetails, err := serviceHelper.Update(ctx, vars)
	if err != nil {
		return domain.UpdateServiceSpec{}, err
	}

	// save instance details
	instance.PlanGUID = newInstanceDetails.PlanId
	if err := broker.store.StoreServiceInstanceDetails(instance); err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
	}

	// save provision request details
	if err = broker.store.StoreProvisionRequestDetails(instanceID, mergedDetails); err != nil {
		return domain.UpdateServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	}

	response.IsAsync = shouldProvisionAsync
	response.DashboardURL = ""
	response.OperationData = newInstanceDetails.OperationId

	return response, nil
}

func validateParameters(rawParams json.RawMessage, validUserInputFields []broker.BrokerVariable, validImportFields []broker.ImportVariable) error {
	if len(rawParams) == 0 {
		return nil
	}

	var params map[string]interface{}
	if err := json.Unmarshal(rawParams, &params); err != nil {
		return ErrInvalidUserInput
	}

	// As this is a new check we have feature-flagged it so that it can easily be disabled
	// if it causes problems.
	if viper.GetBool(DisableRequestPropertyValidation) {
		return nil
	}

	validParams := make(map[string]struct{})
	for _, field := range validUserInputFields {
		validParams[field.FieldName] = struct{}{}
	}
	for _, field := range validImportFields {
		validParams[field.Name] = struct{}{}
	}
	var invalidParams []string
	for k := range params {
		if _, ok := validParams[k]; !ok {
			invalidParams = append(invalidParams, k)
		}
	}

	if len(invalidParams) == 0 {
		return nil
	}

	sort.Strings(invalidParams)
	return fmt.Errorf("additional properties are not allowed: %s", strings.Join(invalidParams, ", "))
}

func mergeJSON(previousParams, newParams json.RawMessage) (json.RawMessage, error) {
	vc, err := varcontext.Builder().
		MergeJsonObject(previousParams).
		MergeJsonObject(newParams).
		Build()
	if err != nil {
		return nil, err
	}

	return vc.ToJson()
}
