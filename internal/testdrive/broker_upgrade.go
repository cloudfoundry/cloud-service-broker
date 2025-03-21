package testdrive

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/steps"
	"github.com/google/uuid"
)

type upgradeConfig struct {
	params         json.RawMessage
	previousValues domain.PreviousValues
}

type UpgradeOption func(*upgradeConfig) error

func (b *Broker) UpgradeService(s ServiceInstance, version string, opts ...UpgradeOption) error {
	maintenanceInfo := domain.MaintenanceInfo{Version: version}

	var cfg upgradeConfig
	for _, o := range opts {
		if err := o(&cfg); err != nil {
			return err
		}
	}

	return steps.RunSequentially(
		func() error {
			updateResponse := b.Client.Update(s.GUID, s.ServiceOfferingGUID, s.ServicePlanGUID, uuid.NewString(), cfg.params, cfg.previousValues, &maintenanceInfo)
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

func WithUpgradePreviousValues(v domain.PreviousValues) UpgradeOption {
	return func(cfg *upgradeConfig) error {
		cfg.previousValues = v
		return nil
	}
}

func WithUpgradeParams(params any) UpgradeOption {
	return func(cfg *upgradeConfig) error {
		jsonParams, err := toJSONRawMessage(params)
		if err != nil {
			return err
		}
		cfg.params = jsonParams
		return nil
	}
}
