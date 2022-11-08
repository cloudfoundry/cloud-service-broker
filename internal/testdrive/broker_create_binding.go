package testdrive

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/pkg/client"
	"github.com/pborman/uuid"
)

type ServiceBinding struct {
	GUID string
	Body string
}

type createBindingConfig struct {
	params json.RawMessage
	guid   string
}

type CreateBindingOption func(*createBindingConfig) error

func (b *Broker) CreateBinding(s ServiceInstance, opts ...CreateBindingOption) (ServiceBinding, error) {
	var bindResponse *client.BrokerResponse
	cfg := createBindingConfig{
		guid: uuid.New(),
	}

	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return ServiceBinding{}, err
		}
	}

	bindResponse = b.Client.Bind(s.GUID, cfg.guid, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New(), cfg.params)
	switch {
	case bindResponse.Error != nil:
		return ServiceBinding{}, bindResponse.Error
	case bindResponse.StatusCode != http.StatusCreated:
		return ServiceBinding{}, fmt.Errorf("unexpected status code %d: %s", bindResponse.StatusCode, bindResponse.ResponseBody)
	default:
		return ServiceBinding{
			GUID: cfg.guid,
			Body: string(bindResponse.ResponseBody),
		}, nil
	}
}

func WithBindingParams(params any) CreateBindingOption {
	return func(cfg *createBindingConfig) error {
		jsonParams, err := toJSONRawMessage(params)
		if err != nil {
			return err
		}
		cfg.params = jsonParams
		return nil
	}
}
