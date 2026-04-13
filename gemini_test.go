package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"google.golang.org/api/googleapi"
)

func TestIsRateLimitErrorDetectsGoogleAPI429(t *testing.T) {
	err := fmt.Errorf("failed to generate content: %w", &googleapi.Error{Code: http.StatusTooManyRequests})
	if !isRateLimitError(err) {
		t.Fatalf("expected 429 googleapi error to be classified as rate limit")
	}
}

func TestRetryAttemptsForGeminiErrorExtendsRateLimitBudget(t *testing.T) {
	err := fmt.Errorf("failed to generate content: %w", &googleapi.Error{Code: http.StatusTooManyRequests})
	attempts := retryAttemptsForGeminiError(1, err)
	if attempts != DefaultGeminiRateLimitAttempts {
		t.Fatalf("expected rate limit attempts to be %d, got %d", DefaultGeminiRateLimitAttempts, attempts)
	}
}

func TestRecommendedRetryDelayUsesRetryHintFromMessage(t *testing.T) {
	err := fmt.Errorf("failed to generate content: googleapi: Error 429: quota exceeded. Please retry in 44.969935565s")
	delay := recommendedRetryDelay(err, 2*time.Second, 1)
	if delay != 46*time.Second {
		t.Fatalf("expected 46s retry delay, got %s", delay)
	}
}

func TestRecommendedRetryDelayUsesRetryAfterHeader(t *testing.T) {
	err := fmt.Errorf("failed to generate content: %w", &googleapi.Error{
		Code:   http.StatusTooManyRequests,
		Header: http.Header{"Retry-After": []string{"7"}},
	})
	delay := recommendedRetryDelay(err, 2*time.Second, 1)
	if delay != 7*time.Second {
		t.Fatalf("expected Retry-After header to win, got %s", delay)
	}
}
