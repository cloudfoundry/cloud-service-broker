package paramparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/brokerapi/v13/domain/apiresponses"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
)

var ErrNoAppGUIDOrCredentialClient = apiresponses.NewFailureResponse(
	errors.New("no app GUID or credential client ID were provided in the binding request"),
	http.StatusUnprocessableEntity,
	"no-app-guid-or-credential-client-id",
)

type BindDetails struct {
	AppGUID             string
	CredentialClientID  string
	PlanID              string
	ServiceID           string
	CredHubActor        string
	RequestParams       map[string]any
	RequestContext      map[string]any
	RequestBindResource map[string]any
}

func ParseBindDetails(input domain.BindDetails) (BindDetails, error) {
	result := BindDetails{
		PlanID:    input.PlanID,
		ServiceID: input.ServiceID,
	}

	result.RequestBindResource = parseBindResource(input.BindResource)

	// get app_guid from bind_resource or fallback to top-level app_guid
	if result.RequestBindResource == nil {
		result.RequestBindResource = make(map[string]any)
	}
	if val, ok := result.RequestBindResource["app_guid"]; ok {
		// use app_guid from bind resource if present
		result.AppGUID = val.(string)
	} else if input.AppGUID != "" {
		// if app_guid is not present in bind_resource and top-level app_guid is not empty use that
		result.AppGUID = input.AppGUID
		result.RequestBindResource["app_guid"] = input.AppGUID
	} else {
		result.AppGUID = ""
		result.RequestBindResource["app_guid"] = ""
	}

	// set credentialClientID
	if val, ok := result.RequestBindResource["credential_client_id"]; ok {
		result.CredentialClientID = val.(string)
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

func parseBindResource(input *domain.BindResource) map[string]any {

	if input == nil {
		return nil
	}

	if input.AppGuid == "" && input.SpaceGuid == "" && input.Route == "" && input.CredentialClientID == "" && !input.BackupAgent {
		return nil
	}

	result := make(map[string]any)
	if input.AppGuid != "" {
		result["app_guid"] = input.AppGuid
	}
	if input.SpaceGuid != "" {
		result["space_guid"] = input.SpaceGuid
	}
	if input.Route != "" {
		result["route"] = input.Route
	}
	if input.CredentialClientID != "" {
		result["credential_client_id"] = input.CredentialClientID
	}
	if input.BackupAgent {
		result["backup_agent"] = input.BackupAgent
	}
	return result
}

func invalidUserInputError(format string, a ...any) error {
	return apiresponses.NewFailureResponse(
		fmt.Errorf(format, a...),
		http.StatusBadRequest,
		"parsing-user-request",
	)
}

func ParseStoredBindRequestDetails(storedDetails storage.BindRequestDetails, PlanID, ServiceID string) (BindDetails, error) {

	// never set as values are not stored at the database:
	// - CredentialClientID
	// - CredHubActor
	// - RequestContext
	details := BindDetails{
		PlanID:              PlanID,
		ServiceID:           ServiceID,
		RequestParams:       storedDetails.Parameters,
		RequestBindResource: storedDetails.BindResource,
	}

	// get app_guid from bind_resource
	if appID, ok := storedDetails.BindResource["app_guid"].(string); ok {
		details.AppGUID = appID
	}

	return details, nil
}
