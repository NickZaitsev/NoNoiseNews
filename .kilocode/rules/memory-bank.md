# NoNoise News Fetcher - Memory Bank

## Project Overview
NoNoise is a Go-based news fetching application designed to aggregate articles from various RSS feeds. It aligns with the "NoNoise" philosophy by providing a framework to selectively gather news, with custom logic to handle non-standard feed formats. It now uses the Gemini API to analyze and summarize the most important news.

**Status**: A modular Go application with Gemini integration using HTTP-based API calls.

## Core Components
- **`main.go`**: The application's entry point. It initializes the configuration, fetches news, and uses the Gemini service to analyze and print the results.
- **`config.go`**: Loads configuration from a `.env` file, including API keys for Gemini and Telegram.
- **`fetcher` package**: Contains all the logic for fetching news from RSS feeds.
  - **`Fetcher` Interface**: Defines the contract for all news fetchers.
  - **`GenericFetcher`**: A fetcher for standard RSS/Atom feeds.
  - **`SvtvFetcher`**: A custom fetcher for `svtv.org` with special date parsing.
- **`gemini.go`**: Interacts with the Gemini API to analyze and summarize news articles.
- **`telegram.go`**: Sends messages to a Telegram bot.

## Key Features
- Fetches news from multiple RSS feeds.
- Cleans HTML tags from the content of news articles.
- Extensible design using a `Fetcher` interface to support new sources.
- Custom date parsing logic to handle non-standard formats.
- Integrates with the Gemini API to analyze and summarize news.
- Sends significant news summaries to a Telegram bot.

## Technology Stack
- **Core**: Go 1.22.5
- **Dependencies**: 
  - `github.com/mmcdole/gofeed` v1.3.0
  - `github.com/joho/godotenv` v1.5.1
  - HTTP-based Gemini API integration

## Architecture Flow
```
.env -> config.go -> main.go -> fetcher package -> HTTP-based Gemini API -> Telegram & Console
```

## Critical Implementation Paths
1. **News Fetching**: `main.go` -> `fetcher.Fetch()`
2. **Gemini Analysis**: `main.go` -> HTTP-based `Gemini API Call`
3. **Telegram Notification**: `main.go` -> `telegram.SendMessage()`

## Key Dependencies
- `github.com/mmcdole/gofeed`
- `github.com/joho/godotenv`
- Standard HTTP client (no external Gemini SDK dependency)

## Technical Constraints
- News sources are hardcoded within the `main` function.
- The application prints news to the console and sends it to Telegram; there is no storage mechanism.
- Error handling is basic, logging messages to standard output.
- Uses direct HTTP API calls to Gemini instead of official SDK due to dependency issues.

## Recent Changes
- **Telegram Integration**: Added functionality to send Gemini's output to a Telegram bot.
- **Attempted SDK Migration**: Tried to migrate to the official Google GenAI Go SDK (`google.golang.org/genai`)
- **Dependency Conflicts**: Encountered version compatibility issues with the latest SDK versions
- **Reverted to HTTP**: Maintained the existing HTTP-based implementation for reliability
- **Clean Dependencies**: Updated go.mod to remove problematic SDK dependencies
- **Successful Build**: Application builds successfully and is ready to run