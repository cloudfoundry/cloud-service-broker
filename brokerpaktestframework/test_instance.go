package brokerpaktestframework

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cloudfoundry/cloud-service-broker/v2/internal/testdrive"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
)

type TestInstance struct {
	brokerBuild string
	workspace   string
	broker      *testdrive.Broker
}

func (instance *TestInstance) Start(logger io.Writer, config []string) error {
	file, err := os.CreateTemp("", "test-db")
	if err != nil {
		return err
	}

	fmt.Printf("Starting broker on workspace %s, with build %s\n", instance.workspace, instance.brokerBuild)
	broker, err := testdrive.StartBroker(instance.brokerBuild, instance.workspace, file.Name(), testdrive.WithEnv(config...))
	if err != nil {
		return err
	}
	instance.broker = broker

	return nil
}

func (instance *TestInstance) Catalog() (*apiresponses.CatalogResponse, error) {
	resp := &apiresponses.CatalogResponse{}
	catalogResponse := instance.broker.Client.Catalog(uuid.New())
	switch {
	case catalogResponse.Error != nil:
		return nil, catalogResponse.Error
	case catalogResponse.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("unexpected status code %d: %s", catalogResponse.StatusCode, catalogResponse.ResponseBody)
	default:
		return resp, json.Unmarshal(catalogResponse.ResponseBody, resp)
	}
}

func (instance *TestInstance) withCatalogLookup(serviceName, planName string, cb func(string, string) error) error {
	catalog, err := instance.Catalog()
	if err != nil {
		return err
	}
	serviceGUID, planGUID, err := FindServicePlanGUIDs(catalog, serviceName, planName)
	if err != nil {
		return err
	}

	return cb(serviceGUID, planGUID)
}

func (instance *TestInstance) Provision(serviceName string, planName string, params map[string]any) (string, error) {
	instanceID := uuid.New()

	err := instance.withCatalogLookup(serviceName, planName, func(serviceID, planID string) error {
		_, err := instance.broker.Provision(serviceID, planID, testdrive.WithProvisionServiceInstanceGUID(instanceID), testdrive.WithProvisionParams(params))
		return err
	})
	if err != nil {
		return "", err
	}

	return instanceID, nil
}

func (instance *TestInstance) Update(instanceGUID string, serviceName string, planName string, params map[string]any) error {
	return instance.withCatalogLookup(serviceName, planName, func(serviceID, planID string) error {
		s := testdrive.ServiceInstance{
			GUID:                instanceGUID,
			ServicePlanGUID:     planID,
			ServiceOfferingGUID: serviceID,
		}
		return instance.broker.UpdateService(s, testdrive.WithUpdateParams(params))
	})
}

func (instance *TestInstance) Bind(serviceName, planName, instanceID string, params map[string]any) (map[string]any, error) {
	var (
		binding testdrive.ServiceBinding
		err     error
	)
	err = instance.withCatalogLookup(serviceName, planName, func(serviceID, planID string) error {
		s := testdrive.ServiceInstance{
			GUID:                instanceID,
			ServicePlanGUID:     planID,
			ServiceOfferingGUID: serviceID,
		}
		binding, err = instance.broker.CreateBinding(s, testdrive.WithBindingParams(params))
		return err
	})
	if err != nil {
		return nil, err
	}

	var receiver apiresponses.BindingResponse
	if err := json.Unmarshal([]byte(binding.Body), &receiver); err != nil {
		return nil, err
	}

	return receiver.Credentials.(map[string]any), nil
}

func (instance *TestInstance) Cleanup() error {
	instance.broker.Stop()
	err := os.RemoveAll(instance.workspace)
	return err
}
