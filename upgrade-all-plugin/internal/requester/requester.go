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

func (r Requester) Get(url string, reciever any) error {
	if reflect.TypeOf(reciever).Kind() != reflect.Ptr {
		return fmt.Errorf("reciever must be of type Pointer")
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", r.APIBaseURL, url), nil)
	if err != nil {
		return fmt.Errorf("TODO")
	}

	request.Header.Set("Authorization", r.APIToken)

	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("TODO")
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("TODO")
	}

	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("TODO")
	}

	err = json.Unmarshal(data, &reciever)
	if err != nil {
		return fmt.Errorf("TODO")
	}
	
	return nil
}
