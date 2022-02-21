package brokerpaktestframework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	file, err := ioutil.TempFile("", "test-db")
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

	return waitForHttpServer("http://localhost:" + instance.port)
}

func waitForHttpServer(s string) error {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 4
	retryClient.RetryWaitMin = time.Millisecond * 200
	retryClient.RetryWaitMax = time.Millisecond * 500
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

func (instance *TestInstance) provision(serviceName string, planName string, params map[string]interface{}) (string, *apiresponses.ProvisioningResponse, error) {
	catalog, err := instance.Catalog()
	if err != nil {
		return "", nil, err
	}
	serviceGuid, planGuid := FindServicePlanGUIDs(catalog, serviceName, planName)
	details := domain.ProvisionDetails{
		ServiceID: serviceGuid,
		PlanID:    planGuid,
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
	instanceId := uuid.New()
	body, status, err := instance.httpInvokeBroker("service_instances/"+instanceId.String()+"?accepts_incomplete=true", "PUT", bytes.NewBuffer(data))
	if err != nil {
		return "", nil, err
	}
	if status != http.StatusAccepted {
		return "", nil, fmt.Errorf("request failed: status %d: body %s", status, body)
	}

	response := apiresponses.ProvisioningResponse{}

	return instanceId.String(), &response, json.Unmarshal(body, &response)
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
	catalogJson, status, err := instance.httpInvokeBroker("catalog", "GET", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("request failed: status %d: body %s", status, catalogJson)
	}

	resp := &apiresponses.CatalogResponse{}
	return resp, json.Unmarshal(catalogJson, resp)
}

func (instance *TestInstance) httpInvokeBroker(subpath string, method string, body io.Reader) ([]byte, int, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(method, instance.BrokerUrl(subpath), body)
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

func (instance *TestInstance) BrokerUrl(subPath string) string {
	return fmt.Sprintf("http://localhost:%s/v2/%s", instance.port, subPath)
}

func (instance *TestInstance) Bind(serviceName, planName, instanceID string, params map[string]interface{}) (map[string]interface{}, error) {
	catalog, err := instance.Catalog()
	if err != nil {
		return nil, err
	}
	serviceGuid, planGuid := FindServicePlanGUIDs(catalog, serviceName, planName)

	bindingResult, err := instance.bind(serviceGuid, planGuid, params, instanceID)
	if err != nil {
		return nil, err
	}

	return bindingResult, nil
}

func (instance *TestInstance) bind(serviceGuid, planGuid string, params map[string]interface{}, instanceID string) (map[string]interface{}, error) {
	bindDetails := domain.BindDetails{
		ServiceID: serviceGuid,
		PlanID:    planGuid,
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
	instance.serverSession.Terminate()
	err := os.RemoveAll(instance.workspace)
	return err
}
