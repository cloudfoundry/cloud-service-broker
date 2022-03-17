package paramparser

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type BindDetails struct {
	AppGUID                string
	PlanID                 string
	ServiceID              string
	BindAppGuid            string
	BindSpaceGuid          string
	BindRoute              string
	BindCredentialClientID string
	BindBackupAgent        bool
	RequestParams          map[string]interface{}
	RequestContext         map[string]interface{}
}

func ParseBindDetails(input domain.BindDetails) (BindDetails, error) {
	if input.BindResource == nil {
		return BindDetails{}, fmt.Errorf("error parsing bind request details: missing bind_resource")
	}

	result := BindDetails{
		AppGUID:                input.AppGUID,
		PlanID:                 input.PlanID,
		ServiceID:              input.ServiceID,
		BindAppGuid:            input.BindResource.AppGuid,
		BindSpaceGuid:          input.BindResource.SpaceGuid,
		BindRoute:              input.BindResource.Route,
		BindCredentialClientID: input.BindResource.CredentialClientID,
		BindBackupAgent:        input.BindResource.BackupAgent,
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

func ParseUnbindDetails(input domain.UnbindDetails) (BindDetails, error) {
	result := BindDetails{
		PlanID:    input.PlanID,
		ServiceID: input.ServiceID,
	}

	return result, nil
}
