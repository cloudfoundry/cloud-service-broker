package testdrive

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/internal/steps"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type updateConfig struct {
	params          json.RawMessage
	previousValues  domain.PreviousValues
	maintenanceInfo *domain.MaintenanceInfo
}

type UpdateOption func(*updateConfig) error

func (b *Broker) UpdateService(s ServiceInstance, opts ...UpdateOption) error {
	var cfg updateConfig
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return err
		}
	}

	return steps.Sequentially(
		func() error {
			updateResponse := b.Client.Update(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.New(), cfg.params, cfg.previousValues, cfg.maintenanceInfo)
			switch {
			case updateResponse.Error != nil:
				return updateResponse.Error
			case updateResponse.StatusCode != http.StatusAccepted:
				return fmt.Errorf("unexpected status code %d: %s", updateResponse.StatusCode, updateResponse.ResponseBody)
			default:
				return nil
			}
		},
		func() error {
			state, err := b.LastOperationFinalState(s.GUID)
			switch {
			case err != nil:
				return err
			case state != domain.Succeeded:
				return fmt.Errorf("update failed with state: %s", state)
			default:
				return nil
			}
		},
	)
}

func WithUpdateParams(params any) UpdateOption {
	return func(cfg *updateConfig) error {
		jsonParams, err := toJSONRawMessage(params)
		if err != nil {
			return err
		}
		cfg.params = jsonParams
		return nil
	}
}

func WithUpdatePreviousValues(v domain.PreviousValues) UpdateOption {
	return func(cfg *updateConfig) error {
		cfg.previousValues = v
		return nil
	}
}

func WithUpdateMaintenanceInfo(m domain.MaintenanceInfo) UpdateOption {
	return func(cfg *updateConfig) error {
		cfg.maintenanceInfo = &m
		return nil
	}
}
