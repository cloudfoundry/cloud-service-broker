package paramparser

import (
	"encoding/json"
	"fmt"

	"github.com/pivotal-cf/brokerapi/v11/domain"
)

type ProvisionDetails struct {
	ServiceID        string
	PlanID           string
	OrganizationGUID string
	SpaceGUID        string
	RequestParams    map[string]any
	RequestContext   map[string]any
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

	// Special case: These two values are maps and must be removed so the HIL parser
	// can access the rest of the values, which are strings.
	// See: https://github.com/cloudfoundry/cloud-service-broker/issues/1086
	if s, ok := result.RequestContext["platform"].(string); ok {
		if s == "cloudfoundry" {
			delete(result.RequestContext, "organization_annotations")
			delete(result.RequestContext, "space_annotations")
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
