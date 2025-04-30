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

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/brokerchapi"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/brokercredstore"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/brokerpak"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/config"
)

type BrokerConfig struct {
	Registry  broker.BrokerRegistry
	Credstore brokercredstore.BrokerCredstore
}

func NewBrokerConfigFromEnv() (*BrokerConfig, error) {
	registry := broker.BrokerRegistry{}
	if err := brokerpak.RegisterAll(registry); err != nil {
		return nil, fmt.Errorf("error loading brokerpaks: %v", err)
	}

	envConfig, err := config.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed loading config: %v", err)
	}

	var credHubStore *brokerchapi.Store
	if envConfig.CredStoreConfig.HasCredHubConfig() {
		var err error
		credHubStore, err = brokerchapi.New(brokerchapi.Config{
			CredHubURL:            envConfig.CredStoreConfig.CredHubURL,
			CACert:                envConfig.CredStoreConfig.CACert,
			UAAURL:                envConfig.CredStoreConfig.UaaURL,
			UAAClientName:         envConfig.CredStoreConfig.UaaClientName,
			UAAClientSecret:       envConfig.CredStoreConfig.UaaClientSecret,
			InsecureSkipTLSVerify: envConfig.CredStoreConfig.SkipSSLValidation,
		})
		if err != nil {
			return nil, err
		}
	}

	return &BrokerConfig{
		Registry:  registry,
		Credstore: brokercredstore.NewBrokerCredstore(credHubStore),
	}, nil
}
