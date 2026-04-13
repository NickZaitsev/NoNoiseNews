package main

import (
	"time"
)

// HTTP constants
const (
	// Timeouts
	DefaultHTTPTimeout = 30 * time.Second
	DefaultAPITimeout  = 60 * time.Second

	// Retry mechanism
	DefaultRetryAttempts           = 3
	DefaultRetryDelay              = 2 * time.Second
	DefaultGeminiRetryAttempts     = 4
	DefaultGeminiRateLimitAttempts = 6
	MaxGeminiRetryDelay            = 2 * time.Minute

	// Content limits
	ContentPreviewLimit = 1000
	MaxMessageLength    = 4000

	// Storage
	DefaultDatabasePath         = "nonoise.db"
	DefaultDuplicateWindowHours = 24

	// Telegram constants
	TelegramMarkdownEscape   = "***"
	MaxTelegramCaptionLength = 1024
	MaxPhotoRetries          = 3
	PhotoRetryDelay          = 3 * time.Second
)

// User agent and headers for HTTP requests
const (
	UserAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"
	Accept     = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
	AcceptLang = "en-US,en;q=0.9"
)

// Gemini API constants
const (
	GeminiModel                 = "gemini-2.5-flash"
	DefaultGeminiItemCharLimit  = 1400
	DefaultGeminiTotalCharLimit = 9000
	MinimumGeminiItemCharLimit  = 500
	MinimumGeminiTotalCharLimit = 2500
	GeminiLimitShrinkPercent    = 70
)

// Russian date parsing constants
var RussianMonthReplacer = map[string]string{
	"Янв": "Jan", "Фев": "Feb", "Мар": "Mar", "Апр": "Apr",
	"Май": "May", "Июн": "Jun", "Июл": "Jul", "Авг": "Aug",
	"Сен": "Sep", "Окт": "Oct", "Ноя": "Nov", "Дек": "Dec",
}

var RussianDayReplacer = map[string]string{
	"Пн": "Mon", "Вт": "Tue", "Ср": "Wed", "Чт": "Thu",
	"Пт": "Fri", "Сб": "Sat", "Вс": "Sun",
}

// News source configurations
const (
	SVTVSourceName   = "SVTV"
	MeduzaSourceName = "Meduza"
)
