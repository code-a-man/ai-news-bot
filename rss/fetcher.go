package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

type channel struct {
	Items []Item `xml:"item"`
}

type rssFeed struct {
	Channel channel `xml:"channel"`
}

func Fetch(url string) ([]Item, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("rss fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("rss fetch: status %d: %s", resp.StatusCode, string(body))
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("rss decode: %w", err)
	}

	return feed.Channel.Items, nil
}

func StripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag && r != '\n':
			b.WriteRune(r)
		case !inTag && r == '\n':
			b.WriteRune(' ')
		}
	}
	return strings.TrimSpace(strings.Join(strings.Fields(b.String()), " "))
}

func FormatMessage(item Item) string {
	title := item.Title
	link := item.Link
	if link == "" {
		link = item.GUID
	}
	title = escapeMD(title)
	updates := ExtractIncidentUpdates(item)
	latestStatus := ExtractLatestStatus(item.Description)

	var b strings.Builder
	b.WriteString("*Claude Status: ")
	b.WriteString(title)
	b.WriteString("*\n\n")
	if len(updates) > 0 {
		for i, u := range updates {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(escapeMD(u.Status))
			b.WriteString(" - ")
			b.WriteString(escapeMD(u.Message))
			b.WriteString("\n")
			b.WriteString("_")
			b.WriteString(escapeMD(u.LocalTime))
			b.WriteString("_")
		}
	} else {
		b.WriteString(escapeMD(latestStatus))
	}
	b.WriteString("\n\n[Details](")
	b.WriteString(link)
	b.WriteString(")")
	return b.String()
}

var statusRegex = regexp.MustCompile(`(?i)\b(Resolved|Monitoring|Identified|Update|Investigating)\b(?:\s*[-:])?`)
var htmlUpdateRegex = regexp.MustCompile(`(?is)<small>\s*([A-Za-z]{3})\s*<var[^>]*>\s*(\d{1,2})\s*</var>\s*,\s*<var[^>]*>\s*(\d{2}:\d{2})\s*</var>\s*UTC\s*</small>\s*<br\s*/?>\s*<strong>\s*(Resolved|Monitoring|Identified|Update|Investigating)\s*</strong>\s*-\s*(.*?)</p>`)
var plainUpdateRegex = regexp.MustCompile(`(?is)\b(Resolved|Monitoring|Identified|Update|Investigating)\b\s*-\s*(.*?)\s*((?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},\s+\d{4}\s*-\s*\d{2}:\d{2}\s+UTC)`)

type IncidentUpdate struct {
	Status    string
	Message   string
	UTCTime   string
	LocalTime string
}

func IncidentStatus(item Item) string {
	if updates := ExtractIncidentUpdates(item); len(updates) > 0 {
		return updates[0].Status
	}
	if status := extractStatus(item.Description); status != "" {
		return status
	}
	return extractStatus(item.Title)
}

func IsOpenIncident(item Item) bool {
	status := IncidentStatus(item)
	if status == "" {
		// If status cannot be inferred, treat as open to avoid missing incidents.
		return true
	}
	return status != "Resolved"
}

func extractStatus(text string) string {
	plain := StripHTML(text)
	if plain == "" {
		return ""
	}
	matches := statusRegex.FindStringSubmatch(plain)
	if len(matches) < 2 {
		return ""
	}
	switch strings.ToLower(matches[1]) {
	case "resolved":
		return "Resolved"
	case "monitoring":
		return "Monitoring"
	case "identified":
		return "Identified"
	case "update":
		return "Update"
	case "investigating":
		return "Investigating"
	default:
		return ""
	}
}

func escapeMD(s string) string {
	return strings.NewReplacer(`\`, `\\`, `]`, `\]`, `*`, `\*`, `_`, `\_`).Replace(s)
}

func ExtractIncidentUpdates(item Item) []IncidentUpdate {
	year := pubYear(item.PubDate)
	matches := htmlUpdateRegex.FindAllStringSubmatch(item.Description, -1)
	if len(matches) == 0 {
		return extractIncidentUpdatesFromPlain(item.Description)
	}

	updates := make([]IncidentUpdate, 0, len(matches))
	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		utcTime := buildUTCTimestamp(m[1], m[2], year, m[3])
		status := normalizeStatus(m[4])
		message := StripHTML(strings.TrimSpace(m[5]))
		localTime := toLocalTime(utcTime)
		updates = append(updates, IncidentUpdate{
			Status:    status,
			Message:   message,
			UTCTime:   utcTime,
			LocalTime: localTime,
		})
	}
	return updates
}

func ExtractLatestStatus(desc string) string {
	statuses := []string{"Resolved", "Monitoring", "Identified", "Update", "Investigating"}
	plain := StripHTML(desc)
	for _, s := range statuses {
		if idx := strings.Index(plain, s); idx >= 0 {
			end := strings.Index(plain[idx:], ".")
			if end < 0 {
				end = len(plain) - idx
			} else {
				end++
			}
			if end > 200 {
				end = 200
			}
			return strings.TrimSpace(plain[idx : idx+end])
		}
	}
	if len(plain) > 300 {
		return plain[:300] + "..."
	}
	return plain
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "resolved":
		return "Resolved"
	case "monitoring":
		return "Monitoring"
	case "identified":
		return "Identified"
	case "update":
		return "Update"
	case "investigating":
		return "Investigating"
	default:
		return strings.TrimSpace(status)
	}
}

func toLocalTime(utcTime string) string {
	t, err := time.Parse("Jan 2, 2006 - 15:04 MST", utcTime)
	if err != nil {
		return utcTime
	}
	return t.In(time.Local).Format("02 Jan 2006 - 15:04 MST")
}

func buildUTCTimestamp(month, day string, year int, hhmm string) string {
	normalizedMonth := strings.Title(strings.ToLower(strings.TrimSpace(month)))
	normalizedDay := strings.TrimSpace(day)
	normalizedHHMM := strings.TrimSpace(hhmm)
	raw := fmt.Sprintf("%s %s %d %s", normalizedMonth, normalizedDay, year, normalizedHHMM)
	t, err := time.Parse("Jan 2 2006 15:04", raw)
	if err != nil {
		return fmt.Sprintf("%s %s, %d - %s UTC", normalizedMonth, normalizedDay, year, normalizedHHMM)
	}
	return t.UTC().Format("Jan 2, 2006 - 15:04 MST")
}

func pubYear(pubDate string) int {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04 MST",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, strings.TrimSpace(pubDate)); err == nil {
			return t.Year()
		}
	}
	return time.Now().UTC().Year()
}

func extractIncidentUpdatesFromPlain(desc string) []IncidentUpdate {
	plain := StripHTML(desc)
	if plain == "" {
		return nil
	}

	matches := plainUpdateRegex.FindAllStringSubmatch(plain, -1)
	if len(matches) == 0 {
		return nil
	}

	updates := make([]IncidentUpdate, 0, len(matches))
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		status := normalizeStatus(m[1])
		message := strings.TrimSpace(m[2])
		utcTime := strings.Join(strings.Fields(strings.TrimSpace(m[3])), " ")
		localTime := toLocalTime(utcTime)
		updates = append(updates, IncidentUpdate{
			Status:    status,
			Message:   message,
			UTCTime:   utcTime,
			LocalTime: localTime,
		})
	}
	return updates
}
