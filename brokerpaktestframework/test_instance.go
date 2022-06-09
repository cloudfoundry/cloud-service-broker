package brokerpaktestframework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/onsi/gomega/gexec"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pivotal-cf/brokerapi/v8/domain"
	"github.com/pivotal-cf/brokerapi/v8/domain/apiresponses"
)

type TestInstance struct {
	brokerBuild   string
	workspace     string
	password      string
	username      string
	port          string
	serverSession *gexec.Session
}

func (instance *TestInstance) Start(logger io.Writer, config []string) error {
	file, err := os.CreateTemp("", "test-db")
	if err != nil {
		return err
	}
	serverCommand := exec.Command(instance.brokerBuild)
	serverCommand.Dir = instance.workspace
	serverCommand.Env = append([]string{
		"DB_PATH=" + file.Name(),
		"DB_TYPE=sqlite3",
		"PORT=" + instance.port,
		"SECURITY_USER_NAME=" + instance.username,
		"SECURITY_USER_PASSWORD=" + instance.password,
	}, config...)
	fmt.Printf("Starting broker on workspace %s, with build %s, with env %v", instance.workspace, instance.brokerBuild, serverCommand.Env)
	start, err := gexec.Start(serverCommand, logger, logger)
	if err != nil {
		return err
	}
	instance.serverSession = start

	return waitForHTTPServer("http://localhost:" + instance.port)
}

func waitForHTTPServer(s string) error {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 4
	retryClient.RetryWaitMin = time.Millisecond * 200
	retryClient.RetryWaitMax = time.Millisecond * 1000
	_, err := retryClient.Get(s)
	return err
}

func (instance *TestInstance) Provision(serviceName string, planName string, params map[string]interface{}) (string, error) {
	instanceID, resp, err := instance.provision(serviceName, planName, params)
	if err != nil {
		return "", err
	}

	return instanceID, instance.pollLastOperation("service_instances/"+instanceID+"/last_operation", resp.OperationData)
}

func (instance *TestInstance) Update(instanceGUID string, serviceName string, planName string, params map[string]interface{}) error {
	resp, err := instance.update(instanceGUID, serviceName, planName, params)

	if err != nil {
		return err
	}

	return instance.pollLastOperation("service_instances/"+instanceGUID+"/last_operation", resp.OperationData)
}

func (instance *TestInstance) provision(serviceName string, planName string, params map[string]interface{}) (string, *apiresponses.ProvisioningResponse, error) {
	instanceID := uuid.New().String()

	catalog, err := instance.Catalog()
	if err != nil {
		return "", nil, err
	}
	serviceGUID, planGUID := FindServicePlanGUIDs(catalog, serviceName, planName)
	details := domain.ProvisionDetails{
		ServiceID: serviceGUID,
		PlanID:    planGUID,
	}
	if params != nil {
		data, err := json.Marshal(&params)
		if err != nil {
			return "", nil, err
		}
		details.RawParameters = json.RawMessage(data)
	}
	data, err := json.Marshal(details)
	if err != nil {
		return "", nil, err
	}
	body, status, err := instance.httpInvokeBroker("service_instances/"+instanceID+"?accepts_incomplete=true", "PUT", bytes.NewBuffer(data))
	if err != nil {
		return "", nil, err
	}
	if status != http.StatusAccepted {
		return "", nil, fmt.Errorf("request failed: status %d: body %s", status, body)
	}

	response := apiresponses.ProvisioningResponse{}

	return instanceID, &response, json.Unmarshal(body, &response)
}

func (instance *TestInstance) update(instanceID, serviceName, planName string, params map[string]interface{}) (*apiresponses.UpdateResponse, error) {
	catalog, err := instance.Catalog()
	if err != nil {
		return nil, err
	}
	serviceGUID, planGUID := FindServicePlanGUIDs(catalog, serviceName, planName)
	details := domain.UpdateDetails{
		ServiceID: serviceGUID,
		PlanID:    planGUID,
	}
	if params != nil {
		data, err := json.Marshal(&params)
		if err != nil {
			return nil, err
		}
		details.RawParameters = json.RawMessage(data)
	}
	data, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}
	body, status, err := instance.httpInvokeBroker("service_instances/"+instanceID+"?accepts_incomplete=true", "PATCH", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if status != http.StatusAccepted {
		return nil, fmt.Errorf("request failed: status %d: body %s", status, body)
	}

	response := apiresponses.UpdateResponse{}

	return &response, json.Unmarshal(body, &response)
}

