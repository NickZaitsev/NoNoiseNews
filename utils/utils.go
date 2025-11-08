package utils

import (
	"log/slog"
	"time"
)

// Retry executes a function with a fixed number of retries and a delay between them.
func Retry[T any](attempts int, sleep time.Duration, fn func() (T, error)) (T, error) {
	var result T
	var err error

	for i := 0; i < attempts; i++ {
		result, err = fn()
		if err == nil {
			return result, nil // Success
		}

		slog.Warn("Retrying after error", "attempt", i+1, "error", err)
		time.Sleep(sleep)
	}

	return result, err // Return the last error
}