package parser

import (
	"strings"
	"testing"
)

func TestParseSummary(t *testing.T) {
	html := `
	<table>
	<tr><td>Summary</td></tr>
	<tr><td>Read time: 4 min 31 sec</td></tr>
	<tr><td><p><u>Top News</u></p></td></tr>
	<tr><td><a href="https://example.com/1?utm_source=alphasignal">▸ First headline</a></td></tr>
	<tr><td><p><u>Signals</u></p></td></tr>
	<tr><td><a href="https://example.com/2?utm_campaign=2026">▸ Second headline</a></td></tr>
	</table>
	`
	items, readTime, err := ParseSummary(html)
	if err != nil {
		t.Fatal(err)
	}
	if readTime != "Read time: 4 min 31 sec" {
		t.Errorf("readTime = %q", readTime)
	}
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
	if items[0].Title != "▸ First headline" || items[0].Category != "Top News" {
		t.Errorf("item0 = %+v", items[0])
	}
	if items[1].Title != "▸ Second headline" || items[1].Category != "Signals" {
		t.Errorf("item1 = %+v", items[1])
	}
}

func TestFormatMessage(t *testing.T) {
	items := []NewsItem{
		{Category: "Top News", Title: "▸ Headline 1", URL: "https://a.com/1?utm_source=test"},
		{Category: "Top News", Title: "▸ Headline 2", URL: "https://a.com/2"},
		{Category: "Signals", Title: "▸ Headline 3", URL: "https://a.com/3"},
	}
	msg := FormatMessage(items, "Read time: 4 min 31 sec", "Test Subject")
	if !strings.Contains(msg, "Top News") || !strings.Contains(msg, "Signals") {
		t.Errorf("unexpected format: %s", msg)
	}
	if !strings.Contains(msg, "Test Subject") {
		t.Errorf("missing subject: %s", msg)
	}
	if strings.Contains(msg, "utm_source") {
		t.Errorf("tracker params not removed: %s", msg)
	}
	if !strings.Contains(msg, "https") || !strings.Contains(msg, "com") {
		t.Errorf("URL missing: %s", msg)
	}
	if !strings.Contains(msg, "[") || !strings.Contains(msg, "](") {
		t.Errorf("missing markdown link format: %s", msg)
	}
}

func TestCleanURL(t *testing.T) {
	u := CleanURL("https://example.com/path?utm_source=foo&utm_campaign=bar&lid=123&real=keep")
	if strings.Contains(u, "utm_source") || strings.Contains(u, "utm_campaign") || strings.Contains(u, "lid") {
		t.Errorf("tracker params not removed: %s", u)
	}
	if !strings.Contains(u, "real=keep") {
		t.Errorf("real param removed: %s", u)
	}
}
