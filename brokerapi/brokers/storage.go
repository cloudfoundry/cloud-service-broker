package brokers

import (
	"encoding/json"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"

	"github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"
)

type Storage interface {
	broker.ServiceProviderStorage

	CreateServiceBindingCredentials(binding storage.ServiceBindingCredentials) error
	GetServiceBindingCredentials(bindingID, serviceInstanceID string) (storage.ServiceBindingCredentials, error)
	ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error)
	DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error
	StoreProvisionRequestDetails(serviceInstanceID string, details json.RawMessage) error
	GetProvisionRequestDetails(serviceInstanceID string) (json.RawMessage, error)
	DeleteProvisionRequestDetails(serviceInstanceID string) error
	StoreServiceInstanceDetails(d storage.ServiceInstanceDetails) error
	GetServiceInstanceDetails(guid string) (storage.ServiceInstanceDetails, error)
	ExistsServiceInstanceDetails(guid string) (bool, error)
	DeleteServiceInstanceDetails(guid string) error
}
