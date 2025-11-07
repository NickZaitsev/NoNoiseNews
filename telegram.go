package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// TelegramService handles sending messages to a Telegram bot.
type TelegramService struct {
	apiKey string
}

// NewTelegramService creates a new TelegramService.
func NewTelegramService(apiKey string) *TelegramService {
	return &TelegramService{
		apiKey: apiKey,
	}
}

// SendMessage sends a message to the specified Telegram chat.
func (s *TelegramService) SendMessage(chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.apiKey)

	requestBody, err := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       message,
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
