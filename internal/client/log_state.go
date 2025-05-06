package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type LogState struct {
	LastPosition int64  `json:"last_position"`
	LastFile     string `json:"last_file"`
	mu           sync.Mutex
	stateFile    string
}

func NewLogState(stateFile string) (*LogState, error) {
	state := &LogState{
		stateFile: stateFile,
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	// Load existing state if available
	if err := state.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return state, nil
}

func (s *LogState) load() error {
	data, err := os.ReadFile(s.stateFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s)
}

func (s *LogState) save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(s.stateFile, data, 0644)
}

func (s *LogState) UpdatePosition(file string, position int64) error {
	s.mu.Lock()
	s.LastPosition = position
	s.LastFile = file
	s.mu.Unlock()

	return s.save()
}

func (s *LogState) GetLastPosition() (string, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.LastFile, s.LastPosition
}
