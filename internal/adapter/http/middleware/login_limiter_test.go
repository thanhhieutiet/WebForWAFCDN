package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoginLimiter_Lockout(t *testing.T) {
	limiter := NewLoginLimiter(3, 50*time.Millisecond, 100*time.Millisecond)

	ip := "198.51.100.5"

	// Initially allowed
	allowed, _ := limiter.CheckAllowed(ip)
	assert.True(t, allowed)

	// Attempt 1: fail
	assert.False(t, limiter.RecordFailure(ip))
	allowed, _ = limiter.CheckAllowed(ip)
	assert.True(t, allowed)

	// Attempt 2: fail
	assert.False(t, limiter.RecordFailure(ip))
	allowed, _ = limiter.CheckAllowed(ip)
	assert.True(t, allowed)

	// Attempt 3: fail -> now locked out!
	assert.True(t, limiter.RecordFailure(ip))
	allowed, rem := limiter.CheckAllowed(ip)
	assert.False(t, allowed)
	assert.Greater(t, rem, time.Duration(0))

	// Wait for lockout duration to elapse
	time.Sleep(60 * time.Millisecond)
	allowed, _ = limiter.CheckAllowed(ip)
	assert.True(t, allowed)

	// Record success resets count
	limiter.RecordFailure(ip)
	limiter.RecordSuccess(ip)
	allowed, _ = limiter.CheckAllowed(ip)
	assert.True(t, allowed)
}
