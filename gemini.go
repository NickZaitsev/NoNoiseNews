package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"news/fetcher"

	"google.golang.org/api/option"
	"google.golang.org/genai"
)

// GeminiService is a service for interacting with the Gemini API.
type GeminiService struct {
	client *genai.GenerativeModel
}

// NewGeminiService creates a new GeminiService.
func NewGeminiService(apiKey string) *GeminiService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("failed to create genai client: %v", err)
	}
	// For text-only input, use the gemini-1.5-flash model
	model := client.GenerativeModel("gemini-1.5-flash")
	return &GeminiService{client: model}
}

// AnalyzeNews analyzes news articles using the Gemini API.
func (s *GeminiService) AnalyzeNews(items []fetcher.NewsItem) (string, error) {
	var newsContent string
	for _, item := range items {
		newsContent += fmt.Sprintf("Title: %s\nContent: %s\n\n", item.Title, item.Content)
	}

	prompt := fmt.Sprintf(`You are an expert global news analyst and editorial curator.

Your goal is to identify only the *most globally consequential* news event.

Evaluate the following articles for **long-term global significance** on a 1–10 scale:

- 10 = an event that will likely be remembered globally for years (e.g., major war outbreak, world leader assassination, historic climate milestone, global financial collapse).
- 9 = a globally influential development with major economic, political, or scientific consequences.
- 8 or below = regionally important or short-term impactful events.

Choose **at most one** article rated 10/10.  
If none truly deserve 10, output nothing.

Before finalizing, compare your chosen article against others — if there’s any doubt that it’s uniquely world-shaping, **do not select any**.

If one qualifies, output a concise 3–5 sentence summary in Markdown:
- Start with a **bold headline**.
- What happened.
- Why it matters globally.
- What might follow.

Separate meaning blocks with two line breaks.  
Be factual, neutral, and restrained — avoid hype or speculation.  
Output only the rewritten text. No explanations, no scores, no lists, no links.

Input news articles:
%s`, newsContent)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := s.client.GenerateContent(ctx, genai.Text(prompt))
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