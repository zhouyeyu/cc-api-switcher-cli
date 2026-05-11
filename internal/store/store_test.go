package store

import (
	"path/filepath"
	"testing"

	"github.com/zhouyeyu/cc-api-switcher-cli/internal/app"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return &Store{Root: filepath.Join(t.TempDir(), ".ccsw")}
}

func TestInitCreatesSkeletonAndEmptyState(t *testing.T) {
	s := newStore(t)
	if err := s.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, a := range app.All() {
		dir := s.ProvidersDir(a.ID)
		if fi, err := statDir(dir); err != nil || !fi.IsDir() {
			t.Fatalf("providers dir for %s not created: %v", a.ID, err)
		}
	}
	st, err := s.LoadState()
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	if len(st.Current) != 0 {
		t.Fatalf("expected empty Current, got %v", st.Current)
	}
}

func TestInitIdempotent(t *testing.T) {
	s := newStore(t)
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	// Pre-populate state so we can assert Init does not clobber it.
	if err := s.SaveState(&State{Current: map[string]string{"claude": "kimi"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}
	st, _ := s.LoadState()
	if st.Current["claude"] != "kimi" {
		t.Fatalf("Init clobbered existing state: %v", st.Current)
	}
}

func TestSaveLoadState(t *testing.T) {
	s := newStore(t)
	if err := s.SaveState(&State{Current: map[string]string{"claude": "kimi", "codex": "openai"}}); err != nil {
		t.Fatal(err)
	}
	st, err := s.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if st.Current["claude"] != "kimi" || st.Current["codex"] != "openai" {
		t.Fatalf("round-trip mismatch: %v", st.Current)
	}
}

func TestLoadStateMissing(t *testing.T) {
	s := newStore(t)
	st, err := s.LoadState()
	if err != nil {
		t.Fatalf("LoadState on missing file: %v", err)
	}
	if st.Current == nil || len(st.Current) != 0 {
		t.Fatalf("expected empty Current map, got %v", st.Current)
	}
}

func TestListAndProviderExists(t *testing.T) {
	s := newStore(t)
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	claude, _ := app.Get("claude")

	// Empty list.
	names, err := s.ListProviders(claude.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("expected 0 providers, got %v", names)
	}
	if s.ProviderExists(claude, "kimi") {
		t.Fatal("ProviderExists should be false for missing provider")
	}

	// Create two providers, one missing its required file.
	writeProviderFile(t, s.ProviderDir(claude.ID, "kimi"), "settings.json", "{}")
	mustMkdir(t, s.ProviderDir(claude.ID, "empty"))

	names, err = s.ListProviders(claude.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Sorted: empty, kimi.
	if len(names) != 2 || names[0] != "empty" || names[1] != "kimi" {
		t.Fatalf("unexpected list: %v", names)
	}
	if !s.ProviderExists(claude, "kimi") {
		t.Fatal("kimi should exist")
	}
	if s.ProviderExists(claude, "empty") {
		t.Fatal("empty provider dir should fail ProviderExists (missing settings.json)")
	}
}
