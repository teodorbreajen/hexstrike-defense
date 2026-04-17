package main

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestClientRateLimiterRace tests for race conditions in the rate limiter
func TestClientRateLimiterRace(t *testing.T) {
	rl := NewClientRateLimiter(100)
	clientID := "test-client"

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each goroutine makes 10 requests
			for j := 0; j < 10; j++ {
				rl.Allow(clientID)
			}
		}()
	}

	wg.Wait()

	// Test passed without race detection
}

// TestRateLimiterConcurrentClients tests concurrent access with multiple clients
func TestRateLimiterConcurrentClients(t *testing.T) {
	rl := NewClientRateLimiter(100)

	var wg sync.WaitGroup
	numGoroutines := 50
	numClients := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clientID := strings.Repeat("client", 1) + string(rune('0'+id%10))
			for j := 0; j < 10; j++ {
				rl.Allow(clientID)
			}
		}(i % numClients)
	}

	wg.Wait()
}

// TestMetricsRace tests for race conditions in metrics
func TestMetricsRace(t *testing.T) {
	m := NewMetrics()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.RecordRequest(id%2 == 0, time.Millisecond*time.Duration(j), 200+(id%5)*100)
			}
		}(i)
	}

	wg.Wait()

	// Concurrent reads while writing
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				m.GetMetrics()
			}
		}()
	}

	wg.Wait()
}

// TestCircuitBreakerRace tests for race conditions in circuit breaker
func TestCircuitBreakerRace(t *testing.T) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond)

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if id%2 == 0 {
					cb.RecordSuccess()
				} else {
					cb.RecordFailure()
				}
				cb.Allow()
				cb.GetState()
			}
		}(i)
	}

	wg.Wait()
}

// TestProxyHandlerRace tests concurrent HTTP requests
func TestProxyHandlerRace(t *testing.T) {
	proxy := &Proxy{
		config: &ProxyConfig{
			MaxBodySize: 1024 * 1024,
			JWTSecret:   "test-secret-key-for-testing-only",
		},
		metrics:   NewMetrics(),
		logger:    NewLogger("test"),
		clientRL:  NewClientRateLimiter(1000),
		semaphore: make(chan struct{}, 100),
	}

	proxy.middlewareChain = []Middleware{
		proxy.rateLimitMiddleware,
	}

	handler := proxy.Handler()

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("POST", "/", strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)
		}(i)
	}

	wg.Wait()
}

// TestGetClientIPRace tests concurrent client IP extraction
func TestGetClientIPRace(t *testing.T) {
	trustedProxies := "10.0.0.1,192.168.1.0/24"

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "10.0.0.1:12345"
			req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.2")
			req.Header.Set("X-Real-IP", "203.0.113.1")

			_ = getClientIP(req, trustedProxies)
		}(i)
	}

	wg.Wait()
}

// TestIsTrustedProxyRace tests concurrent proxy validation
func TestIsTrustedProxyRace(t *testing.T) {
	trustedProxies := "10.0.0.1,192.168.1.0/24,172.16.0.0/12"

	var wg sync.WaitGroup
	numGoroutines := 100

	addresses := []string{
		"10.0.0.1:12345",
		"10.0.0.2:12345",
		"192.168.1.1:12345",
		"192.168.1.255:12345",
		"172.16.0.1:12345",
		"172.31.255.255:12345",
		"8.8.8.8:12345",
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			addr := addresses[id%len(addresses)]
			_ = isTrustedProxy(addr, trustedProxies)
		}(i)
	}

	wg.Wait()
}

// TestIsInternalURLRace tests concurrent SSRF check
func TestIsInternalURLRace(t *testing.T) {
	urls := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://10.0.0.1:8080",
		"http://172.15.0.1:8080",
		"http://172.16.0.1:8080",
		"http://192.168.1.1:8080",
		"http://169.254.169.254:8080",
		"https://example.com",
		"https://api.openai.com",
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			url := urls[id%len(urls)]
			_ = isInternalURL(url)
		}(i)
	}

	wg.Wait()
}

// TestConcurrentTokenBucketRefill tests token bucket refill under concurrent access
func TestConcurrentTokenBucketRefill(t *testing.T) {
	rl := NewClientRateLimiter(60) // 60 requests per minute
	clientID := "refill-test"

	// Exhaust tokens
	for i := 0; i < 60; i++ {
		rl.Allow(clientID)
	}

	// Now test concurrent refill
	var wg sync.WaitGroup
	numGoroutines := 10

	// Wait a bit for time to pass
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				rl.Allow(clientID)
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()
	// Test passed - no race condition detected
}

// TestConcurrentMetricsGetSet tests concurrent read/write of metrics
func TestConcurrentMetricsGetSet(t *testing.T) {
	m := NewMetrics()

	done := make(chan bool)

	// Writer goroutines
	for i := 0; i < 10; i++ {
		go func(id int) {
			for {
				select {
				case <-done:
					return
				default:
					m.RecordRequest(id%2 == 0, time.Duration(id)*time.Millisecond, 200+id)
				}
			}
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 5; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					m.GetMetrics()
				}
			}
		}()
	}

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)
	close(done)
}
