package broker

import (
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/broker"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
)

//go:generate go tool counterfeiter -generate
//counterfeiter:generate . Storage

type Storage interface {
	broker.ServiceProviderStorage

	StoreProvisionRequestDetails(serviceInstanceID string, details storage.JSONObject) error
	GetProvisionRequestDetails(serviceInstanceID string) (storage.JSONObject, error)
	DeleteProvisionRequestDetails(serviceInstanceID string) error

	StoreServiceInstanceDetails(d storage.ServiceInstanceDetails) error
	GetServiceInstanceDetails(guid string) (storage.ServiceInstanceDetails, error)
	ExistsServiceInstanceDetails(guid string) (bool, error)
	DeleteServiceInstanceDetails(guid string) error

	StoreBindRequestDetails(bindingID, instanceID string, bindResource, parameters storage.JSONObject) error
	GetBindRequestDetails(bindingID, instanceID string) (storage.BindRequestDetails, error)
	DeleteBindRequestDetails(bindingID, instanceID string) error

	CreateServiceBindingCredentials(binding storage.ServiceBindingCredentials) error
	GetServiceBindingCredentials(bindingID, serviceInstanceID string) (storage.ServiceBindingCredentials, error)
	ExistsServiceBindingCredentials(bindingID, serviceInstanceID string) (bool, error)
	DeleteServiceBindingCredentials(bindingID, serviceInstanceID string) error
}
