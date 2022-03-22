package broker

import (
	"github.com/cloudfoundry/cloud-service-broker/pkg/broker"

	"github.com/cloudfoundry/cloud-service-broker/internal/storage"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Storage

type Storage interface {
	broker.ServiceProviderStorage

	CreateServiceBindingCredentials(binding storage.ServiceBindingCredentials) error
	GetServiceBindingCredentials(bindingID, serviceInstanceID string) (storage.ServiceBindingCredentials, error)
	ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error)
	DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error
	StoreBindRequestDetails(bindRequestDetails storage.BindRequestDetails) error
	GetBindRequestDetails(bindingID, instanceID string) (storage.JSONObject, error)
	DeleteBindRequestDetails(bindingID, instanceID string) error
	StoreProvisionRequestDetails(serviceInstanceID string, details storage.JSONObject) error
	GetProvisionRequestDetails(serviceInstanceID string) (storage.JSONObject, error)
	DeleteProvisionRequestDetails(serviceInstanceID string) error
	StoreServiceInstanceDetails(d storage.ServiceInstanceDetails) error
	GetServiceInstanceDetails(guid string) (storage.ServiceInstanceDetails, error)
	ExistsServiceInstanceDetails(guid string) (bool, error)
	DeleteServiceInstanceDetails(guid string) error
}
