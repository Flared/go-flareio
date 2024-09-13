package flareio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ApiClient struct {
	tenantId   int
	apiKey     string
	httpClient *http.Client
	baseUrl    string
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
		return "", fmt.Errorf("failed to create test URL: %w", err)
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

	return tokenResponse.Token, nil
}