func (instance *TestInstance) pollLastOperation(pollingURL string, lastOperation string) error {
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out polling %s %s", pollingURL, lastOperation)
		case <-ticker.C:
			data, status, err := instance.httpInvokeBroker(pollingURL, "GET", nil)
			if err != nil {
				return err
			}
			if status != http.StatusOK {
				return fmt.Errorf("request failed: status %d: body %s", status, data)
			}
			resp := apiresponses.LastOperationResponse{}
			err = json.Unmarshal(data, &resp)
			if err != nil {
				return err
			}
			if resp.State != domain.InProgress {
				return nil
			}
		}
	}
}

func (instance *TestInstance) Catalog() (*apiresponses.CatalogResponse, error) {
	catalogJSON, status, err := instance.httpInvokeBroker("catalog", "GET", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("request failed: status %d: body %s", status, catalogJSON)
	}

	resp := &apiresponses.CatalogResponse{}
	return resp, json.Unmarshal(catalogJSON, resp)
}

func (instance *TestInstance) httpInvokeBroker(subpath string, method string, body io.Reader) ([]byte, int, error) {
	client := &http.Client{
		Timeout: time.Second * 0,
	}
	req, err := http.NewRequest(method, instance.BrokerURL(subpath), body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("X-Broker-API-Version", "2.14")
	req.SetBasicAuth(instance.username, instance.password)
	response, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	return contents, response.StatusCode, err
}

func (instance *TestInstance) BrokerURL(subPath string) string {
	return fmt.Sprintf("http://localhost:%s/v2/%s", instance.port, subPath)
}

// BrokerUrl returns the URL of the broker. Use BrokerURL instead.
// Deprecated: due to name that does not conform to Go initialisms:  https://github.com/golang/go/wiki/CodeReviewComments#initialisms
//lint:ignore ST1003 to maintain backwards compatability
func (instance *TestInstance) BrokerUrl(subPath string) string {
	return instance.BrokerURL(subPath)
}

func (instance *TestInstance) Bind(serviceName, planName, instanceID string, params map[string]interface{}) (map[string]interface{}, error) {
	catalog, err := instance.Catalog()
	if err != nil {
		return nil, err
	}
	serviceGUID, planGUID := FindServicePlanGUIDs(catalog, serviceName, planName)

	bindingResult, err := instance.bind(serviceGUID, planGUID, params, instanceID)
	if err != nil {
		return nil, err
	}

	return bindingResult, nil
}

func (instance *TestInstance) bind(serviceGUID, planGUID string, params map[string]interface{}, instanceID string) (map[string]interface{}, error) {
	bindDetails := domain.BindDetails{
		ServiceID: serviceGUID,
		PlanID:    planGUID,
	}

	if params != nil {
		data, err := json.Marshal(&params)
		if err != nil {
			return nil, err
		}
		bindDetails.RawParameters = json.RawMessage(data)
	}
	data, err := json.Marshal(bindDetails)
	if err != nil {
		return nil, err
	}
	bindingID := uuid.New()
	body, status, err := instance.httpInvokeBroker(fmt.Sprintf("service_instances/%s/service_bindings/%s?accepts_incomplete=true", instanceID, bindingID), "PUT", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if !(status == http.StatusCreated) {
		return nil, fmt.Errorf("request failed: status %d: body %s", status, body)
	}
	bindingResponse := apiresponses.BindingResponse{}

	return bindingResponse.Credentials.(map[string]interface{}), json.Unmarshal(body, &bindingResponse)
}

func (instance *TestInstance) Cleanup() error {
	instance.serverSession.Terminate().Wait()
	err := os.RemoveAll(instance.workspace)
	return err
}
