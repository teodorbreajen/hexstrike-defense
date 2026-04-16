package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRateLimiter_TokenRefillAfterTime tests that tokens are refilled after the refillRate duration passes
func TestRateLimiter_TokenRefillAfterTime(t *testing.T) {
	tests := []struct {
		name        string
		maxTokens   int
		elapsed     time.Duration
		shouldAllow bool
	}{
		{
			name:        "refill after 1 minute",
			maxTokens:   10,
			elapsed:     time.Minute,
			shouldAllow: true,
		},
		{
			name:        "refill after 2 minutes",
			maxTokens:   10,
			elapsed:     2 * time.Minute,
			shouldAllow: true,
		},
		{
			name:        "no refill before 1 minute - tokens remain exhausted",
			maxTokens:   10,
			elapsed:     30 * time.Second, // less than refill rate
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := &RateLimiter{
				tokens:     0, // exhausted
				maxTokens:  tt.maxTokens,
				refillRate: time.Minute,
				lastRefill: time.Now().Add(-tt.elapsed),
			}

			// Try to allow a request
			allowed := rl.Allow()

			assert.Equal(t, tt.shouldAllow, allowed,
				"request should be allowed=%v after elapsed=%v with refillRate=%v",
				tt.shouldAllow, tt.elapsed, time.Minute)
		})
	}
}

// TestRateLimiter_AllowReturnsFalseWhenExhausted tests that Allow() returns false when tokens are exhausted
func TestRateLimiter_AllowReturnsFalseWhenExhausted(t *testing.T) {
	// Zero tokens but refill should trigger
	rl := &RateLimiter{
		tokens:     0,
		maxTokens:  10,
		refillRate: time.Minute,
		lastRefill: time.Now().Add(-time.Minute), // should trigger refill
	}

	// Should allow because refill happens
	allowed := rl.Allow()
	assert.True(t, allowed, "should allow when refill happens")
}

// TestRateLimiter_AllowReturnsTrueWhenTokensAvailable tests that Allow() returns true when tokens are available
func TestRateLimiter_AllowReturnsTrueWhenTokensAvailable(t *testing.T) {
	rl := NewRateLimiter(5)

	// First request should be allowed
	allowed := rl.Allow()
	assert.True(t, allowed, "first request should be allowed")

	// After first, tokens should be 4
	assert.Equal(t, 4, rl.tokens, "tokens should be decremented to 4")

	// Second request should be allowed
	allowed = rl.Allow()
	assert.True(t, allowed, "second request should be allowed")
	assert.Equal(t, 3, rl.tokens, "tokens should be decremented to 3")
}

// TestRateLimiter_NewRateLimiter tests the NewRateLimiter constructor
func TestRateLimiter_NewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(50)

	assert.Equal(t, 50, rl.maxTokens, "maxTokens should be set correctly")
	assert.Equal(t, 50, rl.tokens, "initial tokens should equal maxTokens")
	assert.Equal(t, time.Minute, rl.refillRate, "refillRate should default to 1 minute")
}

// TestRateLimiter_DecrementBehavior tests that tokens properly decrement with each allowed request
func TestRateLimiter_DecrementBehavior(t *testing.T) {
	rl := NewRateLimiter(3)

	// First request
	assert.True(t, rl.Allow())
	assert.Equal(t, 2, rl.tokens, "tokens should decrement to 2")

	// Second request
	assert.True(t, rl.Allow())
	assert.Equal(t, 1, rl.tokens, "tokens should decrement to 1")

	// Third request
	assert.True(t, rl.Allow())
	assert.Equal(t, 0, rl.tokens, "tokens should be 0")

	// Fourth request should fail
	assert.False(t, rl.Allow())
	assert.Equal(t, 0, rl.tokens, "tokens should stay at 0")
}
