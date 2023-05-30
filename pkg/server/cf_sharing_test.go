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

package server

import (
	"context"
	"errors"
	"testing"

	"github.com/pivotal-cf/brokerapi/v10/domain"

	"github.com/cloudfoundry/cloud-service-broker/pkg/server/fakes"
)

func TestCfSharingWraper_Services(t *testing.T) {
	cases := map[string]struct {
		Services []domain.Service
		Error    error
	}{
		"nil services": {
			Services: nil,
			Error:    nil,
		},
		"empty services": {
			Services: []domain.Service{},
			Error:    nil,
		},

		"single service": {
			Services: []domain.Service{
				{Name: "foo", Metadata: &domain.ServiceMetadata{}},
			},
			Error: nil,
		},
		"missing metadata": {
			Services: []domain.Service{
				{Name: "foo"},
			},
			Error: nil,
		},
		"multiple services": {
			Services: []domain.Service{
				{Name: "foo", Metadata: &domain.ServiceMetadata{}},
				{Name: "bar", Metadata: &domain.ServiceMetadata{}},
			},
			Error: nil,
		},
		"error passed": {
			Services: nil,
			Error:    errors.New("returned error"),
		},
		"services and err": {
			Services: []domain.Service{
				{Name: "foo", Metadata: &domain.ServiceMetadata{}},
				{Name: "bar", Metadata: &domain.ServiceMetadata{}},
			},
			Error: errors.New("returned error"),
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			wrapped := &fakes.FakeServiceBroker{}
			wrapped.ServicesReturns(tc.Services, tc.Error)

			sw := NewCfSharingWrapper(wrapped)
			services, actualErr := sw.Services(context.Background())

			if tc.Error != actualErr {
				t.Fatalf("Expected error: %v got: %v", tc.Error, actualErr)
			}

			if wrapped.ServicesCallCount() != 1 {
				t.Errorf("Expected 1 call to Services() got %v", wrapped.ServicesCallCount())
			}

			if len(services) != len(tc.Services) {
				t.Errorf("Expected to get back %d services got %d", len(tc.Services), len(services))
			}

			for i, svc := range services {
				if svc.Metadata == nil {
					t.Fatalf("Expected service %d to have metadata, but was nil", i)
				}

				if svc.Metadata.Shareable == nil {
					t.Fatalf("Expected service %d to have shareable, but was nil", i)
				}

				if *svc.Metadata.Shareable != true {
					t.Fatalf("Expected service %d to be shareable, but was %v", i, *svc.Metadata.Shareable)
				}
			}
		})
	}
}
