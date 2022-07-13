// Copyright 2020 Pivotal Software, Inc.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//    http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package credstore

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/credhub-cli/credhub"
	"code.cloudfoundry.org/credhub-cli/credhub/auth"
	"code.cloudfoundry.org/credhub-cli/credhub/permissions"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/cloud-service-broker/pkg/config"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . CredStore

type CredStore interface {
	Put(key string, credentials any) (any, error)
	PutValue(key string, credentials any) (any, error)
	Get(key string) (any, error)
	GetValue(key string) (string, error)
	Delete(key string) error
	AddPermission(path string, actor string, ops []string) (*permissions.Permission, error)
	DeletePermission(path string) error
}

type credhubStore struct {
	credHubClient *credhub.CredHub
	logger        lager.Logger
}

func NewCredhubStore(credStoreConfig *config.CredStoreConfig, logger lager.Logger) (CredStore, error) {
	if os.Getenv("DEV_MODE_ONLY") != "" {
		logger.Debug(fmt.Sprintf("DEV_MODE_ONLY [%+v] - Creating Mock Credhub", os.Getenv("DEV_MODE_ONLY")))
		return &credHubStoreMock{}, nil
	}

	if !credStoreConfig.HasCredHubConfig() {
		return nil, fmt.Errorf("CredHubConfig not found")
	}
	options := []credhub.Option{
		credhub.SkipTLSValidation(credStoreConfig.SkipSSLValidation),
		credhub.Auth(auth.UaaClientCredentials(credStoreConfig.UaaClientName, credStoreConfig.UaaClientSecret)),
		credhub.AuthURL(credStoreConfig.UaaURL),
	}

	if credStoreConfig.CaCertFile != "" {
		dat, err := os.ReadFile(credStoreConfig.CaCertFile)
		if err != nil {
			return nil, err
		}

		if dat == nil {
			return nil, fmt.Errorf("CredHub certificate is not valid: %s", credStoreConfig.CaCertFile)
		}
		options = append(options, credhub.CaCerts(string(dat)))
	}

	ch, err := credhub.New(credStoreConfig.CredHubURL, options...)

	if err != nil {
		return nil, err
	}

	return &credhubStore{
		credHubClient: ch,
		logger:        logger,
	}, err
}

func (c *credhubStore) Put(key string, credentials any) (any, error) {
	return c.credHubClient.SetCredential(key, "json", credentials)
}

func (c *credhubStore) PutValue(key string, credentials any) (any, error) {
	return c.credHubClient.SetCredential(key, "value", credentials)
}

func (c *credhubStore) Get(key string) (any, error) {
	return c.credHubClient.GetLatestValue(key)
}

func (c *credhubStore) GetValue(key string) (string, error) {
	value, err := c.credHubClient.GetLatestValue(key)
	if err != nil {
		return "", err
	}
	return string(value.Value), nil
}

func (c *credhubStore) Delete(key string) error {
	return c.credHubClient.Delete(key)
}

func (c *credhubStore) AddPermission(path string, actor string, ops []string) (*permissions.Permission, error) {
	return c.credHubClient.AddPermission(path, actor, ops)
}

func (c *credhubStore) DeletePermission(path string) error {
	allPermissions, err := c.credHubClient.GetPermissions(path)
	if err != nil {
		return err
	}

	for _, permission := range allPermissions {
		p, err := c.credHubClient.GetPermissionByPathActor(path, permission.Actor)
		if err != nil {
			return err
		}
		_, err = c.credHubClient.DeletePermission(p.UUID)
		if err != nil {
			return err
		}

	}

	return err
}
