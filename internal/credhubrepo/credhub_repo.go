// Package credhubrepo is a repository pattern for saving and deleting a credential in CredHub
package credhubrepo

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/config"
)

type Repo struct {
	credHubURL string
	httpClient *http.Client
	uaaClient  uaaClient
	token      token // should only be used in loadToken() and fetchToken() that use the tokenLock mutex
	tokenLock  sync.Mutex
}

// New creates a new CredHub repo
func New(cfg config.CredStoreConfig) (*Repo, error) {
	httpClient, err := newHTTPClient(cfg.SkipSSLValidation, cfg.CACert)
	if err != nil {
		return nil, err
	}

	return &Repo{
		credHubURL: cfg.CredHubURL,
		httpClient: httpClient,
		uaaClient: uaaClient{
			httpClient: httpClient,
			url:        cfg.UaaURL,
			name:       cfg.UaaClientName,
			secret:     cfg.UaaClientSecret,
		},
	}, nil
}

// Save will add a credential to CredHub and give the specified app permissions to see the credential
func (r *Repo) Save(path string, cred any, actor string) (any, error) {
	// Set the credential value
	setRequestBody := map[string]any{
		"name":  path,
		"type":  "json",
		"value": cred,
	}

	if err := r.http(http.MethodPut, "/api/v1/data", setRequestBody, nil, http.StatusOK); err != nil {
		return nil, fmt.Errorf("failed to store credential %q: %w", path, err)
	}

	// Update permissions for the credential
	permsRequestBody := map[string]any{
		"path":       path,
		"actor":      actor,
		"operations": []string{"read"},
	}

	if err := r.http(http.MethodPost, "/api/v2/permissions", permsRequestBody, nil, http.StatusCreated); err != nil {
		return nil, fmt.Errorf("failed to set permission on credential %q: %w", path, err)
	}

	return map[string]any{"credhub-ref": path}, nil
}

// Delete will remove a credential and all its permissions from CredHub
// It is idempotent so does not fail if the credential does not exist
func (r *Repo) Delete(path string) error {
	// Delete any existing permissions
	// To replicate previous implementation, we use the V1 permissions API to get all
	// permissions for the credential (not available in V2), and iterate over the
	// actors to delete the permissions. This means that we don't need to store the
	// actors.
	var listPermissionsResponseBody struct {
		Permissions []struct {
			Actor string `json:"actor"`
		} `json:"permissions"`
	}

	if err := r.http(http.MethodGet, fmt.Sprintf("/api/v1/permissions?credential_name=%s", path), nil, &listPermissionsResponseBody, http.StatusOK); err != nil {
		return fmt.Errorf("failed to list permissions for credential %q: %w", path, err)
	}

	for _, p := range listPermissionsResponseBody.Permissions {
		query := url.Values{
			"actor": []string{p.Actor},
			"path":  []string{path},
		}.Encode()

		var getPermissionResponseBody struct {
			UUID string `json:"uuid"`
		}

		if err := r.http(http.MethodGet, fmt.Sprintf("/api/v2/permissions?%s", query), nil, &getPermissionResponseBody, http.StatusOK); err != nil {
			return fmt.Errorf("failed to get permission %q for credential %q: %w", p.Actor, path, err)
		}

		if err := r.http(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", getPermissionResponseBody.UUID), nil, nil, http.StatusOK); err != nil {
			return fmt.Errorf("failed to delete permission ID %q: %w", getPermissionResponseBody.UUID, err)
		}
	}

	// Delete the credential
	if err := r.http(http.MethodDelete, fmt.Sprintf("/api/v1/data?name=%s", path), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to delete credential %q: %w", path, err)
	}
	return nil
}

func (r *Repo) http(method, path string, requestBody, responseBody any, okCodes ...int) error {
	tok, cachedToken, err := r.loadToken()
	if err != nil {
		return err
	}

	// Process request body
	var requestBodyReader io.Reader
	if requestBody != nil {
		data, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("unable to marshal JSON body: %w", err)
		}
		requestBodyReader = bytes.NewReader(data)
	}

	// Do the HTTP request
	request, err := r.newHTTPRequest(method, path, tok, requestBodyReader)
	if err != nil {
		return err
	}

	response, err := r.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("error performing http request: %w", err)
	}

	// If we used a cached token, and we got an authorization error, then try
	// to get another UAA token and retry the request. There's no point retrying
	// if we only just fetched the token.
	if cachedToken && response.StatusCode == http.StatusUnauthorized {
		if tok, err = r.fetchToken(); err != nil {
			return err
		}

		request, err = r.newHTTPRequest(method, path, tok, requestBodyReader)
		if err != nil {
			return err
		}

		response, err = r.httpClient.Do(request)
		if err != nil {
			return fmt.Errorf("error performing http request: %w", err)
		}
	}

	defer response.Body.Close()
	responseBodyData, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	if !slices.Contains(okCodes, response.StatusCode) {
		return fmt.Errorf("unexpected status code %d for CredHub endpoint %q, expecting %v, body: %s", response.StatusCode, path, okCodes, responseBodyData)
	}

	if responseBody != nil {
		if err := json.Unmarshal(responseBodyData, responseBody); err != nil {
			return fmt.Errorf("error parsing response body into JSON: %w", err)
		}
	}

	return nil
}

// loadToken will return an existing unexpired token if there is one, and otherwise will
// fetch a new token (compare with fetchToken())
//
// Why do we use sync.Mutex and not sync.RWMutex? While a RWMutex is arguably more correct
// because simultaneous reading of the token should not be problematic, in practice
// it results in an implementation that's more complicated, and there's no reason to think
// it would result in a performance advantage in actual usage.
func (r *Repo) loadToken() (string, bool, error) {
	r.tokenLock.Lock()
	defer r.tokenLock.Unlock()

	if !r.token.expired() {
		return r.token.value, true, nil
	}

	tok, err := r.uaaClient.oauthToken()
	if err != nil {
		return "", false, err
	}
	r.token = tok

	return tok.value, false, nil
}

// fetchToken will always try to get a new token (compare to loadToken())
func (r *Repo) fetchToken() (string, error) {
	r.tokenLock.Lock()
	defer r.tokenLock.Unlock()

	tok, err := r.uaaClient.oauthToken()
	if err != nil {
		return "", err
	}

	r.token = tok
	return tok.value, nil
}

func (r *Repo) newHTTPRequest(method, path, tok string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, r.credHubURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("error creating http request: %w", err)
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok))

	// In theory this is only needed when there's a request body, but the previous implementation always added it
	request.Header.Add("Content-Type", "application/json")

	return request, nil
}

func newHTTPClient(insecureSkipVerify bool, caCert string) (*http.Client, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	if len(caCert) > 0 {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to get system cert pool: %w", err)
		}

		if !pool.AppendCertsFromPEM([]byte(caCert)) {
			return nil, fmt.Errorf("failed to add CA cert to pool")
		}

		tlsConfig.RootCAs = pool
	}

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tlsConfig},
		Timeout:   time.Minute,
	}, nil
}
