package decider

import (
	"errors"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
	"github.com/hashicorp/go-version"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

type Operation int

const (
	Failed Operation = iota
	Update
	Upgrade
)

const upgradeBeforeUpdateError = "service instance needs to be upgraded before updating"

func DecideOperation(planMaintenanceInfoVersion *version.Version, details paramparser.UpdateDetails) (Operation, error) {
	if err := validateMaintenanceInfo(planMaintenanceInfoVersion, details.MaintenanceInfoVersion); err != nil {
		return Failed, err
	}

	if planNotChanged(details) && requestParamsEmpty(details) && requestMaintenanceInfoValuesDiffer(details) {
		return Upgrade, nil
	}

	if err := validatePreviousMaintenanceInfo(details, planMaintenanceInfoVersion); err != nil {
		return Failed, err
	}

	return Update, nil
}

func planNotChanged(details paramparser.UpdateDetails) bool {
	return details.PlanID == details.PreviousPlanID
}

func requestParamsEmpty(details paramparser.UpdateDetails) bool {
	return len(details.RequestParams) == 0
}

func requestMaintenanceInfoValuesDiffer(details paramparser.UpdateDetails) bool {
	return !details.MaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion)
}

func validateMaintenanceInfo(planMaintenanceInfo, requestMaintenanceInfoVersion *version.Version) error {
	if maintenanceInfoDifference(requestMaintenanceInfoVersion, planMaintenanceInfo) {
		if requestMaintenanceInfoVersion == nil {
			return errMaintenanceInfoNilInTheRequest()
		}

		if planMaintenanceInfo == nil {
			return apiresponses.ErrMaintenanceInfoNilConflict
		}

		return apiresponses.ErrMaintenanceInfoConflict
	}

	return nil
}

func validatePreviousMaintenanceInfo(details paramparser.UpdateDetails, planMaintenanceInfoVersion *version.Version) error {
	if details.PreviousMaintenanceInfoVersion != nil {
		if maintenanceInfoDifference(details.PreviousMaintenanceInfoVersion, planMaintenanceInfoVersion) {
			return errInstanceMustBeUpgradedFirst()
		}
	}
	return nil
}

func maintenanceInfoDifference(a, b *version.Version) bool {
	return !a.Equal(b)
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
