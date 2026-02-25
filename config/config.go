package config

import (
	"os"
	"strings"
)

const (
	DefaultAPIURL    = "https://alphasignal.ai/api/last-campaign"
	DefaultStateFile = "./state.json"
)

type Config struct {
	TelegramToken string
	ChatIDs       []string
	StateFile     string
	APIURL        string
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

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatIDs:       chatIDList,
		StateFile:     stateFile,
		APIURL:        apiURL,
	}
}
