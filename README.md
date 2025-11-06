# Concept NoNoise

NoNoise is a post-news publication that does not kill attention. It is radically different from traditional media outlets: it remains silent most of the time and speaks only when it is truly important. Its mission is to relieve readers of information iatrogenesis—the harm caused by an excessive amount of “news” that carries no signal, only noise.

## Formats

*   **NoNoise10** — 10 news stories per year. Each publication is an event of tectonic significance: war, technological breakthrough, scientific discovery, change in the structure of the world.
*   **NoNoise50** — 50 articles per year. A more responsive format: key signals in economics, science, ecology, and technology — without haste or panic.

## Editorial philosophy

*   **Via negativa** — the main thing is not to add, but to remove. We filter out everything that has no long-term significance.
*   **Barbell strategy** — 95% of the time is “conservative silence,” 5% is publications of extreme usefulness.

## Why it matters

Information noise is toxic. Modern media creates neurosis: an endless stream of “events” makes people anxious and stupid.

Frequency kills understanding. When a person follows events 24/7, 99.5% of the signals are noise.

We bring back real time. The reader gets space to think and make decisions, rather than an endless panorama of anxiety.

## Getting Started

### Prerequisites

- Go 1.22.5 or higher
- A Gemini API key
- A Telegram Bot API key and Chat ID

### Installation

1.  **Clone the repository:**
    ```
    git clone https://github.com/your-username/news.git
    cd news
    ```

2.  **Create a `.env` file:**
    Create a `.env` file in the root of the project and add your API keys and chat ID:
    ```
    GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
    TELEGRAM_API_KEY="YOUR_TELEGRAM_API_KEY"
    TELEGRAM_CHAT_ID="YOUR_TELEGRAM_CHAT_ID"
    ```

3.  **Install dependencies:**
    ```
    go mod tidy
    ```

### Running the Application

To run the application, execute the following command:
```
go run .
```

The application will fetch news from the specified RSS feeds, analyze them using the Gemini API, and print the most significant news story to the console. If a significant news story is found, it will also be sent to the configured Telegram chat.