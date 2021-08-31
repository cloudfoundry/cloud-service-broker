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

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/validation"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/toggles"
	"github.com/cloudfoundry-incubator/cloud-service-broker/utils"
)

var (
	// The following flags enable and disable services based on their tags.
	// The guiding philosophy for defaults is optimistic about new technology and pessimistic about old.
	lifecycleTagToggles = map[string]toggles.Toggle{
		"preview":      toggles.Features.Toggle("enable-preview-services", true, `Enable services that are new to the broker this release.`),
		"unmaintained": toggles.Features.Toggle("enable-unmaintained-services", false, `Enable broker services that are unmaintained.`),
		"eol":          toggles.Features.Toggle("enable-eol-services", false, `Enable broker services that are end of life.`),
		"beta":         toggles.Features.Toggle("enable-gcp-beta-services", true, "Enable services that are in GCP Beta. These have no SLA or support policy."),
		"deprecated":   toggles.Features.Toggle("enable-gcp-deprecated-services", false, "Enable services that use deprecated GCP components."),
		"terraform":    toggles.Features.Toggle("enable-terraform-services", false, "Enable services that use the experimental, unstable, Terraform back-end."),
	}

	enableBuiltinServices = toggles.Features.Toggle("enable-builtin-services", true, `Enable services that are built in to the broker i.e. not brokerpaks.`)
)

// BrokerRegistry holds the list of ServiceDefinitions that can be provisioned
// by the Service Broker.
type BrokerRegistry map[string]*ServiceDefinition

// Register registers a ServiceDefinition with the service registry that various commands
// poll to create the catalog, documentation, etc.
func (brokerRegistry BrokerRegistry) Register(service *ServiceDefinition) error {
	name := service.Name

	if _, ok := brokerRegistry[name]; ok {
		return fmt.Errorf("tried to register multiple instances of: %q", name)
	}

	userPlans, err := service.UserDefinedPlans()
	if err != nil {
		return fmt.Errorf("error getting user defined plans: %q, %s", name, err)
	}
	service.Plans = append(service.Plans, userPlans...)

	if err := service.Validate(); err != nil {
		return fmt.Errorf("error validating service %q, %s", name, err)
	}

	brokerRegistry[name] = service
	return nil
}

func (brokerRegistry BrokerRegistry) Validate() (errs *validation.FieldError) {
	services := brokerRegistry.GetAllServices()
	serviceIDs := make(map[string]struct{})
	serviceNames := make(map[string]struct{})
	planIDs := make(map[string]struct{})
	for i, s := range services {
		errs = errs.Also(
			validation.ErrIfDuplicate(s.Id, "Id", serviceIDs).ViaFieldIndex("services", i),
			validation.ErrIfDuplicate(s.Name, "Name", serviceNames).ViaFieldIndex("services", i),
		)

		for j, p := range s.Plans {
			errs = errs.Also(
				validation.ErrIfDuplicate(p.ID, "Id", planIDs).ViaFieldIndex("Plans", j).ViaFieldIndex("services", i),
			)
		}
	}

	return errs
}

// GetEnabledServices returns a list of all registered brokers that the user
// has enabled the use of.
func (brokerRegistry *BrokerRegistry) GetEnabledServices() ([]*ServiceDefinition, error) {
	var out []*ServiceDefinition

	for _, svc := range brokerRegistry.GetAllServices() {
		isEnabled := true

		if svc.IsBuiltin {
			isEnabled = enableBuiltinServices.IsActive()
		}

		entry := svc.CatalogEntry()
		tags := utils.NewStringSet(entry.Tags...)
		for tag, toggle := range lifecycleTagToggles {
			if !toggle.IsActive() && tags.Contains(tag) {
				isEnabled = false
				break
			}
		}

		if isEnabled {
			out = append(out, svc)
		}
	}

	return out, nil
}

// GetAllServices returns a list of all registered brokers whether or not the
// user has enabled them. The brokers are sorted in lexocographic order based
// on name.
func (brokerRegistry BrokerRegistry) GetAllServices() []*ServiceDefinition {
	var out []*ServiceDefinition

	for _, svc := range brokerRegistry {
		out = append(out, svc)
	}

	// Sort by name so there's a consistent order in the UI and tests.
	sort.Slice(out, func(i int, j int) bool { return out[i].Name < out[j].Name })

	return out
}

// GetServiceById returns the service with the given ID, if it does not exist
// or one of the services has a parse error then an error is returned.
func (brokerRegistry BrokerRegistry) GetServiceById(id string) (*ServiceDefinition, error) {
	for _, svc := range brokerRegistry {
		if svc.Id == id {
			return svc, nil
		}
	}

	return nil, fmt.Errorf("unknown service ID: %q", id)
}
