// Command ccsw switches API providers for AI CLIs (Claude Code, Codex, ...).
//
// It stores provider configs under ~/.ccsw/providers/<app>/<name>/ and, on
// switch, atomically copies them over the CLI's native config files
// (e.g. ~/.claude/settings.json, ~/.codex/config.toml).
package main

import (
	"os"

	"github.com/cc-api-switcher-cli/ccsw/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
