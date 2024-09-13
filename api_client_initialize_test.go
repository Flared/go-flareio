package flareio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateClientWithApiKey(t *testing.T) {
	c := NewClient("test-api-key")
	assert.Equal(t, "test-api-key", c.apiKey)
	assert.Equal(t, 0, c.tenantId)
	assert.Equal(t, "https://api.flare.io/", c.baseUrl)
}

func TestCreateClientWithTenantId(t *testing.T) {
	c := NewClient(
		"test-api-key",
		WithTenantId(42),
	)
	assert.Equal(t, "test-api-key", c.apiKey)
	assert.Equal(t, 42, c.tenantId)
}

func TestCreateClientWithBaseUrl(t *testing.T) {
	c := NewClient(
		"test-api-key",
		withBaseUrl("https://test.com/"),
	)
	assert.Equal(t, "https://test.com/", c.baseUrl)
}
