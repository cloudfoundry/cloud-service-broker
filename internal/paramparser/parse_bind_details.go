package paramparser

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type BindDetails struct {
	AppGUID        string
	PlanID         string
	ServiceID      string
	BindAppGUID    string
	RequestParams  map[string]interface{}
	RequestContext map[string]interface{}
}

func ParseBindDetails(input domain.BindDetails) (BindDetails, error) {
	result := BindDetails{
		AppGUID:   input.AppGUID,
		PlanID:    input.PlanID,
		ServiceID: input.ServiceID,
	}

	if input.BindResource != nil {
		result.BindAppGUID = input.BindResource.AppGuid
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
