// Copyright 2018 the Service Broker Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/cloud-service-broker/pkg/broker"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

// RunExamplesForService runs all the examples for a given service name against
// the service broker pointed to by client. All examples in the registry get run
// if serviceName is blank. If exampleName is non-blank then only the example
// with the given name is run.
func RunExamplesForService(allExamples []CompleteServiceExample, client *Client, serviceName, exampleName string, jobCount int) {
	runExamples(jobCount, client, FilterMatchingServiceExamples(allExamples, serviceName, exampleName))
}

// RunExamplesFromFile reads a json-encoded list of CompleteServiceExamples.
// All examples in the list get run if serviceName is blank. If exampleName
// is non-blank then only the example with the given name is run.
func RunExamplesFromFile(client *Client, fileName, serviceName, exampleName string) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	var allExamples []CompleteServiceExample
	json.Unmarshal(data, &allExamples)

	runExamples(1, client, FilterMatchingServiceExamples(allExamples, serviceName, exampleName))
}

func runExamples(workers int, client *Client, examples []CompleteServiceExample) {
	type result struct {
		id, name, service string
		duration          time.Duration
		err               error
	}
	var results []result
	var resultsLock sync.Mutex
	addResult := func(r result) {
		resultsLock.Lock()
		defer resultsLock.Unlock()
		results = append(results, r)
	}

	type work struct {
		id      string
		example CompleteServiceExample
	}
	queue := make(chan work)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			for w := range queue {
				start := time.Now()
				err := runExample(client, w.id, w.example)
				addResult(result{
					id:       w.id,
					name:     w.example.Name,
					service:  w.example.ServiceName,
					duration: time.Since(start),
					err:      err,
				})
			}
			wg.Done()
		}()
	}

	for i, e := range examples {
		queue <- work{
			id:      fmt.Sprintf("%03d", i),
			example: e,
		}
	}
	close(queue)
	wg.Wait()

	failed := 0
	log.Println()
	log.Println("RESULTS:")
	log.Println()
	log.Println("id | name | service | duration | result")
	log.Println("-- | ---- | ------- | -------- | ------")
	for _, r := range results {
		switch r.err {
		case nil:
			log.Printf("%s | %s | %s | %s | PASS\n", r.id, r.name, r.service, r.duration)
		default:
			failed++
			log.Printf("%s | %s | %s | %s | FAILED %s\n", r.id, r.name, r.service, r.duration, r.err)
		}
	}
	log.Println()

	switch failed {
	case 0:
		log.Println("Success")
	default:
		log.Fatalf("FAILED %d examples", failed)
	}
}

type CompleteServiceExample struct {
	broker.ServiceExample `json:",inline"`
	ServiceName           string                 `json:"service_name"`
	ServiceId             string                 `json:"service_id"`
	ExpectedOutput        map[string]interface{} `json:"expected_output"`
}

func GetExamplesForAService(service *broker.ServiceDefinition) ([]CompleteServiceExample, error) {

	var examples []CompleteServiceExample

	for _, example := range service.Examples {
		serviceCatalogEntry := service.CatalogEntry()

		var completeServiceExample = CompleteServiceExample{
			ServiceExample: example,
			ServiceId:      serviceCatalogEntry.ID,
			ServiceName:    service.Name,
			ExpectedOutput: broker.CreateJsonSchema(service.BindOutputVariables),
		}

		examples = append(examples, completeServiceExample)
	}

	return examples, nil
}

// FilterMatchingServiceExamples should not be run example if:
// 1. The service name is specified and does not match the current example's ServiceName
// 2. The service name is specified and matches the current example's ServiceName, and the example name is specified and does not match the current example's ExampleName
func FilterMatchingServiceExamples(allExamples []CompleteServiceExample, serviceName, exampleName string) []CompleteServiceExample {
	var matchingExamples []CompleteServiceExample

	for _, completeServiceExample := range allExamples {

		if (serviceName != "" && serviceName != completeServiceExample.ServiceName) || (exampleName != "" && exampleName != completeServiceExample.ServiceExample.Name) {
			continue
		}

		matchingExamples = append(matchingExamples, completeServiceExample)
	}

	return matchingExamples
}

