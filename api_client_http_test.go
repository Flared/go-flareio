package flareio

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ClientTest struct {
	apiClient  *ApiClient
	httpServer *httptest.Server
}

func newClientTest(
	handler http.HandlerFunc,
) *ClientTest {
	httpServer := httptest.NewServer(
		handler,
	)
	apiClient := NewApiClient(
		"test-api-key",
		withBaseUrl(httpServer.URL),
	)
	clientTest := &ClientTest{
		httpServer: httpServer,
		apiClient:  apiClient,
	}
	return clientTest
}

func (clientTest *ClientTest) Close() {
	defer clientTest.httpServer.Close()
}

func TestGenerateToken(t *testing.T) {
	clientTest := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/tokens/generate", r.URL.Path)
			assert.Equal(t, []string{"test-api-key"}, r.Header["Authorization"])
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"token":"test-api-token"}`))
		}),
	)
	defer clientTest.Close()

	assert.Equal(t, "", clientTest.apiClient.apiToken, "The initial api token should be empty")
	assert.True(t, clientTest.apiClient.isApiTokenExpired(), "The initial api token exp should be before now")

	token, err := clientTest.apiClient.GenerateToken()
	assert.NoError(t, err, "Generating a token")
	assert.Equal(t, "test-api-token", token)
	assert.Equal(t, "test-api-token", clientTest.apiClient.apiToken)
	assert.False(t, clientTest.apiClient.isApiTokenExpired(), "The api token should be unexpired")
}

func TestGetUnauthenticated(t *testing.T) {

	clientTest := newClientTest(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/tokens/generate" {
				assert.Equal(t, []string{"test-api-key"}, r.Header["Authorization"])
				w.Write([]byte(`{"token":"test-api-token"}`))
			} else {
				assert.Equal(t, "/test-endpoint", r.URL.Path)
				assert.Equal(t, []string{"Bearer test-api-token"}, r.Header["Authorization"])
				w.Write([]byte(`"hello"`))
			}
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer clientTest.Close()

	resp, err := clientTest.apiClient.Get("/test-endpoint")
	assert.NoError(t, err, "Failed to make get request")

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Failed to read resp body")

	assert.Equal(t, `"hello"`, string(body), "Didn't get expected response")
	assert.NoError(t, err, "Generating a token")
}
