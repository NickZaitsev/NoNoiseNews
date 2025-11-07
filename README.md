# NoNoise News Fetcher

NoNoise is a Go-based intelligent news aggregation application that remains silent most of the time and speaks only when truly important news is detected.

The application uses AI-powered analysis to relieve readers of the harm caused by an excessive amount of “news” that carries no signal, only noise.

Program takes news from RSS, take the most important ones, and post it to a telegram channel.


## Architecture

```
Configuration (.env)
       ↓
   config.go (loads settings)
       ↓
   main.go (orchestrates workflow)
       ↓
   fetcher/ (news retrieval)
       ↓
   gemini.go (AI analysis)
       ↓
 telegram.go (notifications)
       ↓
   Console & Telegram Channels
```

## Project Structure

- **`main.go`**: Application entry point and workflow orchestration
- **`config.go`**: Configuration loading and management
- **`fetcher/`**: Modular news fetching system
  - **`fetcher.go`**: Fetcher interface and implementations
  - **`GenericFetcher`**: Standard RSS/Atom feed parser
  - **`SvtvFetcher`**: Custom parser for non-standard feed formats
- **`gemini.go`**: Google Gemini AI integration for news analysis
- **`telegram.go`**: Telegram bot API integration
- **`logger.go`**: Structured logging system
- **`constants.go`**: Application constants and configuration defaults

## Current Implementation Status

The NoNoise application is **currently functional and ready to run**, implementing a simplified but effective version of the NoNoise philosophy.

## Technology Stack

- **Core**: Go 1.22.5
- **AI Integration**: HTTP-based Gemini API (no external SDK)
- **RSS Processing**: github.com/mmcdole/gofeed v1.3.0
- **Configuration**: github.com/joho/godotenv v1.5.1
- **HTTP Client**: Standard Go net/http package

## Getting Started

### Prerequisites

- Go 1.22.5 or higher
- Google Gemini API key
- Telegram Bot API key and Chat ID

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd nonoise
   ```

2. **Create configuration file:**
   ```bash
   cp .env.example .env
   ```

3. **Configure your environment variables:**
   Edit the `.env` file with your actual API keys and settings:
   ```env
   # Required: Google Gemini API Key
   GEMINI_API_KEY=your_gemini_api_key_here

   # Required: Telegram Bot API Key
   TELEGRAM_API_KEY=your_telegram_bot_api_key_here

   # Required: Telegram Chat ID for admin notifications
   TELEGRAM_CHAT_ID=your_telegram_chat_id_here

   # Required: News Sources Configuration
   NEWS_SOURCES=SVTV:https://svtv.org/feed/rss/,Meduza:https://meduza.io/rss/all
   ```

4. **Install dependencies:**
   ```bash
   go mod tidy
   ```

### Running the Application

#### Development Mode
```bash
go run .
```

#### Production Build
```bash
go build -o nonoise .
./nonoise
```

The application will:
1. Load configuration from `.env`
2. Fetch news from configured RSS sources
3. Analyze content using Gemini AI
4. Send significant news to configured Telegram channels
5. Provide structured logging output

## Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Google Gemini API key | `AIzaSy...` |
| `TELEGRAM_API_KEY` | Telegram bot API key | `1234567890:ABC...` |
| `TELEGRAM_CHAT_ID` | Admin chat ID for notifications | `-1001234567890` |
| `NEWS_SOURCES` | News sources configuration | `SVTV:https://svtv.org/feed/rss/` |

### Optional Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `CONTENT_PREVIEW_LIMIT` | Content preview characters | `1000` |
| `MAX_MESSAGE_LENGTH` | Telegram message limit | `4000` |
| `API_TIMEOUT` | HTTP request timeout (seconds) | `30` |

### Adding New News Sources

To add a new news source, update the `NEWS_SOURCES` variable in your `.env` file:

```env
NEWS_SOURCES=SVTV:https://svtv.org/feed/rss/,Meduza:https://meduza.io/rss/all,NewSource:https://example.com/rss
```

For sources with non-standard formats, extend the `fetcher` package with custom parsing logic.
