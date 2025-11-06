package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"news/fetcher"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration.
type Config struct {
	GeminiAPIKey string
}

// LoadConfig loads the configuration from a .env file.
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set in .env file")
	}

	return &Config{
		GeminiAPIKey: apiKey,
	}
}

// GeminiRequest represents the request structure for Gemini API
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Content represents a content part in the request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a text part in the content
type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response structure from Gemini API
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content Content `json:"content"`
}

// GeminiService is a service for interacting with the Gemini API.
type GeminiService struct {
	apiKey string
}

// NewGeminiService creates a new GeminiService.
func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{apiKey: apiKey}
}

// AnalyzeNews analyzes news articles using the Gemini API.
func (s *GeminiService) AnalyzeNews(items []fetcher.NewsItem) (string, error) {
	var newsContent string
	for _, item := range items {
		newsContent += fmt.Sprintf("Title: %s\nContent: %s\n\n", item.Title, item.Content)
	}

	prompt := fmt.Sprintf(`For each of these news articles, assign an importance score from 1 to 10, where 10 = extremely significant global event (10 in year max) and 1 = trivial update. Return just text of 1 news article what you think is 10/10 important. Rewrite the news item concisely and factually, removing opinions and adjectives.

You are an expert news analyst. Your task is to identify which news items have long-term significance (signal) and which are transient noise.
Criteria for importance:
- global or structural impact
- technological or geopolitical shift
- enduring relevance (not just event-of-the-day)

Return just text with markdown, no links or something

Output in 3-5 sentences: what happened, why it matters, what might follow.
Tone: factual, calm, timeless.

News Articles:
%s`, newsContent)

	// Create request
	request := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API call
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent?key="+s.apiKey,
		bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API call failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	err = json.Unmarshal(body, &geminiResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no content generated")
}

func main() {
	config := LoadConfig()
	geminiService := NewGeminiService(config.GeminiAPIKey)

	sinceDate := time.Now().AddDate(0, 0, -1)

	allNews := []fetcher.NewsItem{}

	fmt.Println("--- Fetching from Meduza (Generic) ---")
	meduzaFetcher := &fetcher.GenericFetcher{URL: "https://meduza.io/rss/all"}
	meduzaItems, err := meduzaFetcher.Fetch(sinceDate)
	if err != nil {
		log.Printf("Error fetching from Meduza: %v", err)
	} else {
		allNews = append(allNews, meduzaItems...)
	}

	fmt.Println("\n--- Fetching from SVTV (Custom) ---")
	svtvFetcher := &fetcher.SvtvFetcher{URL: "https://svtv.org/feed/rss/"}
	svtvItems, err := svtvFetcher.Fetch(sinceDate)
	if err != nil {
		log.Printf("Error fetching from SVTV: %v", err)
	} else {
		allNews = append(allNews, svtvItems...)
	}

	if len(allNews) > 0 {
		fmt.Println("\n--- Analyzing News with Gemini ---")
		analysis, err := geminiService.AnalyzeNews(allNews)
		if err != nil {
			log.Fatalf("Failed to analyze news: %v", err)
		}
		fmt.Println(analysis)
	} else {
		fmt.Println("No news items to analyze.")
	}
}
