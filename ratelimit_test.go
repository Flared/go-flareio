//go:build go1.23

package flareio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	limiter := newLimiter(1)
	assert.NotNil(t, limiter, "limiter should be created")
	assert.Equal(t, time.Duration(0), limiter.sleptFor, "initial sleep time should be 0")

	// First tick should be instant
	limiter.tick()
	assert.Equal(t, time.Duration(0), limiter.sleptFor, "first tick should be instant")
}
