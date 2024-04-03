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

// Package client enables code to make OSBAPI calls
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/spf13/viper"
)

const (
	// ClientsBrokerAPIVersion is the minimum supported version of the client.
	// Note: This may need to be changed in the future as we use newer versions
	// of the OSB API, but should be kept near the lower end of the systems we
	// expect to be compatible with to ensure any reverse-compatibility measures
	// put in place work.
	ClientsBrokerAPIVersion = "2.13"
)

// NewClientFromEnv creates a new client from the client configuration properties.
func NewClientFromEnv() (*Client, error) {
	user := viper.GetString("api.user")
	pass := viper.GetString("api.password")
	port := viper.GetInt("api.port")

	viper.SetDefault("api.hostname", "localhost")
	host := viper.GetString("api.hostname")
	return New(user, pass, host, port)
}

// New creates a new OSB Client connected to the given resource.
func New(username, password, hostname string, port int) (*Client, error) {
	pwd := url.QueryEscape(password)
	base := fmt.Sprintf("http://%s:%s@%s:%d/v2/", username, pwd, hostname, port)
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	return &Client{BaseURL: baseURL}, nil
}

type Client struct {
	BaseURL *url.URL
}

// Catalog fetches the service catalog
func (client *Client) Catalog(requestID string) *BrokerResponse {
	return client.makeRequest(http.MethodGet, "catalog", requestID, nil)
}

// Provision creates a new service with the given instanceId, of type serviceId,
// from the plan planId, with additional details provisioningDetails
func (client *Client) Provision(instanceID, serviceID, planID, requestID string, provisioningDetails json.RawMessage) *BrokerResponse {
	provisionURL := fmt.Sprintf("service_instances/%s?accepts_incomplete=true", instanceID)

	return client.makeRequest(http.MethodPut, provisionURL, requestID, domain.ProvisionDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: provisioningDetails,
	})
}

// Deprovision destroys a service instance of type instanceId
func (client *Client) Deprovision(instanceID, serviceID, planID, requestID string) *BrokerResponse {
	deprovisionURL := fmt.Sprintf("service_instances/%s?accepts_incomplete=true&service_id=%s&plan_id=%s", instanceID, serviceID, planID)

	return client.makeRequest(http.MethodDelete, deprovisionURL, requestID, nil)
}

// Bind creates an account identified by bindingId and gives it access to instanceId
func (client *Client) Bind(instanceID, bindingID, serviceID, planID, requestID string, parameters json.RawMessage) *BrokerResponse {
	bindURL := fmt.Sprintf("service_instances/%s/service_bindings/%s", instanceID, bindingID)

	return client.makeRequest(http.MethodPut, bindURL, requestID, domain.BindDetails{
		ServiceID:     serviceID,
		PlanID:        planID,
		RawParameters: parameters,
	})
}

// Unbind destroys an account identified by bindingId
func (client *Client) Unbind(instanceID, bindingID, serviceID, planID, requestID string) *BrokerResponse {
	unbindURL := fmt.Sprintf("service_instances/%s/service_bindings/%s?service_id=%s&plan_id=%s", instanceID, bindingID, serviceID, planID)

	return client.makeRequest(http.MethodDelete, unbindURL, requestID, nil)
}

// Update sends a patch request to change the plan
func (client *Client) Update(instanceID, serviceID, planID, requestID string, parameters json.RawMessage, previousValues domain.PreviousValues, maintenanceInfo *domain.MaintenanceInfo) *BrokerResponse {
	updateURL := fmt.Sprintf("service_instances/%s?accepts_incomplete=true", instanceID)

	return client.makeRequest(http.MethodPatch, updateURL, requestID, domain.UpdateDetails{
		ServiceID:       serviceID,
		PlanID:          planID,
		RawParameters:   parameters,
		PreviousValues:  previousValues,
		MaintenanceInfo: maintenanceInfo,
	})
}

// LastOperation queries the status of a long-running job on the server
func (client *Client) LastOperation(instanceID, requestID string) *BrokerResponse {
	lastOperationURL := fmt.Sprintf("service_instances/%s/last_operation", instanceID)

	return client.makeRequest(http.MethodGet, lastOperationURL, requestID, nil)
}

func (client *Client) makeRequest(method, path, requestID string, body any) *BrokerResponse {
	br := BrokerResponse{}

	req, err := client.newRequest(method, path, requestID, body)
	br.UpdateRequest(req)
	br.UpdateError(err)
	if br.InError() {
		return &br
	}

	resp, err := http.DefaultClient.Do(req)

	br.UpdateResponse(resp)
	br.UpdateError(err)

	return &br
}

func (client *Client) newRequest(method, path, requestID string, body any) (*http.Request, error) {
	requestURL, err := client.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var buffer io.ReadWriter
	if body != nil {
		buffer = new(bytes.Buffer)
		enc := json.NewEncoder(buffer)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(method, requestURL.String(), buffer)
	if err != nil {
		return nil, err
	}

	// brokerapi v7 does not support the OSBAPI 'X-Broker-API-Request-Identity' header
	request.Header.Set("X-Correlation-ID", requestID)
	request.Header.Set("X-Broker-Api-Version", ClientsBrokerAPIVersion)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return request, nil
}
