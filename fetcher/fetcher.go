package fetcher

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// NewsItem represents a single news article.
type NewsItem struct {
	Title       string
	Link        string
	Content     string
	PublishedOn time.Time
}

// Fetcher is an interface for fetching news.
type Fetcher interface {
	Fetch(since time.Time) ([]NewsItem, error)
}

// cleanHTML removes HTML tags from a string.
func cleanHTML(rawHTML string) string {
	cleanr := regexp.MustCompile("<.*?>")
	cleantext := cleanr.ReplaceAllString(rawHTML, "")
	return cleantext
}

// GenericFetcher is a fetcher for standard RSS feeds.
type GenericFetcher struct {
	URL string
}

// Fetch fetches news from the feed.
func (f *GenericFetcher) Fetch(since time.Time) ([]NewsItem, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(f.URL)
	if err != nil {
		return nil, fmt.Errorf("error fetching or parsing feed: %w", err)
	}

	fmt.Printf("Fetching news from: %s\n", feed.Title)
	fmt.Println("------------------------------")

	var newsItems []NewsItem

	for _, item := range feed.Items {
		publishedTime := item.PublishedParsed
		if publishedTime == nil {
			log.Printf("Could not determine publication date for: %s", item.Title)
			continue
		}

		if publishedTime.After(since) {
			content := ""
			if item.Content != "" {
				content = item.Content
			} else if item.Description != "" {
				content = item.Description
			}

			newsItems = append(newsItems, NewsItem{
				Title:       item.Title,
				Link:        item.Link,
				Content:     cleanHTML(content),
				PublishedOn: *publishedTime,
			})
		}
	}
	return newsItems, nil
}

// SvtvFetcher is a custom fetcher for svtv.org.
type SvtvFetcher struct {
	URL string
}

var russianDateReplacer = strings.NewReplacer(
	"Янв", "Jan", "Фев", "Feb", "Мар", "Mar", "Апр", "Apr", "Май", "May", "Июн", "Jun",
	"Июл", "Jul", "Авг", "Aug", "Сен", "Sep", "Окт", "Oct", "Ноя", "Nov", "Дек", "Dec",
	"Пн", "Mon", "Вт", "Tue", "Ср", "Wed", "Чт", "Thu", "Пт", "Fri", "Сб", "Sat", "Вс", "Sun",
)

func parseRussianDate(dateStr string) (*time.Time, error) {
	englishDateStr := russianDateReplacer.Replace(dateStr)
	parsedTime, err := time.Parse(time.RFC1123Z, englishDateStr)
	if err != nil {
		return nil, err
	}
	return &parsedTime, nil
}

// Fetch fetches news from the svtv.org feed, handling its custom date format.
func (f *SvtvFetcher) Fetch(since time.Time) ([]NewsItem, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(f.URL)
	if err != nil {
		return nil, fmt.Errorf("error fetching or parsing feed: %w", err)
	}

	fmt.Printf("Fetching news from: %s\n", feed.Title)
	fmt.Println("------------------------------")

	var newsItems []NewsItem

	for _, item := range feed.Items {
		var publishedTime *time.Time
		// Try parsing with the library first
		if item.PublishedParsed != nil {
			publishedTime = item.PublishedParsed
		} else if item.Published != "" {
			// Fallback to manual parsing for Russian dates
			parsedTime, err := parseRussianDate(item.Published)
			if err != nil {
				log.Printf("Could not parse date '%s' for: %s", item.Published, item.Title)
				continue
			}
			publishedTime = parsedTime
		}

		if publishedTime == nil {
			log.Printf("Could not determine publication date for: %s", item.Title)
			continue
		}

		if publishedTime.After(since) {
			content := ""
			if item.Content != "" {
				content = item.Content
			} else if item.Description != "" {
				content = item.Description
			}

			newsItems = append(newsItems, NewsItem{
				Title:       item.Title,
				Link:        item.Link,
				Content:     cleanHTML(content),
				PublishedOn: *publishedTime,
			})
		}
	}
	return newsItems, nil
}