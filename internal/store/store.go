// Package store owns the on-disk layout under ~/.ccsw.
//
// Layout:
//
//	~/.ccsw/
//	├── state.json                       # {"current": {"claude": "kimi", "codex": "openai"}}
//	└── providers/
//	    ├── claude/
//	    │   ├── official/settings.json
//	    │   └── kimi/settings.json
//	    └── codex/
//	        └── openai/
//	            ├── config.toml
//	            └── auth.json
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/zhouyeyu/cc-api-switcher-cli/internal/app"
	"github.com/zhouyeyu/cc-api-switcher-cli/internal/fsutil"
)

// Store is rooted at ~/.ccsw.
type Store struct {
	Root string
}

// Default returns a Store rooted at $HOME/.ccsw.
func Default() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}
	return &Store{Root: filepath.Join(home, ".ccsw")}, nil
}

// ProvidersDir returns the directory holding providers for the given app.
func (s *Store) ProvidersDir(appID string) string {
	return filepath.Join(s.Root, "providers", appID)
}

// ProviderDir returns the directory holding a specific provider's files.
func (s *Store) ProviderDir(appID, name string) string {
	return filepath.Join(s.ProvidersDir(appID), name)
}

// StateFile returns the path to state.json.
func (s *Store) StateFile() string {
	return filepath.Join(s.Root, "state.json")
}

// Init creates the directory skeleton. Safe to call repeatedly.
func (s *Store) Init() error {
	for _, a := range app.All() {
		if err := os.MkdirAll(s.ProvidersDir(a.ID), 0o755); err != nil {
			return err
		}
	}
	if !fsutil.Exists(s.StateFile()) {
		return s.SaveState(&State{Current: map[string]string{}})
	}
	return nil
}

// ListProviders returns the sorted list of provider names for the given app.
func (s *Store) ListProviders(appID string) ([]string, error) {
	dir := s.ProvidersDir(appID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// ProviderExists checks whether the provider directory exists and contains at
// least one expected file for the app.
func (s *Store) ProviderExists(a *app.App, name string) bool {
	dir := s.ProviderDir(a.ID, name)
	if !fsutil.Exists(dir) {
		return false
	}
	for fname := range a.Files {
		if !fsutil.Exists(filepath.Join(dir, fname)) {
			return false
		}
	}
	return true
}

// State represents ~/.ccsw/state.json.
type State struct {
	Current map[string]string `json:"current"`
}

// LoadState reads state.json or returns an empty state if missing.
func (s *Store) LoadState() (*State, error) {
	data, err := os.ReadFile(s.StateFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Current: map[string]string{}}, nil
		}
		return nil, err
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parse state.json: %w", err)
	}
	if st.Current == nil {
		st.Current = map[string]string{}
	}
	return &st, nil
}

// SaveState persists state to state.json atomically.
func (s *Store) SaveState(st *State) error {
	if err := os.MkdirAll(s.Root, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Root, ".state-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.StateFile())
}
