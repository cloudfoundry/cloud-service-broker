// From Kibosh

package config

import (
	"github.com/spf13/viper"
)

const (
	credhubURL = "credhub.url"
	credhubUaaURL = "credhub.uaa_url"
	credhubUaaClientName = "credhub.uaa_client_name"
	credhubUaaClientSecret = "credhub.uaa_client_secret"
	credhubSkipSSLValidation = "credhub.skip_ssl_validation"
	credhubCaCertFile = "credhub.ca_cert_file"
	credhubStoreBindCredentials = "credhub.store_bind_credentials"
)

type CredStoreConfig struct {
	CredHubURL           string `mapstructure:"url"`
	UaaURL               string `mapstructure:"uaa_url"`
	UaaClientName        string `mapstructure:"uaa_client_name"`
	UaaClientSecret      string `mapstructure:"uaa_client_secret"`
	SkipSSLValidation    bool   `mapstructure:"skip_ssl_validation"`
	CaCertFile           string `mapstructure:"ca_cert_file"`
	StoreBindCredentials bool   `mapstructure:"ca_cert_file"`
}

type Config struct {
	CredStoreConfig     CredStoreConfig `mapstructure:"credhub"`
}

func Parse() (*Config, error) {
	c := Config{}
	viper.BindEnv(credhubURL, "CH_CRED_HUB_URL")
	viper.BindEnv(credhubUaaURL, "CH_UAA_URL")
	viper.BindEnv(credhubUaaClientName, "CH_UAA_CLIENT_NAME")
	viper.BindEnv(credhubUaaClientSecret, "CH_UAA_CLIENT_SECRET")
	viper.BindEnv(credhubSkipSSLValidation, "CH_SKIP_SSL_VALIDATION")
	viper.BindEnv(credhubCaCertFile, "CH_CA_CERT_FILE")
	viper.BindEnv(credhubStoreBindCredentials, "CH_STORE_BIND_CREDENTIALS")

	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *CredStoreConfig) HasCredHubConfig() bool {
	return c.CredHubURL != ""
}


