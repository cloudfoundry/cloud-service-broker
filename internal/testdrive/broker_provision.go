package testdrive

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/steps"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

type ServiceInstance struct {
	GUID                string
	ServicePlanGUID     string
	ServiceOfferingGUID string
}

type provisionConfig struct {
	guid   string
	params json.RawMessage
}

type ProvisionOption func(*provisionConfig) error

func (b *Broker) Provision(serviceOfferingGUID, servicePlanGUID string, opts ...ProvisionOption) (ServiceInstance, error) {
	cfg := provisionConfig{guid: uuid.New()}
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return ServiceInstance{}, err
		}
	}

	err := steps.RunSequentially(
		func() error {
			provisionResponse := b.Client.Provision(cfg.guid, serviceOfferingGUID, servicePlanGUID, uuid.New(), cfg.params)
			switch {
			case provisionResponse.Error != nil:
				return provisionResponse.Error
			case provisionResponse.StatusCode != http.StatusAccepted:
				return &UnexpectedStatusError{StatusCode: provisionResponse.StatusCode, ResponseBody: provisionResponse.ResponseBody}
			default:
				return nil
			}
		},
		func() error {
			state, err := b.LastOperationFinalState(cfg.guid)
			switch {
			case err != nil:
				return err
			case state != domain.Succeeded:
				return fmt.Errorf("provision failed with state: %s", state)
			default:
				return nil
			}
		},
	)

	// If it fails, we still return the GUIDs for cleanup
	return ServiceInstance{
		GUID:                cfg.guid,
		ServicePlanGUID:     servicePlanGUID,
		ServiceOfferingGUID: serviceOfferingGUID,
	}, err
}

func WithProvisionParams(params any) ProvisionOption {
	return func(cfg *provisionConfig) error {
		jsonParams, err := toJSONRawMessage(params)
		if err != nil {
			return err
		}
		cfg.params = jsonParams
		return nil
	}
}

func WithProvisionServiceInstanceGUID(guid string) ProvisionOption {
	return func(cfg *provisionConfig) error {
		cfg.guid = guid
		return nil
	}
}
