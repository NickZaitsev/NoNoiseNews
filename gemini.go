package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"news/fetcher"
	"news/utils"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiService is a service for interacting with the Gemini API.
type GeminiService struct {
	genaiClient *genai.Client
	prompt      string
}

// NewGeminiService creates a new GeminiService.
func NewGeminiService(apiKey string, prompt string) *GeminiService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("failed to create genai client: %v", err)
	}
	return &GeminiService{genaiClient: client, prompt: prompt}
}

// AnalyzeNews analyzes news articles using the Gemini API.
func (s *GeminiService) AnalyzeNews(items []fetcher.NewsItem, attempts int, delay time.Duration) (string, error) {
	var newsContent string
	for _, item := range items {
		imagePart := ""
		if item.ImageURL != "" {
			imagePart = fmt.Sprintf("Image: %s\n", item.ImageURL)
		}
		newsContent += fmt.Sprintf("Title: %s\n%sContent: %s\n\n", item.Title, imagePart, item.RawContent)
	}

	fullPrompt := fmt.Sprintf(s.prompt, newsContent)

	analysis, err := utils.Retry(attempts, delay, func() (string, error) {
		model := s.genaiClient.GenerativeModel("gemini-2.5-pro")
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
	})

	return analysis, err
}

// Close closes the Gemini client.
func (s *GeminiService) Close() {
	s.genaiClient.Close()
}
