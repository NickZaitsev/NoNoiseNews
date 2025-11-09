package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TelegramService handles sending messages to a Telegram bot.
type TelegramService struct {
	apiKey         string
	targetChannels map[string]string
}

// NewTelegramService creates a new TelegramService.
func NewTelegramService(apiKey string, targetChannels map[string]string) *TelegramService {
	return &TelegramService{
		apiKey:         apiKey,
		targetChannels: targetChannels,
	}
}

// SendMessage sends a message to the specified Telegram chat.
func (s *TelegramService) SendMessage(chatID, sourceName, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.apiKey)

	messageToSend := message

	// Check if the chatID is a target channel and append the identifier if so
	isTargetChannel := false
	for _, channel := range s.targetChannels {
		if channel == chatID {
			isTargetChannel = true
			break
		}
	}

	if isTargetChannel {
		if identifier, ok := s.targetChannels[sourceName]; ok {
			messageToSend = message + fmt.Sprintf("\n\n%s", identifier)
		}
	}

	requestBody, err := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       messageToSend,
		"parse_mode": "HTML",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Telegram API response: %s", string(body))
		return fmt.Errorf("failed to send message with status code: %d", resp.StatusCode)
	}

	log.Println("Message sent to Telegram successfully.")
	return nil
}

// SendPhoto sends a photo with a caption to the specified Telegram chat.
func (s *TelegramService) SendPhoto(chatID, photoURL, sourceName, caption string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", s.apiKey)

	fullCaption := caption

	// Check if the chatID is a target channel and append the identifier if so
	isTargetChannel := false
	for _, channel := range s.targetChannels {
		if channel == chatID {
			isTargetChannel = true
			break
		}
	}

	if isTargetChannel {
		if identifier, ok := s.targetChannels[sourceName]; ok {
			fullCaption = caption + fmt.Sprintf("\n\n%s", identifier)
		}
	}
	// log.Printf("Caption sent to telegram: %s", fullCaption)
	runes := []rune(fullCaption)
	if len(runes) > MaxTelegramCaptionLength {
		fullCaption = string(runes[:MaxTelegramCaptionLength-3]) + "..."
	}
	// Prepare the request body as JSON
	requestBody, err := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"photo":      photoURL,
		"caption":    fullCaption,
		"parse_mode": "HTML",
	})
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for sendPhoto: %w", err)
	}

	var lastErr error
	for i := 0; i < MaxPhotoRetries; i++ {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			lastErr = fmt.Errorf("failed to send photo by URL: %w", err)
			LogError("Failed to send photo", lastErr, "attempt", i+1)
			if i < MaxPhotoRetries-1 {
				time.Sleep(PhotoRetryDelay)
			}
			continue
		}

		if resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			LogInfo("Photo sent successfully by URL", "chat_id", chatID)
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastErr = fmt.Errorf("telegram API error: %s", string(body))
		LogError("Telegram API error while sending photo", lastErr, "status_code", resp.StatusCode, "response", string(body), "attempt", i+1)

		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			break
		}

		if i < MaxPhotoRetries-1 {
			time.Sleep(PhotoRetryDelay)
		}
	}

	LogError("All retries for sending photo failed, falling back to text message", lastErr, "chat_id", chatID)

	fallbackMessage := fmt.Sprintf("%s\n\n(Image: %s)", caption, photoURL)
	return s.SendMessage(chatID, sourceName, fallbackMessage)
}
