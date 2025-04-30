// Package brokerchapi has just enough of the CredHub API to allow service broker to store binding credentials
package brokerchapi

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	CredHubURL            string
	CACert                string
	UAAURL                string
	UAAClientName         string
	UAAClientSecret       string
	InsecureSkipTLSVerify bool
}

type Store struct {
	credHubURL string
	httpClient *http.Client
	uaaClient  uaaClient
	token      token
}

// New creates a new CredHub API client
func New(cfg Config) (*Store, error) {
	httpClient, err := newHTTPClient(cfg.InsecureSkipTLSVerify, cfg.CACert)
	if err != nil {
		return nil, err
	}

	return &Store{
		credHubURL: cfg.CredHubURL,
		httpClient: httpClient,
		uaaClient: uaaClient{
			httpClient: httpClient,
			url:        cfg.UAAURL,
			name:       cfg.UAAClientName,
			secret:     cfg.UAAClientSecret,
		},
	}, nil
}

// Save will add a credential to CredHub and give the specified app permissions to see the credential
func (s *Store) Save(path string, cred any, actor string) error {
	if err := s.ensureToken(); err != nil {
		return err
	}

	// Set the credential value
	setRequestBody := map[string]any{
		"name":  path,
		"type":  "json",
		"value": cred,
	}

	if err := s.credHubClient().do(http.MethodPut, "/api/v1/data", setRequestBody, nil, http.StatusOK); err != nil {
		return fmt.Errorf("failed to store credential %q: %w", path, err)
	}

	// Update permissions for the credential
	permsRequestBody := map[string]any{
		"path":       path,
		"actor":      actor,
		"operations": []string{"read"},
	}

	if err := s.credHubClient().do(http.MethodPost, "/api/v2/permissions", permsRequestBody, nil, http.StatusCreated); err != nil {
		return fmt.Errorf("failed to set permission on credential %q: %w", path, err)
	}

	return nil
}

// Delete will remove a credential and all its permissions from CredHub
// It is idempotent so does not fail if the credential does not exist
func (s *Store) Delete(path string) error {
	if err := s.ensureToken(); err != nil {
		return err
	}

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

	if err := s.credHubClient().do(http.MethodGet, fmt.Sprintf("/api/v1/permissions?credential_name=%s", path), nil, &listPermissionsResponseBody, http.StatusOK); err != nil {
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

		if err := s.credHubClient().do(http.MethodGet, fmt.Sprintf("/api/v2/permissions?%s", query), nil, &getPermissionResponseBody, http.StatusOK); err != nil {
			return fmt.Errorf("failed to get permission %q for credential %q: %w", p.Actor, path, err)
		}

		if err := s.credHubClient().do(http.MethodDelete, fmt.Sprintf("/api/v2/permissions/%s", getPermissionResponseBody.UUID), nil, nil, http.StatusNoContent); err != nil {
			return fmt.Errorf("failed to delete permission ID %q: %w", getPermissionResponseBody.UUID, err)
		}
	}

	// Delete the credential
	if err := s.credHubClient().do(http.MethodDelete, fmt.Sprintf("/api/v1/data?name=%s", path), nil, nil, http.StatusNoContent); err != nil {
		return fmt.Errorf("failed to delete credential %q: %w", path, err)
	}
	return nil
}

// ensureToken fetches a token from UAA if required
func (s *Store) ensureToken() error {
	if s.token.valid() {
		return nil
	}

	tok, err := s.uaaClient.oauthToken()
	if err != nil {
		return err
	}

	s.token = tok
	return nil
}

func (s *Store) credHubClient() credHubClient {
	return credHubClient{
		httpClient: s.httpClient,
		url:        s.credHubURL,
		token:      s.token.value,
	}
}

func newHTTPClient(insecureSkipVerify bool, caCert string) (*http.Client, error) {
	tlsConfig := tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	if caCert != "" {
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
