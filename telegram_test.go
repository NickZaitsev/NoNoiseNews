package main

import (
	"html"
	"strings"
	"testing"
	"time"
)

func TestPrepareTelegramHTMLTextEscapesAndTruncates(t *testing.T) {
	got := prepareTelegramHTMLText(`Hello <b>world</b> & "quotes"`, 100)
	want := `Hello &lt;b&gt;world&lt;/b&gt; &amp; &#34;quotes&#34;`
	if got != want {
		t.Fatalf("escape mismatch:\n got: %q\nwant: %q", got, want)
	}

	long := strings.Repeat("я", 20)
	truncated := prepareTelegramHTMLText(long, 10)
	if len([]rune(truncated)) != 10 {
		t.Fatalf("expected 10 runes after truncate, got %d (%q)", len([]rune(truncated)), truncated)
	}
	if !strings.HasSuffix(truncated, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", truncated)
	}

	entityHeavy := prepareTelegramHTMLText(strings.Repeat("<", 10), 4)
	unescaped := html.UnescapeString(entityHeavy)
	if unescaped != "<..." {
		t.Fatalf("expected complete escaped entities after truncation, got %q", entityHeavy)
	}
	if len([]rune(unescaped)) != 4 {
		t.Fatalf("expected Telegram-visible length of 4, got %d", len([]rune(unescaped)))
	}
}

func TestNewTelegramServiceUsesTimeoutAndDefaults(t *testing.T) {
	svc := NewTelegramService("key", nil, 5*time.Second, 100)
	if svc.httpClient == nil {
		t.Fatal("expected http client")
	}
	if svc.httpClient.Timeout != 5*time.Second {
		t.Fatalf("expected 5s timeout, got %s", svc.httpClient.Timeout)
	}
	if svc.maxMsgLen != 100 {
		t.Fatalf("expected maxMsgLen 100, got %d", svc.maxMsgLen)
	}

	defaults := NewTelegramService("key", nil, 0, 0)
	if defaults.httpClient.Timeout != DefaultAPITimeout {
		t.Fatalf("expected default timeout %s, got %s", DefaultAPITimeout, defaults.httpClient.Timeout)
	}
	if defaults.maxMsgLen != MaxMessageLength {
		t.Fatalf("expected default maxMsgLen %d, got %d", MaxMessageLength, defaults.maxMsgLen)
	}
}
