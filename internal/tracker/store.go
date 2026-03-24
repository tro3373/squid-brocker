package tracker

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sync"
)

// UsageKey identifies a unique device+group+date combination.
type UsageKey struct {
	Device string `json:"device"`
	Group  string `json:"group"`
	Date   string `json:"date"` // YYYY-MM-DD
}

// UsageValue holds the tracked data for a single usage key.
type UsageValue struct {
	Seconds int
	Label   string
}

// State holds all tracked usage data.
type State map[UsageKey]UsageValue

// Store persists usage state across restarts.
type Store interface {
	Load() (State, error)
	Save(State) error
}

// MemoryStore is an in-memory Store for testing.
type MemoryStore struct {
	mu    sync.Mutex
	state State
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{state: make(State)}
}

func (m *MemoryStore) Load() (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make(State, len(m.state))
	maps.Copy(cp, m.state)
	return cp, nil
}

func (m *MemoryStore) Save(s State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = make(State, len(s))
	maps.Copy(m.state, s)
	return nil
}

// FileStore persists state to a JSON file.
type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (f *FileStore) Load() (State, error) {
	data, err := os.ReadFile(f.path)
	if os.IsNotExist(err) {
		return make(State), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var entries []stateEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	state := make(State, len(entries))
	for _, e := range entries {
		state[UsageKey{Device: e.Device, Group: e.Group, Date: e.Date}] = UsageValue{
			Seconds: e.Seconds,
			Label:   e.Label,
		}
	}
	return state, nil
}

func (f *FileStore) Save(s State) error {
	entries := make([]stateEntry, 0, len(s))
	for k, v := range s {
		entries = append(entries, stateEntry{
			Device:  k.Device,
			Label:   v.Label,
			Group:   k.Group,
			Date:    k.Date,
			Seconds: v.Seconds,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}
	if err := os.WriteFile(f.path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return os.Chmod(f.path, 0644)
}

type stateEntry struct {
	Device  string `json:"device"`
	Label   string `json:"label,omitempty"`
	Group   string `json:"group"`
	Date    string `json:"date"`
	Seconds int    `json:"seconds"`
}
