package retry

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryStrategy defines retry behavior
type RetryStrategy int

const (
	// ConstantBackoff retries with constant delay
	ConstantBackoff RetryStrategy = iota
	// ExponentialBackoff retries with exponentially increasing delay
	ExponentialBackoff
	// JitteredBackoff adds randomness to exponential backoff
	JitteredBackoff
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	Strategy      RetryStrategy
	RetryableFunc func(error) bool // Function to determine if error is retryable
}

// DefaultConfig returns a sensible default retry configuration
func DefaultConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Strategy:     ExponentialBackoff,
		RetryableFunc: func(err error) bool {
			// By default, retry all errors
			return true
		},
	}
}

// WithMaxAttempts sets the maximum number of retry attempts
func (c *RetryConfig) WithMaxAttempts(attempts int) *RetryConfig {
	c.MaxAttempts = attempts
	return c
}

// WithInitialDelay sets the initial delay between retries
func (c *RetryConfig) WithInitialDelay(delay time.Duration) *RetryConfig {
	c.InitialDelay = delay
	return c
}

// WithMaxDelay sets the maximum delay between retries
func (c *RetryConfig) WithMaxDelay(delay time.Duration) *RetryConfig {
	c.MaxDelay = delay
	return c
}

// WithStrategy sets the retry strategy
func (c *RetryConfig) WithStrategy(strategy RetryStrategy) *RetryConfig {
	c.Strategy = strategy
	return c
}

// WithRetryableFunc sets a custom function to determine if an error is retryable
func (c *RetryConfig) WithRetryableFunc(fn func(error) bool) *RetryConfig {
	c.RetryableFunc = fn
	return c
}

// Retry executes the given function with retry logic
func Retry(fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultConfig()
	}

	var lastErr error
	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()

		if err == nil {
			// Success!
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if config.RetryableFunc != nil && !config.RetryableFunc(err) {
			// Non-retryable error, fail immediately
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Don't delay after last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate delay
		delay := config.calculateDelay(attempt)

		// Wait before retry
		time.Sleep(delay)
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for the given attempt
func (c *RetryConfig) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch c.Strategy {
	case ConstantBackoff:
		delay = c.InitialDelay

	case ExponentialBackoff:
		// delay = initialDelay * 2^(attempt-1)
		multiplier := math.Pow(2, float64(attempt-1))
		delay = time.Duration(float64(c.InitialDelay) * multiplier)

	case JitteredBackoff:
		// Exponential with random jitter
		multiplier := math.Pow(2, float64(attempt-1))
		baseDelay := float64(c.InitialDelay) * multiplier

		// Add ±25% jitter
		jitter := baseDelay * 0.25 * (rand.Float64()*2 - 1)
		delay = time.Duration(baseDelay + jitter)
	}

	// Cap at max delay
	if delay > c.MaxDelay {
		delay = c.MaxDelay
	}

	return delay
}

// RetryResult contains information about the retry execution
type RetryResult struct {
	Success      bool
	Attempts     int
	TotalTime    time.Duration
	LastError    error
	AttemptTimes []time.Duration
}

// RetryWithResult executes the function and returns detailed retry information
func RetryWithResult(fn func() error, config *RetryConfig) *RetryResult {
	if config == nil {
		config = DefaultConfig()
	}

	result := &RetryResult{
		AttemptTimes: make([]time.Duration, 0),
	}

	startTime := time.Now()

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		attemptStart := time.Now()
		err := fn()
		attemptDuration := time.Since(attemptStart)

		result.Attempts = attempt
		result.AttemptTimes = append(result.AttemptTimes, attemptDuration)

		if err == nil {
			result.Success = true
			result.TotalTime = time.Since(startTime)
			return result
		}

		result.LastError = err

		if config.RetryableFunc != nil && !config.RetryableFunc(err) {
			break
		}

		if attempt < config.MaxAttempts {
			delay := config.calculateDelay(attempt)
			time.Sleep(delay)
		}
	}

	result.Success = false
	result.TotalTime = time.Since(startTime)
	return result
}

// Common retry configurations

// NetworkRetryConfig returns a config suitable for network operations
func NetworkRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Strategy:     JitteredBackoff,
		RetryableFunc: func(err error) bool {
			// Retry on network errors
			errMsg := err.Error()
			return contains(errMsg, "timeout") ||
				contains(errMsg, "connection refused") ||
				contains(errMsg, "connection reset")
		},
	}
}

// DatabaseRetryConfig returns a config suitable for database operations
func DatabaseRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Strategy:     ExponentialBackoff,
		RetryableFunc: func(err error) bool {
			errMsg := err.Error()
			return contains(errMsg, "deadlock") ||
				contains(errMsg, "connection") ||
				contains(errMsg, "timeout")
		},
	}
}

// QuickRetryConfig returns a config for quick retries (e.g., in-memory operations)
func QuickRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Strategy:     ConstantBackoff,
	}
}

// Helper function
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) && (str[:len(substr)] == substr || str[len(str)-len(substr):] == substr))
}

// Example usage in tests:
//
// import "github.com/cloudfoundry/cli/testhelpers/retry"
//
// err := retry.Retry(func() error {
//     return makeNetworkCall()
// }, retry.NetworkRetryConfig())
//
// Or with custom config:
//
// config := retry.DefaultConfig().
//     WithMaxAttempts(5).
//     WithStrategy(retry.JitteredBackoff)
//
// err := retry.Retry(func() error {
//     return flakeyOperation()
// }, config)
