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

	"code.cloudfoundry.org/brokerapi/v13"
	"code.cloudfoundry.org/brokerapi/v13/domain"
)

//go:generate go tool counterfeiter -generate
//counterfeiter:generate -o ./fakes/servicebroker.go code.cloudfoundry.org/brokerapi/v13/domain.ServiceBroker

// CfSharingWrapper enables the Shareable flag for every service provided by the broker.
type CfSharingWrapper struct {
	brokerapi.ServiceBroker
}

// Services augments the response from the wrapped ServiceBroker by adding
// the shareable flag.
func (w *CfSharingWrapper) Services(ctx context.Context) (services []domain.Service, err error) {
	services, err = w.ServiceBroker.Services(ctx)

	for i := range services {
		if services[i].Metadata == nil {
			services[i].Metadata = &domain.ServiceMetadata{}
		}

		services[i].Metadata.Shareable = domain.BindableValue(true)
	}

	return
}

// NewCfSharingWrapper wraps the given servicebroker with the augmenter that
// sets the Shareable flag on all services.
func NewCfSharingWrapper(wrapped domain.ServiceBroker) domain.ServiceBroker {
	return &CfSharingWrapper{ServiceBroker: wrapped}
}
