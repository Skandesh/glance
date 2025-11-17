package glance

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/client"
)

// StripeClientPool manages a pool of Stripe API clients with circuit breaker and rate limiting
type StripeClientPool struct {
	clients      sync.Map // map[string]*StripeClientWrapper
	maxRetries   int
	retryBackoff time.Duration
}

// StripeClientWrapper wraps a Stripe client with circuit breaker and metrics
type StripeClientWrapper struct {
	client        *client.API
	apiKey        string
	mode          string
	circuitBreaker *CircuitBreaker
	rateLimiter   *RateLimiter
	lastUsed      time.Time
	mu            sync.RWMutex
}

// CircuitBreaker implements the circuit breaker pattern for external API calls
type CircuitBreaker struct {
	maxFailures  uint32
	resetTimeout time.Duration
	failures     uint32
	lastFailTime time.Time
	state        CircuitState
	mu           sync.RWMutex
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

var (
	globalStripePool     *StripeClientPool
	globalStripePoolOnce sync.Once
)

// GetStripeClientPool returns the global Stripe client pool (singleton)
func GetStripeClientPool() *StripeClientPool {
	globalStripePoolOnce.Do(func() {
		globalStripePool = &StripeClientPool{
			maxRetries:   3,
			retryBackoff: 1 * time.Second,
		}
	})
	return globalStripePool
}

// GetClient returns a Stripe client for the given API key with circuit breaker and rate limiting
func (p *StripeClientPool) GetClient(apiKey, mode string) (*StripeClientWrapper, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("stripe API key is required")
	}

	cacheKey := fmt.Sprintf("%s:%s", mode, apiKey[:12]) // Use prefix for cache key

	if cached, ok := p.clients.Load(cacheKey); ok {
		wrapper := cached.(*StripeClientWrapper)
		wrapper.mu.Lock()
		wrapper.lastUsed = time.Now()
		wrapper.mu.Unlock()
		return wrapper, nil
	}

	// Create new client with circuit breaker and rate limiter
	sc := &client.API{}
	sc.Init(apiKey, nil)

	wrapper := &StripeClientWrapper{
		client:   sc,
		apiKey:   apiKey,
		mode:     mode,
		lastUsed: time.Now(),
		circuitBreaker: &CircuitBreaker{
			maxFailures:  5,
			resetTimeout: 60 * time.Second,
			state:        CircuitClosed,
		},
		rateLimiter: &RateLimiter{
			tokens:     100.0,
			maxTokens:  100.0,
			refillRate: 10.0, // 10 requests per second
			lastRefill: time.Now(),
		},
	}

	p.clients.Store(cacheKey, wrapper)
	return wrapper, nil
}

// ExecuteWithRetry executes a function with retry logic, circuit breaker, and rate limiting
func (w *StripeClientWrapper) ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	// Check circuit breaker
	if !w.circuitBreaker.CanExecute() {
		return fmt.Errorf("circuit breaker open for Stripe API: too many failures")
	}

	// Wait for rate limiter
	if err := w.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit exceeded: %w", err)
	}

	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			slog.Info("Retrying Stripe API call",
				"operation", operation,
				"attempt", attempt,
				"backoff", backoff)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := fn()
		if err == nil {
			w.circuitBreaker.RecordSuccess()
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableStripeError(err) {
			w.circuitBreaker.RecordFailure()
			return fmt.Errorf("non-retryable Stripe error in %s: %w", operation, err)
		}

		w.circuitBreaker.RecordFailure()
		slog.Warn("Stripe API call failed",
			"operation", operation,
			"attempt", attempt,
			"error", err)
	}

	return fmt.Errorf("stripe operation %s failed after %d retries: %w", operation, maxRetries, lastErr)
}

// isRetryableStripeError determines if a Stripe error is retryable
func isRetryableStripeError(err error) bool {
	if err == nil {
		return false
	}

	stripeErr, ok := err.(*stripe.Error)
	if !ok {
		// Network errors are retryable
		return true
	}

	// Retry on rate limit, temporary issues, and server errors
	// Check HTTP status code for retryable errors
	if stripeErr.HTTPStatusCode >= 500 {
		return true // Server errors are retryable
	}

	if stripeErr.HTTPStatusCode == 429 {
		return true // Rate limiting is retryable
	}

	// Check error type
	switch stripeErr.Type {
	case "api_error":
		return true
	case "invalid_request_error":
		return false // Don't retry on invalid requests
	case "authentication_error":
		return false // Don't retry on auth errors
	case "card_error":
		return false // Don't retry on card errors
	case "rate_limit_error":
		return true
	default:
		return true
	}
}

// CircuitBreaker methods

func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitHalfOpen
			cb.failures = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitHalfOpen {
		cb.state = CircuitClosed
		cb.failures = 0
		slog.Info("Circuit breaker closed: service recovered")
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		if cb.state != CircuitOpen {
			cb.state = CircuitOpen
			slog.Error("Circuit breaker opened: too many failures",
				"failures", cb.failures,
				"resetTimeout", cb.resetTimeout)
		}
	}
}

// RateLimiter methods

func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = minFloat(rl.maxTokens, rl.tokens+(elapsed*rl.refillRate))
	rl.lastRefill = now

	// If we have tokens, consume one and proceed
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return nil
	}

	// Calculate wait time for next token
	waitTime := time.Duration((1.0-rl.tokens)/rl.refillRate) * time.Second

	// Unlock while waiting
	rl.mu.Unlock()
	select {
	case <-ctx.Done():
		rl.mu.Lock()
		return ctx.Err()
	case <-time.After(waitTime):
		rl.mu.Lock()
		rl.tokens = 0 // Consumed the token we waited for
		return nil
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// CleanupIdleClients removes clients that haven't been used in the specified duration
func (p *StripeClientPool) CleanupIdleClients(maxIdleTime time.Duration) {
	p.clients.Range(func(key, value interface{}) bool {
		wrapper := value.(*StripeClientWrapper)
		wrapper.mu.RLock()
		idle := time.Since(wrapper.lastUsed)
		wrapper.mu.RUnlock()

		if idle > maxIdleTime {
			p.clients.Delete(key)
			slog.Info("Removed idle Stripe client", "key", key, "idleTime", idle)
		}
		return true
	})
}

// GetMetrics returns metrics for monitoring
func (p *StripeClientPool) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"total_clients": 0,
		"circuit_states": map[string]int{
			"closed":    0,
			"open":      0,
			"half_open": 0,
		},
	}

	totalClients := 0
	circuitStates := map[string]int{"closed": 0, "open": 0, "half_open": 0}

	p.clients.Range(func(key, value interface{}) bool {
		totalClients++
		wrapper := value.(*StripeClientWrapper)
		wrapper.circuitBreaker.mu.RLock()
		state := wrapper.circuitBreaker.state
		wrapper.circuitBreaker.mu.RUnlock()

		switch state {
		case CircuitClosed:
			circuitStates["closed"]++
		case CircuitOpen:
			circuitStates["open"]++
		case CircuitHalfOpen:
			circuitStates["half_open"]++
		}
		return true
	})

	metrics["total_clients"] = totalClients
	metrics["circuit_states"] = circuitStates
	return metrics
}
