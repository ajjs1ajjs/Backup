// Utils - Helper utilities for NovaBackup
package utils

import (
	"fmt"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries   int           // Maximum number of retry attempts
	InitialDelay time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay cap
	Multiplier   float64       // Delay multiplier for exponential backoff
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryWithBackoff executes a function with exponential backoff retry logic.
// Returns the result of the first successful execution or the last error after all retries.
func RetryWithBackoff(operation func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		if err := operation(); err == nil {
			return nil // Success
		} else {
			lastErr = err
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxRetries-1 {
			time.Sleep(delay)

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// RetryWithBackoffWithValue executes a function that returns a value with exponential backoff
func RetryWithBackoffWithValue[T any](operation func() (T, error), config *RetryConfig) (T, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var zero T
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		if result, err := operation(); err == nil {
			return result, nil // Success
		} else {
			lastErr = err
		}

		// Don't sleep after the last attempt
		if attempt < config.MaxRetries-1 {
			time.Sleep(delay)

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return zero, fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries, lastErr)
}

// IsRetryableError checks if an error is likely temporary and worth retrying
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network-related errors
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"i/o timeout",
		"temporary failure",
		"no such host",
		"network is unreachable",
		"connection timed out",
		"EOF",
		"lock timeout",
		"deadlock",
		"busy",
		"too many open files",
	}

	for _, retryable := range retryableErrors {
		if containsIgnoreCase(errStr, retryable) {
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsLower(s, substr)))
}

func containsLower(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
