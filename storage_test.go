package main

import (
	"path/filepath"
	"testing"
	"time"

	"news/fetcher"
)

func TestStorageSaveArticlesAndPosts(t *testing.T) {
	storage, err := NewStorage(filepath.Join(t.TempDir(), "nonoise.db"))
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}
	defer storage.Close()

	items := []fetcher.NewsItem{{
		Title:       "Hello",
		Link:        "https://example.com/a",
		Content:     "content",
		RawContent:  "<p>content</p>",
		PublishedOn: time.Now().UTC(),
		ImageURL:    "https://example.com/a.jpg",
	}}
	if err := storage.SaveArticles("Source", items); err != nil {
		t.Fatalf("SaveArticles() error = %v", err)
	}
	if err := storage.SavePost(PostRecord{SourceName: "Source", TargetChannelID: "chan", MessageText: "msg", SentAt: time.Now().UTC()}); err != nil {
		t.Fatalf("SavePost() error = %v", err)
	}
	ok, err := storage.HasSentRecently("Source", "msg", 24*time.Hour)
	if err != nil {
		t.Fatalf("HasSentRecently() error = %v", err)
	}
	if !ok {
		t.Fatalf("expected HasSentRecently() to be true")
	}
}
