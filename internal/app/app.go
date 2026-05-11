package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// App represents a target AI CLI application whose config files ccsw manages.
type App struct {
	// ID is the short identifier used on the command line, e.g. "claude", "codex".
	ID string
	// DisplayName is shown in human-facing output.
	DisplayName string
	// Files maps a logical filename (as stored inside a provider directory)
	// to its absolute target path on disk.
	// All files listed here will be overwritten on switch.
	Files map[string]string
}

// Registry holds all supported apps.
var Registry = map[string]*App{}

func register(a *App) {
	Registry[a.ID] = a
}

// All returns the registered apps in deterministic order.
func All() []*App {
	order := []string{"claude", "codex"}
	out := make([]*App, 0, len(order))
	for _, id := range order {
		if a, ok := Registry[id]; ok {
			out = append(out, a)
		}
	}
	return out
}

// Get returns the app by id or an error if unknown.
func Get(id string) (*App, error) {
	a, ok := Registry[id]
	if !ok {
		return nil, fmt.Errorf("unknown app %q (supported: claude, codex)", id)
	}
	return a, nil
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fall back to "~"; actual operations will fail with a clear error later.
		home = "~"
	}

	register(&App{
		ID:          "claude",
		DisplayName: "Claude Code",
		Files: map[string]string{
			"settings.json": filepath.Join(home, ".claude", "settings.json"),
		},
	})

	register(&App{
		ID:          "codex",
		DisplayName: "Codex",
		Files: map[string]string{
			"config.toml": filepath.Join(home, ".codex", "config.toml"),
			"auth.json":   filepath.Join(home, ".codex", "auth.json"),
		},
	})
}
