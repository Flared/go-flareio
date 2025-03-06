package flareio

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type clientTest struct {
	apiClient  *ApiClient
	httpServer *httptest.Server
}

func newClientTest(
	handler http.HandlerFunc,
) *clientTest {
	httpServer := httptest.NewServer(
		handler,
	)

	apiClient := NewApiClient(
		"test-api-key",
		withBaseUrl(httpServer.URL),
	)
	apiClient.apiToken = "test-api-token"
	apiClient.apiTokenExp = time.Now().Add(time.Minute * 45)

	ct := &clientTest{
		httpServer: httpServer,
		apiClient:  apiClient,
	}

	ct.apiClient.httpClient.RetryWaitMax = 0

	return ct
}

func (ct *clientTest) Close() {
	defer ct.httpServer.Close()
}

func TestGenerateToken(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/tokens/generate", r.URL.Path)
			assert.Equal(t, "test-api-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.Write([]byte(`{"token":"test-api-token"}`))
		}),
	)
	defer ct.Close()

	ct.apiClient.apiToken = ""
	ct.apiClient.apiTokenExp = time.Time{}

	assert.Equal(t, "", ct.apiClient.apiToken, "The initial api token should be empty")
	assert.True(t, ct.apiClient.isApiTokenExpired(), "The initial api token exp should be before now")

	token, err := ct.apiClient.GenerateToken()
	if !assert.NoError(t, err, "Generating a token") {
		return
	}
	assert.Equal(t, "test-api-token", token)
	assert.Equal(t, "test-api-token", ct.apiClient.apiToken)
	assert.False(t, ct.apiClient.isApiTokenExpired(), "The api token should be unexpired")
}

func TestGetUnauthenticated(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/tokens/generate" {
				assert.Equal(t, "test-api-key", r.Header.Get("Authorization"))
				w.Write([]byte(`{"token":"test-api-token"}`))
			} else {
				assert.Equal(t, "/test-endpoint", r.URL.Path)
				assert.Equal(t, "Bearer test-api-token", r.Header.Get("Authorization"))
				w.Write([]byte(`"hello"`))
			}
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Get("/test-endpoint", nil)
	if !assert.NoError(t, err, "Failed to make get request") {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err, "Failed to read resp body") {
		return
	}
	assert.Equal(t, `"hello"`, string(body), "Didn't get expected response")
}

func TestPost(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/hey", r.URL.Path, "didn't get the expected path")
			assert.Equal(t, "application/something-custom", r.Header.Get("Content-Type"), "didn't get the expected path")
			w.Write([]byte(`"ho"`))
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Post("/hey", nil, "application/something-custom", strings.NewReader(`"hey"`))
	if !assert.NoError(t, err, "failed to make post request") {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err, "failed to read response body") {
		return
	}
	assert.Equal(t, `"ho"`, string(body), "Didn't get expected response")
}

func TestGetParams(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/some-path", r.URL.Path, "didn't get the expected path")
			assert.Equal(t, "some-param=some-value", r.URL.RawQuery, "didn't get the expected query")
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Get(
		"/some-path",
		&url.Values{
			"some-param": []string{"some-value"},
		},
	)
	if !assert.NoError(t, err, "failed to make get request") {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err, "failed to read response body") {
		return
	}
	assert.Equal(t, []byte{}, body)
}

func TestGetUserAgent(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "go-flareio/0.1.0", r.Header.Get("User-Agent"), "didn't get the expected User-Agent")
			w.Write([]byte("user-agent-test"))
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Get("/some-path", nil)
	if !assert.NoError(t, err, "failed to make get request") {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if !assert.NoError(t, err, "failed to read response body") {
		return
	}
	assert.Equal(t, []byte("user-agent-test"), body)
}

func TestGetRetry429(t *testing.T) {
	requestsReceived := 0
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestsReceived = requestsReceived + 1
			if requestsReceived < 2 {
				w.WriteHeader(429)
			}
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Get(
		"/some-path",
		nil,
	)
	if !assert.NoError(t, err, "failed to make get request") {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, 2, requestsReceived, "didn't perform the number of expected requests")
}
