package requester

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

type Requester struct {
	APIBaseURL string
	APIToken   string
	client     *http.Client
}

func NewRequester(apiBaseURL, apiToken string, insecureSkipVerify bool) Requester {
	return Requester{
		APIBaseURL: apiBaseURL,
		APIToken:   apiToken,
		client: &http.Client{
			Timeout: time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
			},
		},
	}
}

func (r Requester) Get(url string, receiver any) error {
	if reflect.TypeOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf("receiver must be of type Pointer")
	}
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", r.APIBaseURL, url), nil)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}
	request.Header.Set("Authorization", r.APIToken)

	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("http request error: %s", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("http response: %d", response.StatusCode)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read http response body error: %s", err)
	}
	err = json.Unmarshal(data, &receiver)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response into receiver error: %s", err)
	}

	return nil
}

func (r Requester) Patch(url string, data any) error {
	d, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}
	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s", r.APIBaseURL, url), bytes.NewReader(d))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}
	request.Header.Set("Authorization", r.APIToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("http request error: %s", err)
	}
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http response: %d", response.StatusCode)
	}

	return nil
}
