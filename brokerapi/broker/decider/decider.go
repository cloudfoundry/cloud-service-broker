package decider

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

type Operation int

const (
	Failed Operation = iota
	Update
	Upgrade
)

const upgradeBeforeUpdateError = "service instance needs to be upgraded before updating"

func DecideOperation(planMaintenanceInfo *domain.MaintenanceInfo, details domain.UpdateDetails) (Operation, error) {
	if err := validateMaintenanceInfo(planMaintenanceInfo, details.PlanID, details.MaintenanceInfo); err != nil {
		return Failed, err
	}

	if planNotChanged(details) && requestParamsEmpty(details) && requestMaintenanceInfoValuesDiffer(details) {
		return Upgrade, nil
	}

	if err := validatePreviousMaintenanceInfo(details, planMaintenanceInfo); err != nil {
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

func validateMaintenanceInfo(planMaintenanceInfo *domain.MaintenanceInfo, planID string, catalogMaintenanceInfo *domain.MaintenanceInfo) error {
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

func validatePreviousMaintenanceInfo(details domain.UpdateDetails, planMaintenanceInfo *domain.MaintenanceInfo) error {
	if details.PreviousValues.MaintenanceInfo != nil {
		if maintenanceInfoConflict(details.PreviousValues.MaintenanceInfo, planMaintenanceInfo) {
			return errInstanceMustBeUpgradedFirst()
		}
	}
	return nil
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
