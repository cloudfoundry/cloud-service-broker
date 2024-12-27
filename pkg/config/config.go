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

// Package config implements configuration parsing for CredHub
package config

import (
	"github.com/spf13/viper"
)

const (
	credhubURL                  = "credhub.url"
	credhubUaaURL               = "credhub.uaa_url"
	credhubUaaClientName        = "credhub.uaa_client_name"
	credhubUaaClientSecret      = "credhub.uaa_client_secret"
	credhubSkipSSLValidation    = "credhub.skip_ssl_validation"
	credhubCACert               = "credhub.ca_cert"
	credhubStoreBindCredentials = "credhub.store_bind_credentials"
)

type CredStoreConfig struct {
	CredHubURL           string `mapstructure:"url"`
	UaaURL               string `mapstructure:"uaa_url"`
	UaaClientName        string `mapstructure:"uaa_client_name"`
	UaaClientSecret      string `mapstructure:"uaa_client_secret"`
	SkipSSLValidation    bool   `mapstructure:"skip_ssl_validation"`
	StoreBindCredentials bool   `mapstructure:"store_bind_credentials"`
	CACert               string `mapstructure:"ca_cert"`
}

type Config struct {
	CredStoreConfig CredStoreConfig `mapstructure:"credhub"`
}

func Parse() (*Config, error) {
	c := Config{}
	viper.BindEnv(credhubURL, "CH_CRED_HUB_URL")
	viper.BindEnv(credhubUaaURL, "CH_UAA_URL")
	viper.BindEnv(credhubUaaClientName, "CH_UAA_CLIENT_NAME")
	viper.BindEnv(credhubUaaClientSecret, "CH_UAA_CLIENT_SECRET")
	viper.BindEnv(credhubSkipSSLValidation, "CH_SKIP_SSL_VALIDATION")
	viper.BindEnv(credhubStoreBindCredentials, "CH_STORE_BIND_CREDENTIALS")
	viper.BindEnv(credhubCACert, "CH_CA_CERT")

	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *CredStoreConfig) HasCredHubConfig() bool {
	return c.CredHubURL != ""
}
