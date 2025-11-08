package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"news/fetcher"
)

// processNewsSource orchestrates the entire news processing workflow for a single source.
func processNewsSource(
	fetcher fetcher.Fetcher,
	geminiService *GeminiService,
	telegramService *TelegramService,
	config *Config,
	sourceName string,
	targetChannelIDs []string,
) {
	// Step 1: Fetch news
	items, err := fetchNews(fetcher, sourceName)
	if err != nil {
		handleError(telegramService, config.TelegramChatID, sourceName, err, "fetching")
		return
	}

	// Step 2: Display content preview
	displayContentPreview(items, sourceName)

	// Step 3: Check if we have any items
	if len(items) == 0 {
		handleNoNews(telegramService, config.TelegramChatID, sourceName)
		return
	}

	// Step 4: Analyze news with Gemini
	analysis, err := analyzeNews(geminiService, items, sourceName)
	if err != nil {
		handleError(telegramService, config.TelegramChatID, sourceName, err, "analyzing")
		return
	}

	// Step 5: Send notifications
	sendNotifications(telegramService, config.TelegramChatID, analysis, targetChannelIDs, sourceName)
}

// fetchNews retrieves news items from the given fetcher.
func fetchNews(fetcher fetcher.Fetcher, sourceName string) ([]fetcher.NewsItem, error) {
	fmt.Printf("\n--- Fetching from %s ---\n", sourceName)
	return fetcher.Fetch(time.Now().AddDate(0, 0, -1))
}

// displayContentPreview shows a preview of the first news item's content.
func displayContentPreview(items []fetcher.NewsItem, _ string) {
	if len(items) > 0 && items[0].Content != "" {
		contentPreview := items[0].Content
		if len(contentPreview) > ContentPreviewLimit {
			contentPreview = contentPreview[:ContentPreviewLimit]
		}
		fmt.Printf("RSS Content Preview: %s\n", contentPreview)
	}
}

// analyzeNews uses Gemini AI to analyze and summarize the news items.
func analyzeNews(geminiService *GeminiService, items []fetcher.NewsItem, _ string) (string, error) {
	fmt.Println("--- Analyzing News with Gemini ---")
	return geminiService.AnalyzeNews(items)
}

// sendNotifications sends the analysis to the specified Telegram channels.
func sendNotifications(telegramService *TelegramService, adminChatID, analysis string, targetChannelIDs []string, sourceName string) {
	if analysis != "" && len(analysis) >= 34 {
		fmt.Println(analysis)
		// Escape triple asterisks to prevent Telegram Markdown parsing errors
		sanitizedAnalysis := strings.ReplaceAll(analysis, TelegramMarkdownEscape, "\\*\\*\\*")
		for _, channelID := range targetChannelIDs {
			sendToChannel(telegramService, adminChatID, sanitizedAnalysis, channelID, sourceName)
		}
	} else {
		fmt.Printf("No significant news to report from %s.\n", sourceName)
		telegramService.SendMessage(adminChatID, fmt.Sprintf("No significant news to report from %s.", sourceName))
	}
}

// sendToChannel handles sending the news analysis to the appropriate channel.
func sendToChannel(telegramService *TelegramService, adminChatID, sanitizedAnalysis, channelID, sourceName string) {
	lines := strings.SplitN(sanitizedAnalysis, "\n", 2)
	photoURL := ""
	message := sanitizedAnalysis

	if len(lines) > 1 && (strings.HasPrefix(lines[0], "http://") || strings.HasPrefix(lines[0], "https://")) {
		photoURL = lines[0]
		message = lines[1]
	}

	var err error
	if photoURL != "" {
		err = telegramService.SendPhoto(channelID, photoURL, message)
		if err != nil {
			LogError("Failed to send photo, falling back to text message", err, "channel_id", channelID, "photo_url", photoURL)
			telegramService.SendMessage(adminChatID, fmt.Sprintf("Failed to send photo from %s to %s. Error: %v. Falling back to text.", sourceName, channelID, err))
			// Fallback to sending the original full message as text
			err = telegramService.SendMessage(channelID, sanitizedAnalysis)
		}
	} else {
		err = telegramService.SendMessage(channelID, message)
	}

	if err != nil {
		LogError("Failed to send final message to Telegram channel", err, "channel_id", channelID, "source", sourceName)
		telegramService.SendMessage(adminChatID, fmt.Sprintf("Failed to send news from %s to %s: %v", sourceName, channelID, err))
	} else {
		notification := fmt.Sprintf("News posted to %s from %s", channelID, sourceName)
		if photoURL != "" {
			notification += " (with photo)"
		}
		LogInfo("News posted successfully", "channel_id", channelID, "source", sourceName)
		telegramService.SendMessage(adminChatID, notification)
	}
}

// handleError logs and sends an error message about a failed operation.
func handleError(telegramService *TelegramService, adminChatID, sourceName string, err error, operation string) {
	LogError("Operation failed", err, "operation", operation, "source", sourceName)
	errorMsg := fmt.Sprintf("Error %s from %s: %v", operation, sourceName, err)
	telegramService.SendMessage(adminChatID, errorMsg)
}

// handleNoNews handles the case when no news items are found.
func handleNoNews(telegramService *TelegramService, adminChatID, sourceName string) {
	fmt.Printf("No new items from %s.\n", sourceName)
	telegramService.SendMessage(adminChatID, fmt.Sprintf("No new items from %s.", sourceName))
}

func main() {
	// Initialize structured logging
	initLogger()
	LogInfo("Starting NoNoise news fetcher", "version", "1.0.0")
	
	config, err := LoadConfig()
	if err != nil {
		LogError("Failed to load configuration", err)
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	geminiService := NewGeminiService(config.GeminiAPIKey, config.GeminiPrompt)
	defer geminiService.Close()
	telegramService := NewTelegramService(config.TelegramAPIKey)

	// Process each news source from configuration
	for sourceName, sourceURL := range config.NewsSources {
		var fetcherObj fetcher.Fetcher
		
		// Create appropriate fetcher based on source
		if sourceName == SVTVSourceName {
			fetcherObj = &fetcher.SvtvFetcher{URL: sourceURL}
		} else {
			fetcherObj = &fetcher.GenericFetcher{URL: sourceURL}
		}
		
		// Get the target channel for this source
		targetChannel, exists := config.TargetChannels[sourceName]
		if !exists {
			LogError("No target channel configured for source", nil, "source", sourceName)
			continue
		}
		
		// Use the target channel for news, admin chat ID for notifications
		channelIDs := []string{targetChannel}
		
		processNewsSource(
			fetcherObj,
			geminiService,
			telegramService,
			config,
			sourceName,
			channelIDs,
		)
	}
	
	LogInfo("News fetching completed for all sources")
}
