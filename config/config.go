package config

import (
	"os"
	"strings"
)

const (
	DefaultAPIURL        = "https://alphasignal.ai/api/last-campaign"
	DefaultStateFile    = "./state.json"
	DefaultClaudeRSSURL = "https://status.claude.com/history.rss"
	DefaultStateRSSFile = "./state_rss.json"
)

type Config struct {
	TelegramToken   string
	ChatIDs         []string
	StateFile       string
	APIURL          string
	ClaudeRSSURL    string
	StateRSSFile    string
}

func Load() *Config {
	chatIDs := os.Getenv("TELEGRAM_CHAT_IDS")
	var chatIDList []string
	if chatIDs != "" {
		for _, id := range strings.Split(chatIDs, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				chatIDList = append(chatIDList, trimmed)
			}
		}
	}

	stateFile := os.Getenv("STATE_FILE")
	if stateFile == "" {
		stateFile = DefaultStateFile
	}

	apiURL := os.Getenv("ALPHASIGNAL_API")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	claudeRSS := os.Getenv("CLAUDE_STATUS_RSS_URL")
	if claudeRSS == "" {
		claudeRSS = DefaultClaudeRSSURL
	}

	stateRSS := os.Getenv("STATE_RSS_FILE")
	if stateRSS == "" {
		stateRSS = DefaultStateRSSFile
	}

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatIDs:       chatIDList,
		StateFile:     stateFile,
		APIURL:        apiURL,
		ClaudeRSSURL:  claudeRSS,
		StateRSSFile:  stateRSS,
	}
}
