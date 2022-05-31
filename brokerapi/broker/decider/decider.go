package decider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type Decider struct{}

type Operation int

const (
	Failed Operation = iota
	Update
	Upgrade
)

const upgradeBeforeUpdateError = "service instance needs to be upgraded before updating"

func (d Decider) DecideOperation(service *broker.ServiceDefinition, details domain.UpdateDetails) (Operation, error) {
	if err := validateMaintenanceInfo(service, details.PlanID, details.MaintenanceInfo); err != nil {
		return Failed, err
	}

	if planNotChanged(details) && requestParamsEmpty(details) && requestMaintenanceInfoValuesDiffer(details) {
		return Upgrade, nil
	}

	if err := validatePreviousMaintenanceInfo(details, service); err != nil {
		return Failed, err
	}

	return Update, nil
}

func planNotChanged(details domain.UpdateDetails) bool {
	return details.PlanID == details.PreviousValues.PlanID
}

func requestParamsEmpty(details domain.UpdateDetails) bool {
	if len(details.RawParameters) == 0 {
		return true
	}

	var params map[string]interface{}
	if err := json.Unmarshal(details.RawParameters, &params); err != nil {
		return false
	}
	return len(params) == 0
}

func requestMaintenanceInfoValuesDiffer(details domain.UpdateDetails) bool {
	switch {
	case details.MaintenanceInfo == nil && details.PreviousValues.MaintenanceInfo != nil:
		return true
	case details.MaintenanceInfo != nil && details.PreviousValues.MaintenanceInfo == nil:
		return true
	case details.MaintenanceInfo == nil && details.PreviousValues.MaintenanceInfo == nil:
		return false
	default:
		return !details.MaintenanceInfo.Equals(*details.PreviousValues.MaintenanceInfo)
	}
}

func validateMaintenanceInfo(service *broker.ServiceDefinition, planID string, catalogMaintenanceInfo *domain.MaintenanceInfo) error {
	planMaintenanceInfo, err := getMaintenanceInfoForPlan(service, planID)
	if err != nil {
		return err
	}

	if maintenanceInfoConflict(catalogMaintenanceInfo, planMaintenanceInfo) {
		if catalogMaintenanceInfo == nil {
			return errMaintenanceInfoNilInTheRequest()
		}

		if planMaintenanceInfo == nil {
			return apiresponses.ErrMaintenanceInfoNilConflict
		}

		return apiresponses.ErrMaintenanceInfoConflict
	}

	return nil
}

func validatePreviousMaintenanceInfo(details domain.UpdateDetails, service *broker.ServiceDefinition) error {
	if details.PreviousValues.MaintenanceInfo != nil {
		catalogPreviousPlanMaintenanceInfo, err := getMaintenanceInfoForPlan(service, details.PreviousValues.PlanID)
		if err != nil {
			return fmt.Errorf("service instance needs to be upgraded: %w. Contact the operator for assistance", err)
		}
		if maintenanceInfoConflict(details.PreviousValues.MaintenanceInfo, catalogPreviousPlanMaintenanceInfo) {
			return errInstanceMustBeUpgradedFirst()
		}
	}
	return nil
}

func getMaintenanceInfoForPlan(service *broker.ServiceDefinition, id string) (*domain.MaintenanceInfo, error) {
	for _, plan := range service.Plans {
		if plan.ID == id {
			if plan.MaintenanceInfo != nil {
				return &domain.MaintenanceInfo{
					Version:     plan.MaintenanceInfo.Version,
					Description: plan.MaintenanceInfo.Description,
				}, nil
			}
			return nil, nil
		}
	}

	return nil, fmt.Errorf("plan %s does not exist", id)
}

func maintenanceInfoConflict(a, b *domain.MaintenanceInfo) bool {
	if a != nil && b != nil {
		return !a.Equals(*b)
	}

	if a == nil && b == nil {
		return false
	}

	return true
}

func errInstanceMustBeUpgradedFirst() *apiresponses.FailureResponse {
	return apiresponses.NewFailureResponseBuilder(
		errors.New(upgradeBeforeUpdateError),
		http.StatusUnprocessableEntity,
		"previous-maintenance-info-check",
	).Build()
}

func errMaintenanceInfoNilInTheRequest() *apiresponses.FailureResponse {
	return apiresponses.NewFailureResponseBuilder(
		errors.New(upgradeBeforeUpdateError+": maintenance info defined in broker service catalog, but not passed in request"),
		http.StatusUnprocessableEntity,
		"previous-maintenance-info-check",
	).Build()
}
