package brokers

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"
)

type Storage interface {
	CreateServiceBindingCredentials(binding storage.ServiceBindingCredentials) error
	GetServiceBindingCredentials(bindingID, serviceInstanceID string) (storage.ServiceBindingCredentials, error)
	ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error)
	DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error
	StoreProvisionRequestDetails(serviceInstanceID string, details json.RawMessage) error
	GetProvisionRequestDetails(serviceInstanceID string) (json.RawMessage, error)
	DeleteProvisionRequestDetails(serviceInstanceID string) error
}
