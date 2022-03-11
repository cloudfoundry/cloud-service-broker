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
	"fmt"
	"sort"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/pkg/credstore"
	"github.com/spf13/viper"
)

const (
	credhubClientIdentifier          = "csb"
	DisableRequestPropertyValidation = "request.property.validation.disabled"
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

func validateProvisionParameters(params map[string]interface{}, validUserInputFields []broker.BrokerVariable, validImportFields []broker.ImportVariable, plan *broker.ServicePlan) error {
	if len(params) == 0 {
		return nil
	}

	// As this is a new check we have feature-flagged it so that it can easily be disabled
	// if it causes problems.
	if !viper.GetBool(DisableRequestPropertyValidation) {
		err := validateNoPlanParametersOverrides(params, plan)
		if err != nil {
			return err
		}

		return validateDefinedParams(params, validUserInputFields, validImportFields)
	}

	return nil
}

func validateNoPlanParametersOverrides(params map[string]interface{}, plan *broker.ServicePlan) error {
	var invalidPlanParams []string
	for k := range params {
		if _, ok := plan.ServiceProperties[k]; ok {
			invalidPlanParams = append(invalidPlanParams, k)
		}
	}

	if len(invalidPlanParams) != 0 {
		sort.Strings(invalidPlanParams)
		return fmt.Errorf("plan defined properties cannot be changed: %s", strings.Join(invalidPlanParams, ", "))
	}
	return nil
}

func validateDefinedParams(params map[string]interface{}, validUserInputFields []broker.BrokerVariable, validImportFields []broker.ImportVariable) error {
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
