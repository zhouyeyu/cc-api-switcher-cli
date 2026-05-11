// Package fsutil provides small file-system helpers used across ccsw.
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AtomicCopy copies src to dst using a temp-file + rename so that readers
// never observe a partially written dst. The target directory is created
// if it does not yet exist.
func AtomicCopy(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer in.Close()

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".ccsw-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup on failure paths.
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := io.Copy(tmp, in); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// BackupIfExists renames dst to dst+".bak" (overwriting any previous .bak).
// If dst does not exist, it is a no-op.
func BackupIfExists(dst string) error {
	if _, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	bak := dst + ".bak"
	_ = os.Remove(bak)
	return os.Rename(dst, bak)
}

// Exists reports whether the given path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
