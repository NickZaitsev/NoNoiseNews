package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"news/fetcher"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

type ArticleRecord struct {
	ID          int64
	SourceName  string
	Title       string
	Link        string
	Content     string
	RawContent  string
	PublishedAt time.Time
	ImageURL    string
	FetchedAt   time.Time
}

type PostRecord struct {
	SourceName      string
	TargetChannelID string
	MessageText     string
	ImageURL        string
	SentAt          time.Time
}

type RunRecord struct {
	ID         int64
	SourceName string
	Status     string
	StartedAt  time.Time
	FinishedAt sql.NullTime
	ErrorText  sql.NullString
}

func NewStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.init(); err != nil {
		db.Close()
		return nil, err
	}
	return storage, nil
}

func (s *Storage) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Storage) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_name TEXT NOT NULL,
		title TEXT NOT NULL,
		link TEXT NOT NULL,
		content TEXT NOT NULL,
		raw_content TEXT NOT NULL,
		published_at TIMESTAMP NOT NULL,
		image_url TEXT,
		fetched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_name, link)
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_name TEXT NOT NULL,
		target_channel_id TEXT NOT NULL,
		message_text TEXT NOT NULL,
		image_url TEXT,
		sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS post_articles (
		post_id INTEGER NOT NULL,
		article_id INTEGER NOT NULL,
		PRIMARY KEY (post_id, article_id),
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_name TEXT NOT NULL,
		status TEXT NOT NULL,
		started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		finished_at TIMESTAMP,
		error_text TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_articles_source_published ON articles(source_name, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_posts_source_sent ON posts(source_name, sent_at DESC);
	CREATE INDEX IF NOT EXISTS idx_runs_source_started ON runs(source_name, started_at DESC);
	`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}
	return nil
}

func (s *Storage) StartRun(sourceName string) (int64, error) {
	result, err := s.db.Exec(`INSERT INTO runs (source_name, status) VALUES (?, 'running')`, sourceName)
	if err != nil {
		return 0, fmt.Errorf("start run: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("start run last insert id: %w", err)
	}
	return id, nil
}

func (s *Storage) FinishRun(runID int64, status string, errText string) error {
	var errorValue any
	if strings.TrimSpace(errText) == "" {
		errorValue = nil
	} else {
		errorValue = errText
	}
	_, err := s.db.Exec(`UPDATE runs SET status=?, finished_at=?, error_text=? WHERE id=?`, status, time.Now().UTC(), errorValue, runID)
	if err != nil {
		return fmt.Errorf("finish run: %w", err)
	}
	return nil
}

func (s *Storage) SaveArticles(sourceName string, items []fetcher.NewsItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin save articles: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO articles (source_name, title, link, content, raw_content, published_at, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_name, link) DO UPDATE SET
			title=excluded.title,
			content=excluded.content,
			raw_content=excluded.raw_content,
			published_at=excluded.published_at,
			image_url=excluded.image_url,
			fetched_at=CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("prepare save articles: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.Exec(sourceName, item.Title, item.Link, item.Content, item.RawContent, item.PublishedOn.UTC(), item.ImageURL); err != nil {
			return fmt.Errorf("insert article %q: %w", item.Title, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit save articles: %w", err)
	}
	return nil
}

func (s *Storage) SavePost(post PostRecord, items []fetcher.NewsItem) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin save post: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO posts (source_name, target_channel_id, message_text, image_url, sent_at) VALUES (?, ?, ?, ?, ?)`,
		post.SourceName,
		post.TargetChannelID,
		post.MessageText,
		post.ImageURL,
		post.SentAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert post: %w", err)
	}
	postID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("post last insert id: %w", err)
	}

	for _, item := range items {
		var articleID int64
		err := tx.QueryRow(`SELECT id FROM articles WHERE source_name=? AND link=?`, post.SourceName, item.Link).Scan(&articleID)
		if err != nil {
			return fmt.Errorf("find article for link %q: %w", item.Link, err)
		}
		if _, err := tx.Exec(`INSERT OR IGNORE INTO post_articles (post_id, article_id) VALUES (?, ?)`, postID, articleID); err != nil {
			return fmt.Errorf("link post to article %q: %w", item.Link, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit save post: %w", err)
	}
	return nil
}

func (s *Storage) HasSentRecently(sourceName, messageText string, within time.Duration) (bool, error) {
	if within <= 0 {
		return false, nil
	}
	var count int
	threshold := time.Now().Add(-within).UTC()
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM posts WHERE source_name=? AND message_text=? AND sent_at>=?`,
		sourceName,
		strings.TrimSpace(messageText),
		threshold,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check recent post: %w", err)
	}
	return count > 0, nil
}
