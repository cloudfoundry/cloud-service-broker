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
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/validation"
	"github.com/pivotal-cf/brokerapi/v12/domain"
)

// Service overrides the canonical Service Broker service type using a custom
// type for Plans, everything else is the same.
type Service struct {
	domain.Service

	Plans []ServicePlan `json:"plans"`
}

func (s *Service) Validate() (errs *validation.FieldError) {
	names := make(map[string]struct{})
	ids := make(map[string]struct{})
	for i, v := range s.Plans {
		errs = errs.Also(
			v.Validate().ViaFieldIndex("Plans", i),
			validation.ErrIfDuplicate(v.Name, "Name", names).ViaFieldIndex("Plans", i),
			validation.ErrIfDuplicate(v.ID, "Id", ids).ViaFieldIndex("Plans", i),
		)
	}
	return errs
}

// ToPlain converts this service to a plain PCF Service definition.
func (s Service) ToPlain() domain.Service {
	plain := s.Service
	var plainPlans []domain.ServicePlan

	for _, plan := range s.Plans {
		plainPlans = append(plainPlans, plan.ServicePlan)
	}

	plain.Plans = plainPlans

	return plain
}

// ServicePlan extends the OSB ServicePlan by including a map of key/value
// pairs that can be used to pass additional information to the back-end.
type ServicePlan struct {
	domain.ServicePlan

	ServiceProperties  map[string]any `json:"service_properties"`
	ProvisionOverrides map[string]any `json:"provision_overrides,omitempty"`
	BindOverrides      map[string]any `json:"bind_overrides,omitempty"`
}

// Validate implements validation.Validatable.
func (sp *ServicePlan) Validate() (errs *validation.FieldError) {
	return errs.Also(
		validation.ErrIfBlank(sp.Name, "Name"),
		validation.ErrIfNotUUID(sp.ID, "ID"),
	)
}

// GetServiceProperties gets the plan settings variables as a string->interface map.
func (sp *ServicePlan) GetServiceProperties() map[string]any {
	return sp.ServiceProperties
}
