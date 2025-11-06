package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"news/fetcher"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration.
type Config struct {
	GeminiAPIKey    string
	TelegramAPIKey  string
	TelegramChatID  string
	SvtvChannelID   string
	MeduzaChannelID string
}

// LoadConfig loads the configuration from a .env file.
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY not set in .env file")
	}

	telegramAPIKey := os.Getenv("TELEGRAM_API_KEY")
	if telegramAPIKey == "" {
		log.Fatal("TELEGRAM_API_KEY not set in .env file")
	}

	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
	if telegramChatID == "" {
		log.Fatal("TELEGRAM_CHAT_ID not set in .env file")
	}

	svtvChannelID := os.Getenv("SVTV_CHANNEL_ID")
	if svtvChannelID == "" {
		log.Fatal("SVTV_CHANNEL_ID not set in .env file")
	}

	meduzaChannelID := os.Getenv("MEDUZA_CHANNEL_ID")
	if meduzaChannelID == "" {
		log.Fatal("MEDUZA_CHANNEL_ID not set in .env file")
	}

	return &Config{
		GeminiAPIKey:    geminiAPIKey,
		TelegramAPIKey:  telegramAPIKey,
		TelegramChatID:  telegramChatID,
		SvtvChannelID:   svtvChannelID,
		MeduzaChannelID: meduzaChannelID,
	}
}

// processNewsSource fetches, analyzes, and sends news for a single source.
func processNewsSource(
	fetcher fetcher.Fetcher,
	geminiService *GeminiService,
	telegramService *TelegramService,
	config *Config,
	sourceName string,
	targetChannelIDs []string,
) {
	fmt.Printf("\n--- Fetching from %s ---\n", sourceName)
	items, err := fetcher.Fetch(time.Now().AddDate(0, 0, -1))
	if err != nil {
		errorMsg := fmt.Sprintf("Error fetching from %s: %v", sourceName, err)
		log.Print(errorMsg)
		telegramService.SendMessage(config.TelegramChatID, errorMsg)
		return
	}

	if len(items) == 0 {
		fmt.Printf("No new items from %s.\n", sourceName)
		return
	}

	fmt.Println("--- Analyzing News with Gemini ---")
	analysis, err := geminiService.AnalyzeNews(items)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to analyze news from %s: %v", sourceName, err)
		log.Print(errorMsg)
		telegramService.SendMessage(config.TelegramChatID, errorMsg)
		return
	}

	if analysis != "" && len(analysis) >= 34 {
		fmt.Println(analysis)
		// Escape triple asterisks to prevent Telegram Markdown parsing errors
		sanitizedAnalysis := strings.ReplaceAll(analysis, "***", "\\*\\*\\*")
		for _, channelID := range targetChannelIDs {
			// Add the channel ID to the message for context
			message := fmt.Sprintf("%s\n\n*%s*", sanitizedAnalysis, channelID)
			err = telegramService.SendMessage(channelID, message)
			if err != nil {
				log.Printf("Failed to send message to Telegram channel %s: %v", channelID, err)
			} else {
				// Send a confirmation to the admin chat with post text
				// Truncate the analysis to avoid exceeding Telegram's message size limit
				maxLength := 4000 // Telegram's message limit is around 4096 characters
				analysisPreview := sanitizedAnalysis
				if len(analysisPreview) > maxLength {
					analysisPreview = analysisPreview[:maxLength-3] + "..."
				}
				notification := fmt.Sprintf("Posted to %s: %s", channelID, analysisPreview)
				telegramService.SendMessage(config.TelegramChatID, notification)
			}
		}
	} else {
		fmt.Printf("No significant news to report from %s.\n", sourceName)
		telegramService.SendMessage(config.TelegramChatID, fmt.Sprintf("No significant news to report from %s.", sourceName))
	}
}

func main() {
	config := LoadConfig()
	geminiService := NewGeminiService(config.GeminiAPIKey)
	defer geminiService.Close()
	telegramService := NewTelegramService(config.TelegramAPIKey)

	// Process SVTV
	svtvFetcher := &fetcher.SvtvFetcher{URL: "https://svtv.org/feed/rss/"}
	processNewsSource(
		svtvFetcher,
		geminiService,
		telegramService,
		config,
		"SVTV",
		[]string{config.SvtvChannelID},
	)

	// Process Meduza
	meduzaFetcher := &fetcher.GenericFetcher{URL: "https://meduza.io/rss/all"}
	processNewsSource(
		meduzaFetcher,
		geminiService,
		telegramService,
		config,
		"Meduza",
		[]string{config.MeduzaChannelID},
	)
}
