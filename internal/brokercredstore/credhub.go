package brokercredstore

import (
	"fmt"

	"code.cloudfoundry.org/brokerapi/v13/domain"
	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/credstore"
)

type credHubStore struct {
	credstore credstore.CredStore
}

// Store saves the actual credential in CredHub, replacing the binding credential with the CredHub
// reference to the actual credential. Diego will reverse this process so Apps will only see the
// actual credential.
func (c credHubStore) Store(credentials any, serviceName, bindingID, appGUID string) (any, error) {
	credentialName := computeCredHubReference(serviceName, bindingID)

	_, err := c.credstore.Put(credentialName, credentials)
	if err != nil {
		return domain.Binding{}, fmt.Errorf("unable to put credentials in Credstore: %w", err)
	}

	_, err = c.credstore.AddPermission(credentialName, "mtls-app:"+appGUID, []string{"read"})
	if err != nil {
		return domain.Binding{}, fmt.Errorf("unable to add Credstore permissions to app: %w", err)
	}

	return map[string]any{"credhub-ref": credentialName}, nil
}

// Delete will remove the credential from CredHub, and it is tolerant to failure. Failure tolerance is useful because:
//   - because if some credentials were created when CredHub was not enabled, then we won't incorrectly
//     fail to delete something that we never created
//   - if we re-try to delete a failed binding, the process will be idempotent
func (c credHubStore) Delete(logger lager.Logger, serviceName, bindingID string) {
	credentialName := computeCredHubReference(serviceName, bindingID)

	if err := c.credstore.DeletePermission(credentialName); err != nil {
		logger.Error(fmt.Sprintf("failed to delete permissions on the CredHub key %s", credentialName), err)
	}

	if err := c.credstore.Delete(credentialName); err != nil {
		logger.Error(fmt.Sprintf("failed to delete CredHub key %s", credentialName), err)
	}
}

func computeCredHubReference(serviceName, bindingID string) string {
	const credhubClientIdentifier = "csb"
	return fmt.Sprintf("/c/%s/%s/%s/secrets-and-services", credhubClientIdentifier, serviceName, bindingID)
}
