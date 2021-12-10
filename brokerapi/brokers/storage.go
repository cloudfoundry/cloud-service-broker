package brokers

import "github.com/cloudfoundry-incubator/cloud-service-broker/internal/storage"

type Storage interface {
	CreateServiceBindingCredentials(binding storage.ServiceBindingCredentials) error
	GetServiceBindingCredentials(bindingID, serviceInstanceID string) (storage.ServiceBindingCredentials, error)
	ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error)
	DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error
}
