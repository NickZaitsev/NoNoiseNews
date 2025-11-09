package main

import (
	"context"
	"fmt"
	"log"
	"strings"
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
func (s *GeminiService) AnalyzeNews(items []fetcher.NewsItem, attempts int, delay time.Duration) (string, string, error) {
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
		model := s.genaiClient.GenerativeModel(GeminiModel)
		ctx, cancel := context.WithTimeout(context.Background(), APITimeout)
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

	if err != nil {
		return "", "", err
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

// Close closes the Gemini client.
func (s *GeminiService) Close() {
	s.genaiClient.Close()
}
