package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration.
type Config struct {
	GeminiAPIKey        string
	TelegramAPIKey      string
	TelegramChatID      string
	NewsSources         map[string]string
	TargetChannels      map[string]string
	ContentPreviewLimit int
	MaxMessageLength    int
	APITimeout          int
}

// LoadConfig loads the configuration from a .env file.
func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Load required API keys
	geminiAPIKey := getEnv("GEMINI_API_KEY", true)
	telegramAPIKey := getEnv("TELEGRAM_API_KEY", true)
	telegramChatID := getEnv("TELEGRAM_CHAT_ID", true)

	// Load optional settings with defaults
	contentPreviewLimit := getEnvAsInt("CONTENT_PREVIEW_LIMIT", ContentPreviewLimit)
	maxMessageLength := getEnvAsInt("MAX_MESSAGE_LENGTH", MaxMessageLength)
	apiTimeout := getEnvAsInt("API_TIMEOUT", int(DefaultHTTPTimeout/time.Second))

	// Load news sources from environment variable
	newsSourcesEnv := getEnv("NEWS_SOURCES", true)
	newsSources := parseNewsSources(newsSourcesEnv)

	// Load target channels from environment variable
	targetChannelsEnv := getEnv("TARGET_CHANNELS", true)
	targetChannels := parseTargetChannels(targetChannelsEnv)

	return &Config{
		GeminiAPIKey:        geminiAPIKey,
		TelegramAPIKey:      telegramAPIKey,
		TelegramChatID:      telegramChatID,
		NewsSources:         newsSources,
		TargetChannels:      targetChannels,
		ContentPreviewLimit: contentPreviewLimit,
		MaxMessageLength:    maxMessageLength,
		APITimeout:          apiTimeout,
	}, nil
}

// getEnv retrieves an environment variable and validates it if required.
func getEnv(key string, required bool) string {
	value := os.Getenv(key)
	if required && value == "" {
		log.Fatalf("%s is not set in .env file", key)
	}
	return value
}

// getEnvAsInt retrieves an environment variable and converts it to an integer.
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// parseNewsSources parses the NEWS_SOURCES environment variable.
func parseNewsSources(newsSourcesEnv string) map[string]string {
	sources := make(map[string]string)
	
	// Expected format: "name1:url1,name2:url2,name3:url3"
	pairs := strings.Split(newsSourcesEnv, ",")
	
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			url := strings.TrimSpace(parts[1])
			if name != "" && url != "" {
				sources[name] = url
			}
		}
	}
	
	if len(sources) == 0 {
		log.Fatal("No valid news sources found in NEWS_SOURCES environment variable")
	}
	
	return sources
}

// parseTargetChannels parses the TARGET_CHANNELS environment variable.
func parseTargetChannels(targetChannelsEnv string) map[string]string {
	channels := make(map[string]string)
	
	// Expected format: "SourceName:ChannelID,SourceName2:ChannelID2"
	pairs := strings.Split(targetChannelsEnv, ",")
	
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			sourceName := strings.TrimSpace(parts[0])
			channelID := strings.TrimSpace(parts[1])
			if sourceName != "" && channelID != "" {
				channels[sourceName] = channelID
			}
		}
	}
	
	if len(channels) == 0 {
		log.Fatal("No valid target channels found in TARGET_CHANNELS environment variable")
	}
	
	return channels
}