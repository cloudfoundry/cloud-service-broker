package paramparser

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/pivotal-cf/brokerapi/v12/domain"
)

type UpdateDetails struct {
	ServiceID                      string
	PlanID                         string
	MaintenanceInfoVersion         *version.Version
	PreviousPlanID                 string
	PreviousServiceID              string
	PreviousOrgID                  string
	PreviousSpaceID                string
	PreviousMaintenanceInfoVersion *version.Version
	RequestParams                  map[string]any
	RequestContext                 map[string]any
}

func ParseUpdateDetails(input domain.UpdateDetails) (UpdateDetails, error) {
	result := UpdateDetails{
		ServiceID:         input.ServiceID,
		PlanID:            input.PlanID,
		PreviousPlanID:    input.PreviousValues.PlanID,
		PreviousServiceID: input.PreviousValues.ServiceID,
		PreviousOrgID:     input.PreviousValues.OrgID,
		PreviousSpaceID:   input.PreviousValues.SpaceID,
	}

	if len(input.RawParameters) > 0 {
		if err := json.Unmarshal(input.RawParameters, &result.RequestParams); err != nil {
			return UpdateDetails{}, fmt.Errorf("error parsing request parameters: %w", err)
		}
	}

	if len(input.RawContext) > 0 {
		if err := json.Unmarshal(input.RawContext, &result.RequestContext); err != nil {
			return UpdateDetails{}, fmt.Errorf("error parsing request context: %w", err)
		}
	}

	if input.MaintenanceInfo != nil && len(input.MaintenanceInfo.Version) != 0 {
		v, err := version.NewVersion(input.MaintenanceInfo.Version)
		if err != nil {
			return UpdateDetails{}, fmt.Errorf("error parsing maintenance info: %w", err)
		}
		result.MaintenanceInfoVersion = v
	}

	if input.PreviousValues.MaintenanceInfo != nil && len(input.PreviousValues.MaintenanceInfo.Version) != 0 {
		v, err := version.NewVersion(input.PreviousValues.MaintenanceInfo.Version)
		if err != nil {
			return UpdateDetails{}, fmt.Errorf("error parsing previous maintenance info: %w", err)
		}
		result.PreviousMaintenanceInfoVersion = v
	}

	return result, nil
}
