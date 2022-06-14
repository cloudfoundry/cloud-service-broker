package main

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

func NewRequester(apiBaseURL, apiToken string) Requester {
	return Requester{
		APIBaseURL: apiBaseURL,
		APIToken:   apiToken,
		client: &http.Client{
			Timeout: time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

type Requester struct {
	APIBaseURL string
	APIToken   string
	client     *http.Client
}

func (r Requester) Get(url string, receiver any) {
	if reflect.TypeOf(receiver).Kind() != reflect.Ptr {
		panic("receiver must be a pointer")
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", r.APIBaseURL, url), nil)
	if err != nil {
		panic(err)
	}
	request.Header.Set("Authorization", r.APIToken)

	response, err := r.client.Do(request)
	if err != nil {
		panic(err)
	}
	if response.StatusCode != http.StatusOK {
		panic("request failed")
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, receiver); err != nil {
		panic(err)
	}
}

func (r Requester) Patch(url string, data any) {
	d, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s", r.APIBaseURL, url), bytes.NewReader(d))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Authorization", r.APIToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := r.client.Do(request)
	if err != nil {
		panic(err)
	}
	if response.StatusCode != http.StatusAccepted {
		panic(fmt.Sprintf("request failed: %d", response.StatusCode))
	}
}
