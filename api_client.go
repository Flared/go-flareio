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

func (client *ApiClient) do(request *http.Request) (*http.Response, error) {
	if client.isApiTokenExpired() {
		if _, err := client.GenerateToken(); err != nil {
			return nil, err
		}
	}
	apiToken := client.apiToken

	request.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", apiToken),
	)

	return client.httpClient.Do(request)
}

func (client *ApiClient) Get(path string) (*http.Response, error) {
	destUrl, err := url.JoinPath(client.baseUrl, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest URL: %w", err)
	}

	request, err := http.NewRequest("GET", destUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	return client.do(request)
}

func (client *ApiClient) Post(path, contentType string, body io.Reader) (*http.Response, error) {
	destUrl, err := url.JoinPath(client.baseUrl, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest URL: %w", err)
	}

	request, err := http.NewRequest("POST", destUrl, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	return client.do(request)
}
