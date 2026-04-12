package main

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"news/fetcher"
)

func TestStorageSaveArticlesPostsAndRuns(t *testing.T) {
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
	if err := storage.SavePost(PostRecord{SourceName: "Source", TargetChannelID: "chan", MessageText: "msg", SentAt: time.Now().UTC()}, items); err != nil {
		t.Fatalf("SavePost() error = %v", err)
	}
	ok, err := storage.HasSentRecently("Source", "msg", 24*time.Hour)
	if err != nil {
		t.Fatalf("HasSentRecently() error = %v", err)
	}
	if !ok {
		t.Fatalf("expected HasSentRecently() to be true")
	}

	runID, err := storage.StartRun("Source")
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if err := storage.FinishRun(runID, "sent", ""); err != nil {
		t.Fatalf("FinishRun() error = %v", err)
	}

	var postArticles int
	if err := storage.db.QueryRow(`SELECT COUNT(1) FROM post_articles`).Scan(&postArticles); err != nil {
		t.Fatalf("count post_articles error = %v", err)
	}
	if postArticles != 1 {
		t.Fatalf("expected 1 post_articles row, got %d", postArticles)
	}

	var status string
	var finishedAt sql.NullTime
	if err := storage.db.QueryRow(`SELECT status, finished_at FROM runs WHERE id=?`, runID).Scan(&status, &finishedAt); err != nil {
		t.Fatalf("query run error = %v", err)
	}
	if status != "sent" {
		t.Fatalf("expected run status sent, got %s", status)
	}
	if !finishedAt.Valid {
		t.Fatalf("expected finished_at to be set")
	}
}
