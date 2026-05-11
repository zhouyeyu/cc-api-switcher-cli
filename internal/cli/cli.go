// Package cli wires the command-line interface.
//
// Commands are dispatched from a single subcommand argument (os.Args[1]).
// Each command is implemented as a small function that parses its own flags
// so we avoid adding a CLI framework dependency.
package cli

import (
	"fmt"
	"os"
)

// Version is the released version string. It is overridden at build time via
//
//	go build -ldflags "-X github.com/zhouyeyu/cc-api-switcher-cli/internal/cli.Version=v0.1.0"
var Version = "dev"

const usage = `ccsw - switch API providers for AI CLIs (Claude Code, Codex)

Usage:
  ccsw <command> [args...]

Commands:
  init                               Create ~/.ccsw directory skeleton
  ls [app]                           List providers (all apps, or one app)
  current                            Show the currently active provider per app
  use <app> <name>                   Switch <app> to provider <name>
  add <app> <name> <src-dir>         Import provider files from <src-dir>
  new <app> <name> [--template <t>]  Create provider via wizard (paste or template)
  templates                          List built-in provider templates
  rm  <app> <name>                   Delete a stored provider
  edit <app> <name>                  Open the provider directory in $EDITOR
  path <app> <name>                  Print the provider directory path
  version                            Print ccsw version

Supported apps: claude, codex

Examples:
  ccsw new claude deepseek --template deepseek   # only enter your API key
  ccsw new claude myprovider                     # paste any config block
  ccsw templates                                 # list built-in templates
`

// Run is the entry point. Returns the process exit code.
func Run(args []string) int {
	if len(args) < 1 {
		fmt.Print(usage)
		return 2
	}
	cmd, rest := args[0], args[1:]

	switch cmd {
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	case "-v", "--version", "version":
		fmt.Println(Version)
		return 0
	case "init":
		return runInit(rest)
	case "ls", "list":
		return runList(rest)
	case "current":
		return runCurrent(rest)
	case "use":
		return runUse(rest)
	case "add":
		return runAdd(rest)
	case "new":
		return runNew(rest)
	case "templates":
		return runTemplates(rest)
	case "rm", "remove":
		return runRemove(rest)
	case "edit":
		return runEdit(rest)
	case "path":
		return runPath(rest)
	default:
		fmt.Fprintf(os.Stderr, "ccsw: unknown command %q\n\n", cmd)
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
}

func errf(format string, a ...any) int {
	fmt.Fprintf(os.Stderr, "ccsw: "+format+"\n", a...)
	return 1
}
