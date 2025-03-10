package testdrive

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/steps"
	"github.com/google/uuid"
)

type updateConfig struct {
	params          json.RawMessage
	previousValues  domain.PreviousValues
	maintenanceInfo *domain.MaintenanceInfo
	servicePlanGUID string
}

type UpdateOption func(*updateConfig) error

func (b *Broker) UpdateService(s ServiceInstance, opts ...UpdateOption) error {
	cfg := updateConfig{
		servicePlanGUID: s.ServicePlanGUID,
	}
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return err
		}
	}

	return steps.RunSequentially(
		func() error {
			updateResponse := b.Client.Update(s.GUID, s.ServiceOfferingGUID, cfg.servicePlanGUID, uuid.NewString(), cfg.params, cfg.previousValues, cfg.maintenanceInfo)
			switch {
			case updateResponse.Error != nil:
				return updateResponse.Error
			case updateResponse.StatusCode != http.StatusAccepted:
				return &UnexpectedStatusError{StatusCode: updateResponse.StatusCode, ResponseBody: updateResponse.ResponseBody}
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

func WithUpdatePlan(servicePlanGUID string) UpdateOption {
	return func(cfg *updateConfig) error {
		cfg.servicePlanGUID = servicePlanGUID
		return nil
	}
}
