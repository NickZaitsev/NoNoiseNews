package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"news/fetcher"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiService is a service for interacting with the Gemini API.
type GeminiService struct {
	genaiClient *genai.Client
}

// NewGeminiService creates a new GeminiService.
func NewGeminiService(apiKey string) *GeminiService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("failed to create genai client: %v", err)
	}
	return &GeminiService{genaiClient: client}
}

// AnalyzeNews analyzes news articles using the Gemini API.
func (s *GeminiService) AnalyzeNews(items []fetcher.NewsItem) (string, error) {
	var newsContent string
	for _, item := range items {
		newsContent += fmt.Sprintf("Title: %s\nContent: %s\n\n", item.Title, item.Content)
	}

	prompt := fmt.Sprintf(`Вы являетесь экспертом по глобальным новостям и редактором. Ваша цель — определить **самое глобально значимое** событие. Оцените следующие статьи по **долгосрочному глобальному значению** по шкале от 1 до 10:

* 10 = событие, которое, вероятно, будет помнить во всем мире в течение многих лет (например, начало крупной войны, убийство мирового лидера, исторический климатический рубеж, глобальный финансовый крах).
* 9 = глобально значимое событие с крупными экономическими, политическими или научными последствиями.
* 8 или ниже = событие, важное регионально или краткосрочно.

Выберите **не более одной** статьи с рейтингом 10/10. Если ни одна статья не заслуживает 10, **ничего не выводите** (верните пустую строку). Если есть сомнения в её уникальной мировой значимости, **не выбирайте ничего** (верните ""). 

Если статья подходит, то выведите краткое резюме в формате MarkdownV2 (ВАЖНО!): 

* Начните с *жирного заголовка*.
* Краткое содержание новости с самыми важными

Разделяйте смысловые блоки двойным переносом строки. Не пиши слишком длинные и сложные предложения. Длина новости должна быть не больше чем 6 предложений! 

Вывод должен быть только переписанным текстом, без объяснений, оценок или ссылок.

Входные новости: %s`, newsContent)

	model := s.genaiClient.GenerativeModel("gemini-2.5-pro")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
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

// Close closes the Gemini client.
func (s *GeminiService) Close() {
	s.genaiClient.Close()
}
