# NoNoise News Fetcher - Memory Bank

## Project Overview
NoNoise is a Go-based news fetching application designed to aggregate articles from various RSS feeds. It aligns with the "NoNoise" philosophy by providing a framework to selectively gather news, with custom logic to handle non-standard feed formats. It now uses the official Google GenAI Go SDK to analyze and summarize the most important news.

**Status**: A modular Go application with official Gemini SDK integration, structured logging, and comprehensive configuration management.

## Core Components
- **`main.go`**: Refactored application entry point with configuration-based source management, structured error handling, and modular workflow orchestration.
- **`config.go`**: Comprehensive configuration loader that parses environment variables for API keys, news sources, target channels, and application settings.
- **`logger.go`**: Structured logging implementation using Go's built-in `slog` package for production-ready logging.
- **`constants.go`**: Centralized constants for HTTP timeouts, content limits, Telegram settings, and Russian date parsing.
- **`fetcher` package**: Contains all the logic for fetching news from RSS feeds.
  - **`Fetcher` Interface**: Defines the contract for all news fetchers.
  - **`GenericFetcher`**: A fetcher for standard RSS/Atom feeds with proper HTTP client configuration.
  - **`SvtvFetcher`**: A custom fetcher for `svtv.org` with special Russian date parsing.
- **`gemini.go`**: Interacts with the official Google GenAI Go SDK to analyze and summarize news articles.
- **`telegram.go`**: Enhanced Telegram service with proper error handling and message length management.

## Key Features
- Fetches news from multiple RSS feeds with configuration-based source management.
- Cleans HTML tags from the content of news articles.
- Extensible design using a `Fetcher` interface to support new sources.
- Custom date parsing logic to handle non-standard formats (Russian dates).
- Integrates with the official Gemini API Go SDK for intelligent news analysis.
- Sends significant news summaries to configurable Telegram channels.
- Structured logging with context-aware error tracking.
- Environment-based configuration for all settings.

## Technology Stack
- **Core**: Go 1.22
- **Dependencies**: 
  - `github.com/mmcdole/gofeed` v1.3.0
  - `github.com/joho/godotenv` v1.5.1
  - `google.golang.org/genai` (official Gemini SDK)
  - `google.golang.org/api/option` (Google API options)

## Architecture Flow
```
.env -> config.go -> main.go -> fetcher package -> Official Gemini SDK -> Telegram & Structured Logging
```

## Critical Implementation Paths
1. **Configuration Loading**: `main.go` -> `LoadConfig()` -> Environment Variables
2. **News Fetching**: `main.go` -> `fetcher.Fetch()` -> RSS Parsing
3. **Gemini Analysis**: `main.go` -> `gemini.AnalyzeNews()` -> Official SDK Call
4. **Telegram Notification**: `main.go` -> `telegram.SendMessage()` -> Channel Distribution
5. **Logging**: All operations -> `logger.go` -> Structured JSON Output

## Key Dependencies
- `github.com/mmcdole/gofeed` - RSS feed parsing
- `github.com/joho/godotenv` - Environment variable loading
- `google.golang.org/genai` - Official Gemini AI SDK
- `google.golang.org/api/option` - Google API configuration options
- Standard library `log/slog` - Structured logging

## Configuration Management
- **Environment Variables**: All configuration loaded from `.env` file
- **News Sources**: Configurable via `NEWS_SOURCES` (format: "name:url,name2:url2")
- **Target Channels**: Configurable via `TARGET_CHANNELS` (format: "SourceName:ChannelID")
- **API Settings**: Content preview limits, message length limits, timeouts all configurable
- **Required Variables**: `GEMINI_API_KEY`, `TELEGRAM_API_KEY`, `TELEGRAM_CHAT_ID`

## Technical Constraints
- News sources and target channels are now configured via environment variables, not hardcoded.
- The application maintains no storage mechanism - processes and distributes news in real-time.
- Error handling is comprehensive with structured logging and Telegram notifications.
- Uses official Google GenAI Go SDK instead of raw HTTP calls for better reliability and features.

## Recent Changes
- **Official SDK Migration**: Successfully migrated to Google GenAI Go SDK (`google.golang.org/genai`)
- **Configuration Management**: Implemented comprehensive environment-based configuration
- **Structured Logging**: Added production-ready structured logging with `slog`
- **Architecture Refactoring**: Complete main.go refactor with modular workflow functions
- **Error Handling**: Enhanced error handling with context and structured logging
- **Telegram Integration**: Improved with message length management and multi-channel support
- **Constants Centralization**: Moved all constants to dedicated `constants.go` file
- **Dependency Management**: Clean dependency tree with proper indirect dependency handling
- **Russian Date Support**: Enhanced support for Russian date parsing in SvtvFetcher
- **HTTP Client Standardization**: Implemented consistent HTTP client configuration across fetchers

## Deployment & Operations
- **Environment-Based**: All configuration externalized to environment variables
- **Structured Logging**: JSON-formatted logs for production monitoring
- **Error Notifications**: Automatic error reporting to Telegram admin chat
- **Health Monitoring**: Comprehensive logging of all operations and failures
- **Resource Management**: Proper HTTP client cleanup and connection management