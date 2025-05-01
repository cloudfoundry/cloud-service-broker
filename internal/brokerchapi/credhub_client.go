package brokerchapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
)

type credHubClient struct {
	httpClient *http.Client
	url        string
	token      string
}

func (c credHubClient) do(method, path string, requestBody, responseBody any, code ...int) error {
	var requestBodyReader io.Reader

	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("unable to marshal JSON body: %w", err)
		}
		requestBodyReader = bytes.NewReader(data)
	}

	request, err := http.NewRequest(method, c.url+path, requestBodyReader)
	if err != nil {
		return fmt.Errorf("error creating http request: %w", err)
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))

	// In theory this is only needed when there's a request body, but the previous implementation always added it
	request.Header.Add("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error performing http request: %w", err)
	}

	defer response.Body.Close()
	responseBodyData, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if !slices.Contains(code, response.StatusCode) {
		return fmt.Errorf("unexpected status code %d, expecting %v, body: %s", response.StatusCode, code, responseBodyData)
	}

	if responseBody != nil {
		if err := json.Unmarshal(responseBodyData, responseBody); err != nil {
			return fmt.Errorf("error parsing response body into JSON: %w", err)
		}
	}

	return nil
}
