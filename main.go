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
	storage *Storage,
	config *Config,
	sourceName string,
	targetChannelIDs []string,
) {
	// Step 1: Fetch news
	items, err := fetchNews(fetcher, sourceName, config)
	if err != nil {
		handleError(telegramService, config.TelegramChatID, sourceName, err, "fetching")
		return
	}

	// Step 2: Display content preview
	displayContentPreview(items, sourceName)

	if err := storage.SaveArticles(sourceName, items); err != nil {
		handleError(telegramService, config.TelegramChatID, sourceName, err, "saving_articles")
		return
	}

	// Step 3: Check if we have any items
	if len(items) == 0 {
		handleNoNews(telegramService, config.TelegramChatID, sourceName)
		return
	}

	// Step 4: Analyze news with Gemini
	geminiImageURL, analysis, err := analyzeNews(geminiService, items, sourceName, config)
	if err != nil {
		handleError(telegramService, config.TelegramChatID, sourceName, err, "analyzing")
		return
	}

	// Step 5: Send notifications
	sendNotifications(telegramService, storage, config, analysis, geminiImageURL, items, targetChannelIDs, sourceName)
}

// fetchNews retrieves news items from the given fetcher.
func fetchNews(fetcher fetcher.Fetcher, sourceName string, config *Config) ([]fetcher.NewsItem, error) {
	fmt.Printf("\n--- Fetching from %s ---\n", sourceName)
	return fetcher.Fetch(time.Now().AddDate(0, 0, -1), config.RetryAttempts, config.RetryDelay)
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
func analyzeNews(geminiService *GeminiService, items []fetcher.NewsItem, _ string, config *Config) (string, string, error) {
	fmt.Println("--- Analyzing News with Gemini ---")
	return geminiService.AnalyzeNews(items, config.RetryAttempts, config.RetryDelay)
}

// sendNotifications sends the analysis to the specified Telegram channels.
func sendNotifications(telegramService *TelegramService, storage *Storage, config *Config, analysis, geminiImageURL string, items []fetcher.NewsItem, targetChannelIDs []string, sourceName string) {
	adminChatID := config.TelegramChatID
	if analysis != "" && len(strings.TrimSpace(analysis)) >= 34 {
		fmt.Println(analysis)
		sanitizedAnalysis := strings.ReplaceAll(analysis, TelegramMarkdownEscape, "\\*\\*\\*")

		// Prioritize Gemini image URL, otherwise fall back to the first item's image
		var bestImageURL string
		if geminiImageURL != "" {
			bestImageURL = geminiImageURL
		} else if len(items) > 0 {
			bestImageURL = items[0].ImageURL
		}

		duplicateWindow := time.Duration(config.DuplicateWindowHours) * time.Hour
		for _, channelID := range targetChannelIDs {
			alreadySent, err := storage.HasSentRecently(sourceName, sanitizedAnalysis, duplicateWindow)
			if err != nil {
				handleError(telegramService, adminChatID, sourceName, err, "checking_duplicate_post")
				continue
			}
			if alreadySent {
				LogInfo("Skipping duplicate post", "source", sourceName, "channel_id", channelID)
				notifyAdmin(telegramService, adminChatID, sourceName, fmt.Sprintf("Skipped duplicate post for %s to %s", sourceName, channelID))
				continue
			}
			sendToChannel(telegramService, storage, adminChatID, sanitizedAnalysis, bestImageURL, channelID, sourceName)
		}
	} else {
		fmt.Printf("No significant news to report from %s.\n", sourceName)
		notifyAdmin(telegramService, adminChatID, sourceName, fmt.Sprintf("No significant news to report from %s.", sourceName))
	}
}

// sendToChannel handles sending the news analysis to the appropriate channel.
func sendToChannel(telegramService *TelegramService, storage *Storage, adminChatID, message, photoURL, channelID, sourceName string) {
	var err error
	if photoURL != "" {
		err = telegramService.SendPhoto(channelID, photoURL, sourceName, message)
		if err != nil {
			LogError("Failed to send photo, falling back to text message", err, "channel_id", channelID, "photo_url", photoURL)
			notifyAdmin(telegramService, adminChatID, sourceName, fmt.Sprintf("Failed to send photo from %s to %s. Error: %v. Falling back to text.", sourceName, channelID, err))
			// Fallback to sending the original full message as text
			err = telegramService.SendMessage(channelID, sourceName, message)
		}
	} else {
		err = telegramService.SendMessage(channelID, sourceName, message)
	}

	if err != nil {
		LogError("Failed to send final message to Telegram channel", err, "channel_id", channelID, "source", sourceName)
		notifyAdmin(telegramService, adminChatID, sourceName, fmt.Sprintf("Failed to send news from %s to %s: %v", sourceName, channelID, err))
	} else {
		if err := storage.SavePost(PostRecord{SourceName: sourceName, TargetChannelID: channelID, MessageText: message, ImageURL: photoURL, SentAt: time.Now()}); err != nil {
			LogError("Failed to persist sent post", err, "channel_id", channelID, "source", sourceName)
			notifyAdmin(telegramService, adminChatID, sourceName, fmt.Sprintf("Posted to %s from %s, but failed to save post record: %v", channelID, sourceName, err))
		}
		notification := fmt.Sprintf("News posted to %s from %s", channelID, sourceName)
		if photoURL != "" {
			notification += " (with photo)"
		}
		LogInfo("News posted successfully", "channel_id", channelID, "source", sourceName)
		notifyAdmin(telegramService, adminChatID, sourceName, notification)
	}
}

func notifyAdmin(telegramService *TelegramService, adminChatID, sourceName, message string) {
	if err := telegramService.SendMessage(adminChatID, sourceName, message); err != nil {
		LogError("Failed to send admin notification", err, "chat_id", adminChatID, "source", sourceName)
	}
}

// handleError logs and sends an error message about a failed operation.
func handleError(telegramService *TelegramService, adminChatID, sourceName string, err error, operation string) {
	LogError("Operation failed", err, "operation", operation, "source", sourceName)
	errorMsg := fmt.Sprintf("Error %s from %s: %v", operation, sourceName, err)
	notifyAdmin(telegramService, adminChatID, sourceName, errorMsg)
}

// handleNoNews handles the case when no news items are found.
func handleNoNews(telegramService *TelegramService, adminChatID, sourceName string) {
	fmt.Printf("No new items from %s.\n", sourceName)
	notifyAdmin(telegramService, adminChatID, sourceName, fmt.Sprintf("No new items from %s.", sourceName))
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

	storage, err := NewStorage(config.DatabasePath)
	if err != nil {
		LogError("Failed to initialize storage", err, "database_path", config.DatabasePath)
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	geminiService := NewGeminiService(config.GeminiAPIKey, config.GeminiPrompt, config.GeminiModel, config.APITimeout)
	defer geminiService.Close()
	telegramService := NewTelegramService(config.TelegramAPIKey, config.TargetChannels)

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
			storage,
			config,
			sourceName,
			channelIDs,
		)
	}

	LogInfo("News fetching completed for all sources")
}
