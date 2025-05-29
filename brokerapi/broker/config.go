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
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/credhubrepo"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/brokerpak"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/config"
)

//counterfeiter:generate . CredStore
type CredStore interface {
	Save(ctx context.Context, path string, cred any, actor string) (any, error)
	Delete(ctx context.Context, path string) error
}
type BrokerConfig struct {
	Registry  broker.BrokerRegistry
	CredStore CredStore
}

func NewBrokerConfigFromEnv(logger lager.Logger) (*BrokerConfig, error) {
	registry := broker.BrokerRegistry{}
	if err := brokerpak.RegisterAll(registry); err != nil {
		return nil, fmt.Errorf("error loading brokerpaks: %v", err)
	}

	envConfig, err := config.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed loading config: %v", err)
	}

	var credStore CredStore = NoopCredStore{}
	if envConfig.CredStoreConfig.HasCredHubConfig() {
		logger.Info("using-credhub")
		var err error
		credStore, err = credhubrepo.New(logger, envConfig.CredStoreConfig)
		if err != nil {
			return nil, err
		}
	}

	return &BrokerConfig{
		Registry:  registry,
		CredStore: credStore,
	}, nil
}

type NoopCredStore struct{}

func (NoopCredStore) Save(ctx context.Context, path string, cred any, actor string) (any, error) {
	return cred, nil
}

func (NoopCredStore) Delete(ctx context.Context, path string) error {
	return nil
}
