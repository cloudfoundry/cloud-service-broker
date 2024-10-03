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

package broker

import (
	"context"
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/paramparser"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/request"
)

// Provision creates a new instance of a service.
// It is bound to the `PUT /v2/service_instances/:instance_id` endpoint and can be called using the `cf create-service` command.
func (broker *ServiceBroker) Provision(ctx context.Context, instanceID string, details domain.ProvisionDetails, clientSupportsAsync bool) (domain.ProvisionedServiceSpec, error) {
	broker.Logger.Info("Provisioning", correlation.ID(ctx), lager.Data{
		"instanceId":         instanceID,
		"accepts_incomplete": clientSupportsAsync,
		"details":            details,
	})

	if !clientSupportsAsync {
		return domain.ProvisionedServiceSpec{}, apiresponses.ErrAsyncRequired
	}

	// make sure that instance hasn't already been provisioned
	exists, err := broker.store.ExistsServiceInstanceDetails(instanceID)
	switch {
	case err != nil:
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("database error checking for existing instance: %s", err)
	case exists:
		return domain.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
	}

	parsedDetails, err := paramparser.ParseProvisionDetails(details)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, ErrInvalidUserInput
	}

	serviceDefinition, serviceProvider, err := broker.getDefinitionAndProvider(parsedDetails.ServiceID)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// verify the service exists and the plan exists
	plan, err := serviceDefinition.GetPlanByID(parsedDetails.PlanID)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// Give the user a better error message if they give us a bad request
	if err := validateProvisionParameters(
		parsedDetails.RequestParams,
		serviceDefinition.ProvisionInputVariables,
		serviceDefinition.ImportInputVariables,
		plan); err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// validate parameters meet the service's schema and merge the user vars with
	// the plan's
	vars, err := serviceDefinition.ProvisionVariables(instanceID, parsedDetails, *plan, request.DecodeOriginatingIdentityHeader(ctx))
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	err = serviceProvider.Provision(ctx, vars)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, err
	}

	// save instance details
	instanceDetails := storage.ServiceInstanceDetails{
		ServiceGUID:      parsedDetails.ServiceID,
		GUID:             instanceID,
		PlanGUID:         parsedDetails.PlanID,
		SpaceGUID:        parsedDetails.SpaceGUID,
		OrganizationGUID: parsedDetails.OrganizationGUID,
	}

	if err := broker.store.StoreServiceInstanceDetails(instanceDetails); err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("error saving instance details to database: %s. WARNING: this instance cannot be deprovisioned through cf. Contact your operator for cleanup", err)
	}

	// save provision request details
	delete(parsedDetails.RequestParams, "vacant")
	if err := broker.store.StoreProvisionRequestDetails(instanceID, parsedDetails.RequestParams); err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("error saving provision request details to database: %s. Services relying on async provisioning will not be able to complete provisioning", err)
	}

	operationId := generateTFInstanceID(instanceDetails.GUID)

	return domain.ProvisionedServiceSpec{IsAsync: true, DashboardURL: "", OperationData: operationId}, nil
}
