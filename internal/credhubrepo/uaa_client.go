package credhubrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/correlation"
)

type uaaClient struct {
	httpClient *http.Client
	url        string
	name       string
	secret     string
	logger     lager.Logger
}

func (u uaaClient) oauthToken(ctx context.Context) (token, error) {
	requestBody := make(url.Values)
	requestBody.Add("client_id", u.name)
	requestBody.Add("client_secret", u.secret)
	requestBody.Add("grant_type", "client_credentials")
	requestBody.Add("response_type", "token")

	const path = "/oauth/token"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, u.url+path, strings.NewReader(requestBody.Encode()))
	if err != nil {
		return token{}, fmt.Errorf("error creating http request: %w", err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Accept", "application/json")

	u.logger.Debug("http-request-uaa", correlation.ID(ctx), lager.Data{"path": path})
	response, err := u.httpClient.Do(request)
	if err != nil {
		return token{}, fmt.Errorf("error performing http request: %w", err)
	}
	u.logger.Debug("response-code-uaa", correlation.ID(ctx), lager.Data{"code": response.Status})

	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return token{}, fmt.Errorf("error reading response body: %w", err)
	}

	const expectedCode = http.StatusOK
	if response.StatusCode != expectedCode {
		return token{}, fmt.Errorf("unexpected status code %d, expecting %d, body: %s", response.StatusCode, expectedCode, responseBody)
	}

	var responseReceiver struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(responseBody, &responseReceiver); err != nil {
		return token{}, fmt.Errorf("error parsing response body as JSON: %w", err)
	}

	return token{
		value:  responseReceiver.AccessToken,
		expiry: time.Now().Add(time.Duration(responseReceiver.ExpiresIn) * time.Second),
	}, nil
}
