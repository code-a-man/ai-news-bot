package state

import (
	"encoding/json"
	"fmt"
	"os"
)

type State struct {
	LastID        string `json:"last_id"`
	LastTimestamp string `json:"last_timestamp"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("state load: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("state unmarshal: %w", err)
	}
	return &s, nil
}

func Save(path string, id, timestamp string) error {
	s := State{LastID: id, LastTimestamp: timestamp}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("state marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func (s *State) HasChanged(id, timestamp string) bool {
	return s.LastID != id || s.LastTimestamp != timestamp
}
