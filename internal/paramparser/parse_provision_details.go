package paramparser

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/brokerapi/v8/domain"
)

type ProvisionDetails struct {
	ServiceID        string
	PlanID           string
	OrganizationGUID string
	SpaceGUID        string
	RequestParams    map[string]interface{}
	RequestContext   map[string]interface{}
}

func ParseProvisionDetails(input domain.ProvisionDetails) (ProvisionDetails, error) {
	result := ProvisionDetails{
		ServiceID:        input.ServiceID,
		PlanID:           input.PlanID,
		OrganizationGUID: input.OrganizationGUID,
		SpaceGUID:        input.SpaceGUID,
	}

	if len(input.RawParameters) > 0 {
		if err := json.Unmarshal(input.RawParameters, &result.RequestParams); err != nil {
			return ProvisionDetails{}, fmt.Errorf("error parsing request parameters: %w", err)
		}
	}

	if len(input.RawContext) > 0 {
		if err := json.Unmarshal(input.RawContext, &result.RequestContext); err != nil {
			return ProvisionDetails{}, fmt.Errorf("error parsing request context: %w", err)
		}
	}

	if s, ok := result.RequestContext["organization_guid"].(string); ok {
		result.OrganizationGUID = s
	}

	if s, ok := result.RequestContext["space_guid"].(string); ok {
		result.SpaceGUID = s
	}

	return result, nil
}
