package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Campaign struct {
	ID        string `json:"_id"`
	HTML      string `json:"html"`
	Subject   string `json:"subject"`
	Timestamp string `json:"timestamp"`
}

func Fetch(apiURL string) (*Campaign, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch: status %d: %s", resp.StatusCode, string(body))
	}

	var campaign Campaign
	if err := json.NewDecoder(resp.Body).Decode(&campaign); err != nil {
		return nil, fmt.Errorf("fetch decode: %w", err)
	}

	return &campaign, nil
}
