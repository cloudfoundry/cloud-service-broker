// Package brokercredstore manages the storing of binding credentials in CredHub when enabled
package brokercredstore

import (
	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/brokerchapi"
)

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . BrokerCredstore
type BrokerCredstore interface {
	Store(credentials any, serviceName, bindingID, appGUID string) (any, error)
	Delete(logger lager.Logger, serviceName, bindingID string)
}

func NewBrokerCredstore(store *brokerchapi.Store) BrokerCredstore {
	switch store {
	case nil:
		return noopStore{}
	default:
		return credHubStore{store: store}
	}
}
