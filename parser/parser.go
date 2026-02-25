package parser

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var trackerParams = map[string]bool{
	"utm_source": true, "utm_medium": true, "utm_campaign": true,
	"utm_term": true, "utm_content": true, "lid": true,
}

type NewsItem struct {
	Category string
	Title    string
	URL      string
}

var readTimeRe = regexp.MustCompile(`Read time:\s*\d+\s*min\s*\d+\s*sec`)

func ParseSummary(html string) ([]NewsItem, string, error) {
	summaryEnd := strings.Index(html, "Today's Author")
	if summaryEnd < 0 {
		summaryEnd = len(html)
	}
	summaryHTML := html[:summaryEnd]

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(summaryHTML))
	if err != nil {
		return nil, "", err
	}

	var items []NewsItem
	seen := make(map[string]bool)
	var readTime string
	currentCategory := ""

	summaryTable := doc.Find("td").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return strings.TrimSpace(s.Text()) == "Summary"
	}).ParentsFiltered("table").First()

	var searchArea *goquery.Selection
	if summaryTable.Length() > 0 {
		searchArea = summaryTable.Find("td")
	} else {
		searchArea = doc.Find("td")
	}

	searchArea.Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		if readTimeRe.MatchString(text) {
			readTime = strings.TrimSpace(readTimeRe.FindString(text))
		}

		s.Find("p u, u").Each(func(_ int, u *goquery.Selection) {
			cat := strings.TrimSpace(u.Text())
			if cat != "" && isSummaryCategory(cat) {
				currentCategory = cat
			}
		})

		s.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
			href, ok := a.Attr("href")
			if !ok || !isSummaryLink(href) {
				return
			}
			if seen[href] {
				return
			}
			seen[href] = true
			title := strings.TrimSpace(a.Text())
			if title == "" || !strings.Contains(title, "▸") {
				return
			}
			items = append(items, NewsItem{
				Category: currentCategory,
				Title:    title,
				URL:      href,
			})
		})
	})

	return items, readTime, nil
}

func isSummaryCategory(s string) bool {
	switch s {
	case "Top News", "Top Paper", "Signals":
		return true
	}
	return false
}

func isSummaryLink(href string) bool {
	if !strings.HasPrefix(href, "http") {
		return false
	}
	if strings.Contains(href, "typeform.com") ||
		strings.Contains(href, "unsubscribe") ||
		strings.Contains(href, "x.com") ||
		strings.Contains(href, "alphasignal.ai/?utm_source=email") {
		return false
	}
	return strings.Contains(href, "utm_source=alphasignal") ||
		strings.Contains(href, "utm_campaign=")
}

func CleanURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	for k := range q {
		if trackerParams[strings.ToLower(k)] {
			q.Del(k)
		}
	}
	u.RawQuery = q.Encode()
	return strings.TrimSuffix(u.String(), "?")
}

func escapeMarkdownLink(s string) string {
	return strings.NewReplacer(`\`, `\\`, `]`, `\]`).Replace(s)
}

func FormatMessage(items []NewsItem, readTime, subject string) string {
	var sb strings.Builder
	if subject != "" {
		sb.WriteString("*")
		sb.WriteString(escapeMarkdownLink(subject))
		sb.WriteString("*\n\n")
	}
	lastCategory := ""
	for _, item := range items {
		if item.Category == "AgentField" {
			continue
		}
		if item.Category != "" && item.Category != lastCategory {
			if lastCategory != "" {
				sb.WriteString("\n")
			}
			sb.WriteString("*")
			sb.WriteString(escapeMarkdownLink(item.Category))
			sb.WriteString("*\n")
			lastCategory = item.Category
		}
		cleanURL := CleanURL(item.URL)
		title := strings.TrimSpace(strings.TrimPrefix(item.Title, "▸"))
		title = strings.TrimSpace(title)
		sb.WriteString("• [")
		sb.WriteString(escapeMarkdownLink(title))
		sb.WriteString("](")
		sb.WriteString(cleanURL)
		sb.WriteString(")\n")
	}
	return strings.TrimSpace(sb.String())
}
