package paramparser

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/brokerapi/v13/domain"
)

type BindDetails struct {
	AppGUID            string
	CredentialClientID string
	PlanID             string
	ServiceID          string
	RequestParams      map[string]any
	RequestContext     map[string]any
}

func ParseBindDetails(input domain.BindDetails) (BindDetails, error) {
	result := BindDetails{
		AppGUID:   input.AppGUID,
		PlanID:    input.PlanID,
		ServiceID: input.ServiceID,
	}

	if input.BindResource != nil {
		result.AppGUID = input.BindResource.AppGuid
		result.CredentialClientID = input.BindResource.CredentialClientID
	}

	if len(input.RawParameters) > 0 {
		if err := json.Unmarshal(input.RawParameters, &result.RequestParams); err != nil {
			return BindDetails{}, fmt.Errorf("error parsing request parameters: %w", err)
		}
	}

	if len(input.RawContext) > 0 {
		if err := json.Unmarshal(input.RawContext, &result.RequestContext); err != nil {
			return BindDetails{}, fmt.Errorf("error parsing request context: %w", err)
		}
	}

	return result, nil
}
