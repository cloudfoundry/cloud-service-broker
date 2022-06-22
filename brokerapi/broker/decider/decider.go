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
	requestHasParams := len(details.RequestParams) != 0
	requestHasPlanChange := details.PlanID != "" && details.PlanID != details.PreviousPlanID
	requestHasUpdate := requestHasParams || requestHasPlanChange

	requestHasMI := details.MaintenanceInfoVersion != nil
	requestHasUpgrade := details.MaintenanceInfoVersion != nil && details.PreviousMaintenanceInfoVersion != nil && !details.MaintenanceInfoVersion.Equal(details.PreviousMaintenanceInfoVersion)
	requestIntroducesMI := requestHasMI && details.PreviousMaintenanceInfoVersion == nil
	requestRemovesMI := !requestHasMI && details.PreviousMaintenanceInfoVersion != nil

	planAndRequestMIDiffer := requestHasMI && !details.MaintenanceInfoVersion.Equal(planMaintenanceInfoVersion)

	switch {
	case planAndRequestMIDiffer && planMaintenanceInfoVersion == nil:
		return Failed, apiresponses.ErrMaintenanceInfoNilConflict
	case planAndRequestMIDiffer:
		return Failed, apiresponses.ErrMaintenanceInfoConflict
	case requestHasUpdate && requestHasUpgrade:
		return Failed, errInstanceMustBeUpgradedFirst()
	case requestHasUpdate:
		return Update, nil
	case requestHasUpgrade, requestIntroducesMI, requestRemovesMI:
		return Upgrade, nil
	default:
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
