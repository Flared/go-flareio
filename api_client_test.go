package flareio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateClientWithApiKey(t *testing.T) {
	c := NewClient("test-key")
	assert.Equal(t, "test-key", c.apiKey)
	assert.Equal(t, 0, c.tenantId)
}

func TestCreateClientWithTenantId(t *testing.T) {
	c := NewClient(
		"test-key",
		WithTenantId(42),
	)
	assert.Equal(t, "test-key", c.apiKey)
	assert.Equal(t, 42, c.tenantId)
}
