package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

type ChatMessage struct {
	MessageID    int    `json:"message_id"`
	LastDescHash string `json:"last_desc_hash"`
}

type RSSIncidentState struct {
	Messages map[string]ChatMessage `json:"messages"`
}

type RSSState struct {
	Incidents map[string]RSSIncidentState `json:"incidents"`
}

func LoadRSS(path string) (*RSSState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &RSSState{Incidents: make(map[string]RSSIncidentState)}, nil
		}
		return nil, fmt.Errorf("rss state load: %w", err)
	}

	var s RSSState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("rss state unmarshal: %w", err)
	}
	if s.Incidents == nil {
		s.Incidents = make(map[string]RSSIncidentState)
	}
	return &s, nil
}

func SaveRSS(path string, s *RSSState) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("rss state marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (s *RSSState) GetMessage(guid, chatID string) (messageID int, lastHash string, ok bool) {
	inc, ok := s.Incidents[guid]
	if !ok {
		return 0, "", false
	}
	msg, ok := inc.Messages[chatID]
	if !ok {
		return 0, "", false
	}
	return msg.MessageID, msg.LastDescHash, true
}

func (s *RSSState) SetMessage(guid, chatID string, messageID int, descHash string) {
	if s.Incidents == nil {
		s.Incidents = make(map[string]RSSIncidentState)
	}
	inc := s.Incidents[guid]
	if inc.Messages == nil {
		inc.Messages = make(map[string]ChatMessage)
	}
	inc.Messages[chatID] = ChatMessage{MessageID: messageID, LastDescHash: descHash}
	s.Incidents[guid] = inc
}

func HashDescription(desc string) string {
	h := sha256.Sum256([]byte(desc))
	return hex.EncodeToString(h[:])
}
