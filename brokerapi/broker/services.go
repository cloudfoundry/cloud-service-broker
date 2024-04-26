package broker

import (
	"context"
	"fmt"

	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/broker"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

// Services lists services in the broker's catalog.
// It is called through the `GET /v2/catalog` endpoint or the `cf marketplace` command.
func (broker *ServiceBroker) Services(_ context.Context) ([]domain.Service, error) {
	var svcs []domain.Service

	enabledServices, err := broker.registry.GetEnabledServices()
	if err != nil {
		return nil, err
	}
	for _, service := range enabledServices {
		entry := service.CatalogEntry()
		svcs = append(svcs, entry.ToPlain())
	}

	return svcs, nil
}

func (broker *ServiceBroker) getDefinitionAndProvider(serviceID string) (*broker.ServiceDefinition, broker.ServiceProvider, error) {
	defn, err := broker.registry.GetServiceByID(serviceID)
	if err != nil {
		return nil, nil, err
	}

	providerBuilder := defn.ProviderBuilder(broker.Logger, broker.store)
	return defn, providerBuilder, nil
}

func (broker *ServiceBroker) getServiceName(def *broker.ServiceDefinition) string {
	return def.Name
}

func getCredentialName(serviceName, bindingID string) string {
	return fmt.Sprintf("/c/%s/%s/%s/secrets-and-services", credhubClientIdentifier, serviceName, bindingID)
}
