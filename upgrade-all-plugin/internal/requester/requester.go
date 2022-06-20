package requester

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Requester
type Requester interface {
	Get(url string, receiver any) error
}

type HttpRequester struct {
	APIBaseURL string
	APIToken   string
	client     *http.Client
}

func NewRequester(apiBaseURL, apiToken string, insecureSkipVerify bool) HttpRequester {
	return HttpRequester{
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

func (r HttpRequester) Get(url string, receiver any) error {
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
