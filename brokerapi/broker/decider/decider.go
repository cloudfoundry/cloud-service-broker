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
	requestHasMI := details.MaintenanceInfoVersion != nil
	requestHasPreviousMI := details.PreviousMaintenanceInfoVersion != nil

	requestHasParams := len(details.RequestParams) != 0
	requestHasPlanChange := details.PlanID != "" && details.PlanID != details.PreviousPlanID
	requestHasUpdate := requestHasParams || requestHasPlanChange

	// There's an update, new and previous MI, and the new MI does not match the plan
	invalidUpdateAndMIChange := requestHasUpdate && requestHasMI && requestHasPreviousMI && !details.MaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion)

	// There's an update, new MI, no previous MI, and the new MI does not match the plan
	invalidUpdateAndMIAddition := requestHasUpdate && requestHasMI && !requestHasPreviousMI && !details.MaintenanceInfoVersion.Equal(planMaintenanceInfoVersion)

	// There's an update, no new MI, plan does not have MI, and there's a previous MI value
	invalidUpdateAndMIRemoval := requestHasUpdate && !requestHasMI && requestHasPreviousMI && planMaintenanceInfoVersion == nil

	switch {
	case details.MaintenanceInfoVersion != nil && planMaintenanceInfoVersion == nil:
		// new MI is specified in request, but plan does not have MI
		return Failed, apiresponses.ErrMaintenanceInfoNilConflict
	case details.MaintenanceInfoVersion != nil && !planMaintenanceInfoVersion.Equal(details.MaintenanceInfoVersion):
		// new MI is specified, and doesn't match the plan MI
		return Failed, apiresponses.ErrMaintenanceInfoConflict
	case invalidUpdateAndMIChange, invalidUpdateAndMIAddition, invalidUpdateAndMIRemoval:
		// invalid mixing of an update with an attempt to add, remove or alter MI
		return Failed, errInstanceMustBeUpgradedFirst()
	case !requestHasUpdate && !details.MaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion):
		// MI changed and no updates, so must be an upgrade
		return Upgrade, nil
	default:
		// It's not a valid or invalid upgrade, so must be an update
		return Update, nil
	}
}

func errInstanceMustBeUpgradedFirst() *apiresponses.FailureResponse {
	return apiresponses.NewFailureResponseBuilder(
		errors.New(upgradeBeforeUpdateError),
		http.StatusUnprocessableEntity,
		"previous-maintenance-info-check",
	).Build()
}
