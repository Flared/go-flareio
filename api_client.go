package flareio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ApiClient struct {
	tenantId    int
	apiKey      string
	httpClient  *http.Client
	baseUrl     string
	apiToken    string
	apiTokenExp time.Time
}

type ApiClientOption func(*ApiClient)

// WithTenantId allows configuring the tenant id.
func WithTenantId(tenantId int) ApiClientOption {
	return func(client *ApiClient) {
		client.tenantId = tenantId
	}
}

// withBaseUrl allows configuring the base url, for testing only.
func withBaseUrl(baseUrl string) ApiClientOption {
	return func(client *ApiClient) {
		client.baseUrl = baseUrl
	}
}

// NewApiClient can be used to create a new ApiClient
// instance.
func NewApiClient(
	apiKey string,
	optionFns ...ApiClientOption,
) *ApiClient {
	c := &ApiClient{
		apiKey:     apiKey,
		baseUrl:    "https://api.flare.io/",
		httpClient: &http.Client{},
	}
	for _, optionFn := range optionFns {
		optionFn(c)
	}
	return c
}

// GenerateToken creates a Flare API token using the
// API Client's API key.
func (client *ApiClient) GenerateToken() (string, error) {
	// Prepare payload
	type GeneratePayload struct {
		TenantId int `json:"tenant_id,omitempty"`
	}
	payload := &GeneratePayload{
		TenantId: client.tenantId,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generate payload: %w", err)
	}

	// Create dest URL
	destUrl, err := url.JoinPath(client.baseUrl, "/tokens/generate")
	if err != nil {
		return "", fmt.Errorf("failed to create dest URL: %w", err)
	}

	// Prepare the request
	request, err := http.NewRequest(
		"POST",
		destUrl,
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}
	request.Header.Set("Authorization", client.apiKey)

	// Fire the request
	resp, err := client.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to generate API token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected response code: %d", resp.StatusCode)
	}

	// Parse response
	type TokenResponse struct {
		Token string `json:"token"`
	}
	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	client.apiToken = tokenResponse.Token
	client.apiTokenExp = time.Now().Add(time.Minute * 45)

	return tokenResponse.Token, nil
}

func (client *ApiClient) isApiTokenExpired() bool {
	return client.apiTokenExp.Before(time.Now())
}

func (client *ApiClient) getOrGenerateToken() (string, error) {
	if !client.isApiTokenExpired() {
		return client.apiToken, nil
	}
	return client.GenerateToken()
}

func (client *ApiClient) createDestUrl(
	path string,
	params *url.Values,
) (string, error) {
	destUrl, err := url.JoinPath(client.baseUrl, path)
	if err != nil {
		return "", fmt.Errorf("failed to create dest URL: %w", err)
	}
	if params != nil {
		destUrl = destUrl + "?" + params.Encode()
	}
	return destUrl, nil
}

func (client *ApiClient) do(request *http.Request) (*http.Response, error) {
	if apiToken, err := client.getOrGenerateToken(); err != nil {
		return nil, err
	} else {
		request.Header.Add(
			"Authorization",
			fmt.Sprintf("Bearer %s", apiToken),
		)
	}
	return client.httpClient.Do(request)
}

// Get peforms an authenticated Get request at the given path.
// Includes params in the query string.
func (client *ApiClient) Get(path string, params *url.Values) (*http.Response, error) {
	destUrl, err := client.createDestUrl(path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest URL: %w", err)
	}

	request, err := http.NewRequest("GET", destUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	return client.do(request)
}

// Post performs an authenticated Post request at the given path.
// Includes params in the query string.
// The provided ContentType should describe the content of the body.
func (client *ApiClient) Post(
	path string,
	params *url.Values,
	contentType string,
	body io.Reader,
) (*http.Response, error) {
	destUrl, err := client.createDestUrl(path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest URL: %w", err)
	}

	request, err := http.NewRequest("POST", destUrl, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	request.Header.Set("Content-Type", contentType)

	return client.do(request)
}
