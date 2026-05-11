package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAtomicCopyCreatesTarget(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "nested", "dst.txt")
	writeFile(t, src, "hello")

	if err := AtomicCopy(src, dst); err != nil {
		t.Fatalf("AtomicCopy: %v", err)
	}
	if got := readFile(t, dst); got != "hello" {
		t.Fatalf("dst contents = %q, want %q", got, "hello")
	}
	// No leftover temp files next to dst.
	entries, _ := os.ReadDir(filepath.Dir(dst))
	for _, e := range entries {
		if e.Name() != "dst.txt" {
			t.Fatalf("unexpected leftover file: %s", e.Name())
		}
	}
}

func TestAtomicCopyOverwrites(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	writeFile(t, src, "new")
	writeFile(t, dst, "old")

	if err := AtomicCopy(src, dst); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, dst); got != "new" {
		t.Fatalf("dst = %q, want %q", got, "new")
	}
}

func TestAtomicCopyMissingSource(t *testing.T) {
	dir := t.TempDir()
	err := AtomicCopy(filepath.Join(dir, "nope"), filepath.Join(dir, "dst"))
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

func TestBackupIfExists(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "f")
	writeFile(t, dst, "v1")

	if err := BackupIfExists(dst); err != nil {
		t.Fatal(err)
	}
	if Exists(dst) {
		t.Fatal("original should have been renamed away")
	}
	if got := readFile(t, dst+".bak"); got != "v1" {
		t.Fatalf(".bak = %q, want %q", got, "v1")
	}

	// Second round: .bak should be overwritten, only one generation kept.
	writeFile(t, dst, "v2")
	if err := BackupIfExists(dst); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, dst+".bak"); got != "v2" {
		t.Fatalf("second .bak = %q, want %q", got, "v2")
	}
}

func TestBackupIfExistsMissing(t *testing.T) {
	dir := t.TempDir()
	// No-op on missing file, no error.
	if err := BackupIfExists(filepath.Join(dir, "nope")); err != nil {
		t.Fatalf("BackupIfExists on missing: %v", err)
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if Exists(filepath.Join(dir, "nope")) {
		t.Fatal("Exists lied about missing file")
	}
	p := filepath.Join(dir, "x")
	writeFile(t, p, "")
	if !Exists(p) {
		t.Fatal("Exists lied about present file")
	}
}
