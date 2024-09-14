package flareio

import (
	"io"
	"net/http"
	"net/http/httptest"
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
	return ct
}

func (ct *clientTest) Close() {
	defer ct.httpServer.Close()
}

func TestGenerateToken(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/tokens/generate", r.URL.Path)
			assert.Equal(t, []string{"test-api-key"}, r.Header["Authorization"])
			w.Write([]byte(`{"token":"test-api-token"}`))
		}),
	)
	defer ct.Close()

	ct.apiClient.apiToken = ""
	ct.apiClient.apiTokenExp = time.Time{}

	assert.Equal(t, "", ct.apiClient.apiToken, "The initial api token should be empty")
	assert.True(t, ct.apiClient.isApiTokenExpired(), "The initial api token exp should be before now")

	token, err := ct.apiClient.GenerateToken()
	assert.NoError(t, err, "Generating a token")
	assert.Equal(t, "test-api-token", token)
	assert.Equal(t, "test-api-token", ct.apiClient.apiToken)
	assert.False(t, ct.apiClient.isApiTokenExpired(), "The api token should be unexpired")
}

func TestGetUnauthenticated(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/tokens/generate" {
				assert.Equal(t, []string{"test-api-key"}, r.Header["Authorization"])
				w.Write([]byte(`{"token":"test-api-token"}`))
			} else {
				assert.Equal(t, "/test-endpoint", r.URL.Path)
				assert.Equal(t, []string{"Bearer test-api-token"}, r.Header["Authorization"])
				w.Write([]byte(`"hello"`))
			}
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Get("/test-endpoint")
	assert.NoError(t, err, "Failed to make get request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Failed to read resp body")
	assert.Equal(t, `"hello"`, string(body), "Didn't get expected response")
}

func TestPost(t *testing.T) {
	ct := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/hey", r.URL.Path, "didn't get the expected path")
			assert.Equal(t, []string{"application/something-custom"}, r.Header["Content-Type"], "didn't get the expected path")
			w.Write([]byte(`"ho"`))
		}),
	)
	defer ct.Close()

	resp, err := ct.apiClient.Post("/hey", "application/something-custom", strings.NewReader(`"hey"`))
	assert.NoError(t, err, "failed to make post request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "failed to read response body")
	assert.Equal(t, `"ho"`, string(body), "Didn't get expected response")
}