// RunExample runs a single example against the given service on the broker
// pointed to by client.
func runExample(client *Client, id string, serviceExample CompleteServiceExample) error {
	logger := newExampleLogger(id)
	executor, err := newExampleExecutor(logger, id, client, serviceExample)
	if err != nil {
		return err
	}

	executor.LogTestInfo(logger)

	// Cleanup the test if it fails partway through
	defer func() {
		logger.Println("Cleaning up the environment")
		executor.Unbind()
		executor.Deprovision()
	}()

	if err := executor.Provision(); err != nil {
		logger.Printf("Failed to provision %v: %v", serviceExample.ServiceName, err)
		return err
	}

	bindResponse, bindErr := executor.Bind()
	if bindErr != nil {
		if serviceExample.BindCanFail {
			log.Printf("WARNING: bind failed: %v, but marked 'can fail' so treated as warning.", bindErr)
		} else {
			log.Printf("Failed to bind %v: %v", serviceExample.ServiceName, bindErr)
			return bindErr
		}
	} else if err := executor.Unbind(); err != nil {
		log.Printf("Failed to unbind %v: %v", serviceExample.ServiceName, err)
		return err
	}

	if err := executor.Deprovision(); err != nil {
		log.Printf("Failed to deprovision %v: %v", serviceExample.ServiceName, err)
		return err
	}

	if bindErr == nil {
		// Check that the binding response has the same fields as expected
		var binding domain.Binding
		err = json.Unmarshal(bindResponse, &binding)
		if err != nil {
			return err
		}

		credentialsEntry := binding.Credentials.(map[string]interface{})

		if err := broker.ValidateVariablesAgainstSchema(credentialsEntry, serviceExample.ExpectedOutput); err != nil {
			log.Printf("Error: results don't match JSON Schema: %v", err)
			log.Printf("Schema: %v\n, Actual: %v", serviceExample.ExpectedOutput, credentialsEntry)
			return err
		}
	}

	return nil
}

func retry(timeout, period time.Duration, function func() (tryAgain bool, err error)) error {
	to := time.After(timeout)
	tick := time.NewTicker(period).C

	if tryAgain, err := function(); !tryAgain {
		return err
	}

	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		case <-to:
			return errors.New("timeout while waiting for result")
		case <-tick:
			tryAgain, err := function()

			if !tryAgain {
				return err
			}
		}
	}
}

func newExampleExecutor(logger *exampleLogger, id string, client *Client, serviceExample CompleteServiceExample) (*exampleExecutor, error) {
	provisionParams, err := json.Marshal(serviceExample.ServiceExample.ProvisionParams)
	if err != nil {
		return nil, err
	}

	bindParams, err := json.Marshal(serviceExample.ServiceExample.BindParams)
	if err != nil {
		return nil, err
	}

	return &exampleExecutor{
		Name:       fmt.Sprintf("%s/%s", serviceExample.ServiceName, serviceExample.ServiceExample.Name),
		ServiceId:  serviceExample.ServiceId,
		PlanId:     serviceExample.ServiceExample.PlanId,
		InstanceId: uuid.New(),
		BindingId:  uuid.New(),

		ProvisionParams: provisionParams,
		BindParams:      bindParams,

		logger: logger,
		client: client,
	}, nil
}

type exampleExecutor struct {
	Name string

	ServiceId  string
	PlanId     string
	InstanceId string
	BindingId  string

	ProvisionParams json.RawMessage
	BindParams      json.RawMessage

	logger *exampleLogger
	client *Client
}

// Provision attempts to create a service instance from the example.
// Multiple calls to provision will attempt to create a resource with the same
// ServiceId and details.
// If the response is an async result, Provision will attempt to wait until
// the Provision is complete.
func (ee *exampleExecutor) Provision() error {
	requestID := uuid.New()
	ee.logger.Printf("Provisioning %s (id: %s)\n", ee.Name, requestID)

	resp := ee.client.Provision(ee.InstanceId, ee.ServiceId, ee.PlanId, requestID, ee.ProvisionParams)

	ee.logger.Println(resp.String())
	if resp.InError() {
		return resp.Error
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusAccepted:
		return ee.pollUntilFinished()
	default:
		return fmt.Errorf("unexpected response code %d", resp.StatusCode)
	}
}

