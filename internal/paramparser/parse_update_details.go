package paramparser

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type UpdateDetails struct {
	ServiceID         string
	PlanID            string
	PreviousPlanID    string
	PreviousServiceID string
	PreviousOrgID     string
	PreviousSpaceID   string
	RequestParams     map[string]interface{}
	RequestContext    map[string]interface{}
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

	return result, nil
}
