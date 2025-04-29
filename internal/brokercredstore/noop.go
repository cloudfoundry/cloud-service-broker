package brokercredstore

import "code.cloudfoundry.org/lager/v3"

// noopStore is used when CredHub is not enabled
// It simplifies the code as there's no longer a need for lots of "if"s to check whether CredHub is enabled
type noopStore struct{}

func (noopStore) Store(creds any, serviceName, bindingID, appGUID string) (any, error) {
	return creds, nil
}

func (noopStore) Delete(logger lager.Logger, serviceName, bindingID string) {}
