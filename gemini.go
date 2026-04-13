package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"news/fetcher"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// GeminiService is a service for interacting with the Gemini API.
type GeminiService struct {
	genaiClient *genai.Client
	prompt      string
	model       string
	timeout     time.Duration
}

// NewGeminiService creates a new GeminiService.
func NewGeminiService(apiKey string, prompt string, model string, timeout time.Duration) *GeminiService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("failed to create genai client: %v", err)
	}
	if timeout <= 0 {
		timeout = DefaultAPITimeout
	}
	if model == "" {
		model = GeminiModel
	}

	return &GeminiService{
		genaiClient: client,
		prompt:      prompt,
		model:       model,
		timeout:     timeout,
	}
}

// AnalyzeNews analyzes news articles using the Gemini API.
func (s *GeminiService) AnalyzeNews(items []fetcher.NewsItem, attempts int, delay time.Duration) (string, string, error) {
	itemCharLimit := DefaultGeminiItemCharLimit
	totalCharLimit := DefaultGeminiTotalCharLimit
	if attempts < 1 {
		attempts = 1
	}
	if delay <= 0 {
		delay = DefaultRetryDelay
	}

	var analysis string
	var err error
	for attempt := 1; ; attempt++ {
		fullPrompt := fmt.Sprintf(s.prompt, s.buildNewsContent(items, itemCharLimit, totalCharLimit))
		analysis, err = s.generateContent(fullPrompt)
		if err == nil {
			break
		}

		if !isRetriableGeminiError(err) {
			return "", "", err
		}

		maxAttempts := retryAttemptsForGeminiError(attempts, err)
		if attempt >= maxAttempts {
			return "", "", err
		}

		waitFor := recommendedRetryDelay(err, delay, attempt)
		if shouldShrinkGeminiPayload(err) {
			itemCharLimit, totalCharLimit = shrinkPromptLimits(itemCharLimit, totalCharLimit)
			LogWarn(
				"Gemini request failed, retrying with smaller payload",
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"wait_for", waitFor.String(),
				"item_char_limit", itemCharLimit,
				"total_char_limit", totalCharLimit,
				"error", err,
			)
		} else if isRateLimitError(err) {
			LogWarn(
				"Gemini rate limited, retrying after backoff",
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"wait_for", waitFor.String(),
				"error", err,
			)
		} else {
			LogWarn(
				"Gemini request failed, retrying",
				"attempt", attempt,
				"max_attempts", maxAttempts,
				"wait_for", waitFor.String(),
				"error", err,
			)
		}

		time.Sleep(waitFor)
	}

	// Extract image URL from the first line and the rest of the analysis
	parts := strings.SplitN(analysis, "\n", 2)
	if len(parts) > 0 && (strings.HasPrefix(parts[0], "http://") || strings.HasPrefix(parts[0], "https://")) {
		imageURL := parts[0]
		analysisText := ""
		if len(parts) > 1 {
			analysisText = parts[1]
		}
		return imageURL, analysisText, nil
	}

	return "", analysis, nil
}

func (s *GeminiService) generateContent(fullPrompt string) (string, error) {
	model := s.genaiClient.GenerativeModel(s.model)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if txt, ok := part.(genai.Text); ok {
				return string(txt), nil
			}
		}
	}

	return "", nil
}

func (s *GeminiService) buildNewsContent(items []fetcher.NewsItem, itemCharLimit, totalCharLimit int) string {
	var builder strings.Builder
	usedChars := 0

	for i, item := range items {
		if usedChars >= totalCharLimit {
			break
		}

		content := item.Content
		if content == "" {
			content = item.RawContent
		}
		content = normalizeWhitespace(content)
		content, _ = truncateRunes(content, itemCharLimit)

		entry := fmt.Sprintf("Item %d\nTitle: %s\nLink: %s\n", i+1, normalizeWhitespace(item.Title), item.Link)
		if item.ImageURL != "" {
			entry += fmt.Sprintf("Image: %s\n", item.ImageURL)
		}
		entry += fmt.Sprintf("Content: %s\n\n", content)

		remaining := totalCharLimit - usedChars
		if remaining <= 0 {
			break
		}

		entry, _ = truncateRunes(entry, remaining)
		builder.WriteString(entry)
		usedChars += len([]rune(entry))
	}

	return builder.String()
}

func normalizeWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func truncateRunes(text string, limit int) (string, bool) {
	if limit <= 0 {
		return "", len(text) > 0
	}

	runes := []rune(text)
	if len(runes) <= limit {
		return text, false
	}

	if limit <= 3 {
		return string(runes[:limit]), true
	}

	return string(runes[:limit-3]) + "...", true
}

func shrinkPromptLimits(itemCharLimit, totalCharLimit int) (int, int) {
	nextItemLimit := max(MinimumGeminiItemCharLimit, itemCharLimit*GeminiLimitShrinkPercent/100)
	nextTotalLimit := max(MinimumGeminiTotalCharLimit, totalCharLimit*GeminiLimitShrinkPercent/100)
	return nextItemLimit, nextTotalLimit
}

func isRetriableGeminiError(err error) bool {
	return isDeadlineError(err) || isRateLimitError(err) || isTransientGoogleAPIError(err)
}

func shouldShrinkGeminiPayload(err error) bool {
	return isDeadlineError(err)
}

func isDeadlineError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "context deadline exceeded")
}

func isRateLimitError(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusTooManyRequests {
		return true
	}

	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "error 429") ||
		strings.Contains(lowerErr, "quota exceeded") ||
		strings.Contains(lowerErr, "rate limit")
}

func isTransientGoogleAPIError(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == http.StatusInternalServerError ||
			apiErr.Code == http.StatusBadGateway ||
			apiErr.Code == http.StatusServiceUnavailable ||
			apiErr.Code == http.StatusGatewayTimeout
	}

	lowerErr := strings.ToLower(err.Error())
	return strings.Contains(lowerErr, "error 500") ||
		strings.Contains(lowerErr, "error 502") ||
		strings.Contains(lowerErr, "error 503") ||
		strings.Contains(lowerErr, "error 504")
}

func retryAttemptsForGeminiError(baseAttempts int, err error) int {
	if baseAttempts < 1 {
		baseAttempts = 1
	}

	if isRateLimitError(err) {
		return max(baseAttempts, DefaultGeminiRateLimitAttempts)
	}

	if isRetriableGeminiError(err) {
		return max(baseAttempts, DefaultGeminiRetryAttempts)
	}

	return baseAttempts
}

var retryAfterPattern = regexp.MustCompile(`(?i)retry in ([0-9]+(?:\.[0-9]+)?)s`)

func recommendedRetryDelay(err error, fallback time.Duration, attempt int) time.Duration {
	if fallback <= 0 {
		fallback = DefaultRetryDelay
	}

	if delay, ok := retryAfterHeaderDelay(err); ok {
		return clampRetryDelay(delay)
	}

	if matches := retryAfterPattern.FindStringSubmatch(err.Error()); len(matches) == 2 {
		if seconds, parseErr := time.ParseDuration(matches[1] + "s"); parseErr == nil {
			return clampRetryDelay(time.Duration(math.Ceil(seconds.Seconds()+1)) * time.Second)
		}
	}

	if attempt < 1 {
		attempt = 1
	}

	backoff := fallback * time.Duration(1<<(attempt-1))
	if isDeadlineError(err) {
		return clampRetryDelay(backoff)
	}

	if isRateLimitError(err) {
		return clampRetryDelay(maxDuration(5*time.Second, backoff))
	}

	if isTransientGoogleAPIError(err) {
		return clampRetryDelay(backoff)
	}

	return clampRetryDelay(fallback)
}

func retryAfterHeaderDelay(err error) (time.Duration, bool) {
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) || apiErr.Header == nil {
		return 0, false
	}

	retryAfter := strings.TrimSpace(apiErr.Header.Get("Retry-After"))
	if retryAfter == "" {
		return 0, false
	}

	if seconds, convErr := time.ParseDuration(retryAfter + "s"); convErr == nil {
		return seconds, true
	}

	retryTime, parseErr := http.ParseTime(retryAfter)
	if parseErr != nil {
		return 0, false
	}

	return time.Until(retryTime), true
}

func clampRetryDelay(delay time.Duration) time.Duration {
	if delay < time.Second {
		delay = time.Second
	}
	if delay > MaxGeminiRetryDelay {
		return MaxGeminiRetryDelay
	}
	return delay
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// Close closes the Gemini client.
func (s *GeminiService) Close() {
	s.genaiClient.Close()
}