func (ee *exampleExecutor) pollUntilFinished() error {
	return retry(45*time.Minute, 30*time.Second, func() (bool, error) {
		requestID := uuid.New()
		ee.logger.Printf("Polling for async job (id: %s)\n", requestID)

		resp := ee.client.LastOperation(ee.InstanceId, requestID)
		if resp.InError() {
			return false, resp.Error
		}

		if resp.StatusCode != http.StatusOK {
			ee.logger.Printf("Bad status code %d, needed 200", resp.StatusCode)
			return false, fmt.Errorf("broker responded with statuscode %v", resp.StatusCode)
		}

		var responseBody map[string]string
		err := json.Unmarshal(resp.ResponseBody, &responseBody)
		if err != nil {
			return false, err
		}

		state := responseBody["state"]
		eq := state == string(brokerapi.Succeeded)

		if state == string(brokerapi.Failed) {
			ee.logger.Printf("Last operation for %q was %q: %s\n", ee.InstanceId, state, responseBody["description"])
			return false, fmt.Errorf(responseBody["description"])
		}

		ee.logger.Printf("Last operation for %q was %q\n", ee.InstanceId, state)
		return !eq, nil
	})
}

// Deprovision destroys the instance created by a call to Provision.
func (ee *exampleExecutor) Deprovision() error {
	requestID := uuid.New()
	ee.logger.Printf("Deprovisioning %s (id: %s)\n", ee.Name, requestID)
	resp := ee.client.Deprovision(ee.InstanceId, ee.ServiceId, ee.PlanId, requestID)

	ee.logger.Println(resp.String())
	if resp.InError() {
		return resp.Error
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusAccepted:
		return ee.pollUntilFinished()
	default:
		return fmt.Errorf("unexpected response code %d", resp.StatusCode)
	}
}

// Unbind unbinds the exact binding created by a call to Bind.
func (ee *exampleExecutor) Unbind() error {
	return retry(15*time.Minute, 15*time.Second, func() (bool, error) {
		requestID := uuid.New()
		ee.logger.Printf("Unbinding %s (id: %s)\n", ee.Name, requestID)
		resp := ee.client.Unbind(ee.InstanceId, ee.BindingId, ee.ServiceId, ee.PlanId, requestID)

		ee.logger.Println(resp.String())
		if resp.InError() {
			return false, resp.Error
		}

		if resp.StatusCode == http.StatusOK {
			return false, nil
		}

		return false, fmt.Errorf("unexpected response code %d", resp.StatusCode)
	})
}

// Bind executes the bind portion of the create, this can only be called
// once successfully as subsequent binds will attempt to create bindings with
// the same ID.
func (ee *exampleExecutor) Bind() (json.RawMessage, error) {
	requestID := uuid.New()
	ee.logger.Printf("Binding %s (id: %s)\n", ee.Name, requestID)
	resp := ee.client.Bind(ee.InstanceId, ee.BindingId, ee.ServiceId, ee.PlanId, requestID, ee.BindParams)

	ee.logger.Println(resp.String())
	if resp.InError() {
		return nil, resp.Error
	}

	if resp.StatusCode == http.StatusCreated {
		return resp.ResponseBody, nil
	}

	return nil, fmt.Errorf("unexpected response code %d", resp.StatusCode)
}

// LogTestInfo writes information about the running example and a manual backout
// strategy if the test dies part of the way through.
func (ee *exampleExecutor) LogTestInfo(logger *exampleLogger) {
	logger.Printf("Running Example: %s\n", ee.Name)

	ips := fmt.Sprintf("--instanceid %q --planid %q --serviceid %q", ee.InstanceId, ee.PlanId, ee.ServiceId)
	logger.Printf("cloud-service-broker client provision %s --params %q\n", ips, ee.ProvisionParams)
	logger.Printf("cloud-service-broker client bind %s --bindingid %q --params %q\n", ips, ee.BindingId, ee.BindParams)
	logger.Printf("cloud-service-broker client unbind %s --bindingid %q\n", ips, ee.BindingId)
	logger.Printf("cloud-service-broker client deprovision %s\n", ips)
}
