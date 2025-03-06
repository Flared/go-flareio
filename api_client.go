package flareio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type ApiClient struct {
	tenantId    int
	apiKey      string
	httpClient  *retryablehttp.Client
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

func defaultHttpClient() *retryablehttp.Client {
	c := retryablehttp.NewClient()
	c.Logger = nil

	// Match the Python SDK retry settings:
	// - https://github.com/Flared/python-flareio/blob/d24061a086137e6a6fc7f467d6773660edf851f2/flareio/api_client.py#L44
	c.RetryMax = 5
	c.RetryWaitMin = time.Second * 2
	c.RetryWaitMax = time.Second * 15

	return c
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
		httpClient: defaultHttpClient(),
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

	// Prepare the request
	request, err := client.newRequest(
		"POST",
		"/tokens/generate",
		nil,
		bytes.NewReader(payloadBytes),
	)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %w", err)
	}
	request.Header.Set("Authorization", client.apiKey)

	// Tenant scoping like {"tenant_id": 123} is ignored if Content-Type not set to "application/json"
	request.Header.Set("Content-Type", "application/json")

	// Fire the request
	resp, err := client.do(request, false)
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

func (client *ApiClient) newRequest(
	method string,
	path string,
	params *url.Values,
	body io.Reader,
) (*http.Request, error) {
	destUrl, err := url.JoinPath(client.baseUrl, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create dest URL: %w", err)
	}
	if params != nil {
		destUrl = destUrl + "?" + params.Encode()
	}
	return http.NewRequest(method, destUrl, body)
}

func (client *ApiClient) do(
	request *http.Request,
	authenticated bool,
) (*http.Response, error) {
	if authenticated {
		apiToken, err := client.getOrGenerateToken()
		if err != nil {
			return nil, err
		}
		request.Header.Add(
			"Authorization",
			fmt.Sprintf("Bearer %s", apiToken),
		)
	}

	// Just like Go's User-Agent is hardcoded to "Go-http-client/1.1", we hardcode ours.
	// It isn't meant to reflect the actual library version.
	request.Header.Set("User-Agent", "go-flareio/0.1.0")

	retryableRequest, err := retryablehttp.FromRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare retryable request: %w", err)
	}
	return client.httpClient.Do(retryableRequest)
}

// Get peforms an authenticated GET request at the given path.
// Includes params in the query string.
func (client *ApiClient) Get(path string, params *url.Values) (*http.Response, error) {
	request, err := client.newRequest("GET", path, params, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	return client.do(request, true)
}

// Post performs an authenticated POST request at the given path.
// Includes params in the query string.
// The provided ContentType should describe the content of the body.
func (client *ApiClient) Post(
	path string,
	params *url.Values,
	contentType string,
	body io.Reader,
) (*http.Response, error) {
	request, err := client.newRequest("POST", path, params, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	request.Header.Set("Content-Type", contentType)
	return client.do(request, true)
}
