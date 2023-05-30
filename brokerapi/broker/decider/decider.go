// Package decider works out whether a service instance update is an Upgrade or and Update
package decider

import (
	"errors"
	"net/http"

	"github.com/hashicorp/go-version"
	"github.com/pivotal-cf/brokerapi/v10/domain/apiresponses"

	"github.com/cloudfoundry/cloud-service-broker/internal/paramparser"
)

type Operation int

const (
	Failed Operation = iota
	Update
	Upgrade
)

const upgradeBeforeUpdateError = "service instance needs to be upgraded before updating"

// DecideOperation works out whether the platform (typically CloudFoundry) is trying to perform an Update
// of plan/parameters, or an Upgrade of Maintenance Info version. Is it an error to perform both at the
// same time. CloudFoundry provides only the fields it intends to change in the request body,
// but provides previous values for all fields. Where there is ambiguity, we rely on this CloudFoundry behavior.
func DecideOperation(serviceMaintenanceInfoVersion *version.Version, details paramparser.UpdateDetails) (Operation, error) {
	requestHasMI := details.MaintenanceInfoVersion != nil
	requestHasPreviousMI := details.PreviousMaintenanceInfoVersion != nil

	requestHasParams := len(details.RequestParams) != 0
	requestHasPlanChange := details.PlanID != "" && details.PlanID != details.PreviousPlanID
	requestHasUpdate := requestHasParams || requestHasPlanChange

	// There's an update to plan/parameters, new MI, and the new MI does not match the previous MI.
	// So it's an attempt to upgrade and update at the same time.
	invalidUpdateAndMIChange := requestHasUpdate && requestHasMI && !details.MaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion)

	// When there is no MI in the request, this might be because the MI is not being changed (Update),
	// or it might be because MI is being removed (which would be an Upgrade)
	// Here we have:
	// - an update to plan/parameters (which would be invalid to combine with an Upgrade)
	// - no new MI, which could imply no change to MI (as in Update of plan/params), or could be a removal of MI (for Upgrade)
	// - previous MI, so the instance currently has a maintenance info version
	// - no plan MI, so the requested MI and service MI match - hence it would not trigger the ErrMaintenanceInfoConflict error
	// This is an error because it's a combined Update and Upgrade. The valid MI change should be performed first,
	// and then the other fields can be updated in a later request
	invalidUpdateAndMIRemoval := requestHasUpdate && !requestHasMI && requestHasPreviousMI && serviceMaintenanceInfoVersion == nil

	switch {
	case requestHasMI && serviceMaintenanceInfoVersion == nil:
		// error: new MI is specified in request, but service does not have MI
		return Failed, apiresponses.ErrMaintenanceInfoNilConflict
	case requestHasMI && !serviceMaintenanceInfoVersion.Equal(details.MaintenanceInfoVersion):
		// error: new MI is specified, and doesn't match the service MI
		return Failed, apiresponses.ErrMaintenanceInfoConflict
	case invalidUpdateAndMIChange, invalidUpdateAndMIRemoval:
		// error: invalid mixing of an update with an attempt to change MI
		return Failed, errInstanceMustBeUpgradedFirst()
	case !requestHasMI && !requestHasUpdate && serviceMaintenanceInfoVersion == nil && details.PreviousMaintenanceInfoVersion != nil:
		// MI removal: no MI in request because MI is being removed to match service, and previous MI means it's a valid Upgrade
		return Upgrade, nil
	case requestHasMI && !requestHasUpdate && !details.MaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion):
		// add or change MI: MI changed and no updates, so must be an upgrade
		return Upgrade, nil
	case !requestHasMI && !serviceMaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion):
		// platform out of sync: No new MI, but previous MI and service do not match, so platform out of sync with broker
		return Failed, apiresponses.ErrMaintenanceInfoConflict
	default:
		// It's not an error or an Upgrade, so must be an Update
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
