package paramparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
)

var ErrNoAppGUIDOrCredentialClient = apiresponses.NewFailureResponse(
	errors.New("no app GUID or credential client ID were provided in the binding request"),
	http.StatusUnprocessableEntity,
	"no-app-guid-or-credential-client-id",
)

type BindDetails struct {
	AppGUID            string
	CredentialClientID string
	PlanID             string
	ServiceID          string
	CredHubActor       string
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
		result.CredentialClientID = input.BindResource.CredentialClientID

		if input.BindResource.AppGuid != "" {
			result.AppGUID = input.BindResource.AppGuid
		}
	}

	switch {
	case result.AppGUID != "":
		result.CredHubActor = fmt.Sprintf("mtls-app:%s", result.AppGUID)
	case result.CredentialClientID != "":
		result.CredHubActor = fmt.Sprintf("uaa-client:%s", result.CredentialClientID)
	default:
		return BindDetails{}, ErrNoAppGUIDOrCredentialClient
	}

	if len(input.RawParameters) > 0 {
		if err := json.Unmarshal(input.RawParameters, &result.RequestParams); err != nil {
			return BindDetails{}, invalidUserInputError("error parsing request parameters: %w", err)
		}
	}

	if len(input.RawContext) > 0 {
		if err := json.Unmarshal(input.RawContext, &result.RequestContext); err != nil {
			return BindDetails{}, invalidUserInputError("error parsing request context: %w", err)
		}
	}

	return result, nil
}

func invalidUserInputError(format string, a ...any) error {
	return apiresponses.NewFailureResponse(
		fmt.Errorf(format, a...),
		http.StatusBadRequest,
		"parsing-user-request",
	)
}
