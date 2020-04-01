// From Kibosh

package config

import (
	"github.com/kelseyhightower/envconfig"
)

type CredStoreConfig struct {
	CredHubURL           string `envconfig:"CH_CRED_HUB_URL"`
	UaaURL               string `envconfig:"CH_UAA_URL"`
	UaaClientName        string `envconfig:"CH_UAA_CLIENT_NAME"`
	UaaClientSecret      string `envconfig:"CH_UAA_CLIENT_SECRET"`
	SkipSSLValidation    bool   `envconfig:"CH_SKIP_SSL_VALIDATION"`
	CaCertFile           string `envconfig:"CH_CA_CERT_FILE"`
	StoreBindCredentials bool   `envconfig:"CH_STORE_BIND_CREDENTIALS"`
}

type Config struct {
	CredStoreConfig     *CredStoreConfig
}

func Parse() (*Config, error) {
	c := &Config{}
	err := envconfig.Process("", c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *CredStoreConfig) HasCredHubConfig() bool {
	return c.CredHubURL != ""
}


