package store

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func statDir(p string) (fs.FileInfo, error) { return os.Stat(p) }

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeProviderFile(t *testing.T, dir, name, content string) {
	t.Helper()
	mustMkdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
