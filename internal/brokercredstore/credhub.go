package brokercredstore

import (
	"fmt"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/brokerchapi"
)

type credHubStore struct {
	store *brokerchapi.Store
}

// Store saves the actual credential in CredHub, replacing the binding credential with the CredHub
// reference to the actual credential. Diego will reverse this process so Apps will only see the
// actual credential.
func (c credHubStore) Store(credentials any, serviceName, bindingID, appGUID string) (any, error) {
	credentialName := computeCredHubReference(serviceName, bindingID)

	if err := c.store.Save(credentials, credentialName, appGUID); err != nil {
		return nil, fmt.Errorf("failed to save credential %q in CredHub: %w", credentialName, err)
	}

	return map[string]any{"credhub-ref": credentialName}, nil
}

// Delete will remove the credential from CredHub, and it is tolerant to failure. Failure tolerance is useful because:
//   - because if some credentials were created when CredHub was not enabled, then we won't incorrectly
//     fail to delete something that we never created
//   - if we re-try to delete a failed binding, the process will be idempotent
func (c credHubStore) Delete(logger lager.Logger, serviceName, bindingID string) {
	credentialName := computeCredHubReference(serviceName, bindingID)

	if err := c.store.Delete(credentialName); err != nil {
		logger.Error(fmt.Sprintf("failed to delete credential %q from CredHub", credentialName), err)
	}
}

func computeCredHubReference(serviceName, bindingID string) string {
	const credhubClientIdentifier = "csb"
	return fmt.Sprintf("/c/%s/%s/%s/secrets-and-services", credhubClientIdentifier, serviceName, bindingID)
}
