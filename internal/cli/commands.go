package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cc-api-switcher-cli/ccsw/internal/app"
	"github.com/cc-api-switcher-cli/ccsw/internal/fsutil"
	"github.com/cc-api-switcher-cli/ccsw/internal/store"
)

func runInit(_ []string) int {
	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	if err := s.Init(); err != nil {
		return errf("init store: %v", err)
	}
	fmt.Printf("initialized %s\n", s.Root)
	for _, a := range app.All() {
		fmt.Printf("  providers/%s/\n", a.ID)
	}
	return 0
}

func runList(args []string) int {
	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	st, err := s.LoadState()
	if err != nil {
		return errf("%v", err)
	}

	var apps []*app.App
	if len(args) == 0 {
		apps = app.All()
	} else {
		a, err := app.Get(args[0])
		if err != nil {
			return errf("%v", err)
		}
		apps = []*app.App{a}
	}

	for i, a := range apps {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s (%s):\n", a.DisplayName, a.ID)
		names, err := s.ListProviders(a.ID)
		if err != nil {
			return errf("list %s: %v", a.ID, err)
		}
		if len(names) == 0 {
			fmt.Println("  (none)")
			continue
		}
		active := st.Current[a.ID]
		for _, n := range names {
			marker := "  "
			if n == active {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, n)
		}
	}
	return 0
}

func runCurrent(_ []string) int {
	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	st, err := s.LoadState()
	if err != nil {
		return errf("%v", err)
	}
	for _, a := range app.All() {
		cur := st.Current[a.ID]
		if cur == "" {
			cur = "(none)"
		}
		fmt.Printf("%-7s %s\n", a.ID, cur)
	}
	return 0
}

func runUse(args []string) int {
	if len(args) != 2 {
		return errf("usage: ccsw use <app> <name>")
	}
	a, err := app.Get(args[0])
	if err != nil {
		return errf("%v", err)
	}
	name := args[1]

	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	if !s.ProviderExists(a, name) {
		return errf("provider %q not found or incomplete under %s", name, s.ProviderDir(a.ID, name))
	}

	// Back up every target file before touching anything. Remember which ones
	// actually had a pre-existing file so we can restore them on rollback.
	backedUp := make([]string, 0, len(a.Files))
	for _, dst := range a.Files {
		existed := fsutil.Exists(dst)
		if err := fsutil.BackupIfExists(dst); err != nil {
			return errf("backup %s: %v", dst, err)
		}
		if existed {
			backedUp = append(backedUp, dst)
		}
	}

	// Atomically copy each provider file to its target. If any write fails
	// mid-way, restore every already-touched target from its .bak so the user
	// is not left with a half-switched (e.g. config.toml from the new provider,
	// auth.json from the old one) state.
	providerDir := s.ProviderDir(a.ID, name)
	written := make([]string, 0, len(a.Files))
	for fname, dst := range a.Files {
		src := filepath.Join(providerDir, fname)
		if err := fsutil.AtomicCopy(src, dst); err != nil {
			rollback(written, backedUp)
			return errf("write %s: %v", dst, err)
		}
		written = append(written, dst)
	}

	// Update state.
	st, err := s.LoadState()
	if err != nil {
		return errf("%v", err)
	}
	st.Current[a.ID] = name
	if err := s.SaveState(st); err != nil {
		return errf("save state: %v", err)
	}

	fmt.Printf("switched %s -> %s\n", a.ID, name)
	for _, dst := range a.Files {
		fmt.Printf("  wrote %s\n", dst)
	}
	return 0
}

func runAdd(args []string) int {
	if len(args) != 3 {
		return errf("usage: ccsw add <app> <name> <src-dir>")
	}
	a, err := app.Get(args[0])
	if err != nil {
		return errf("%v", err)
	}
	name, srcDir := args[1], args[2]

	// Validate that srcDir contains all expected files.
	missing := []string{}
	for fname := range a.Files {
		if !fsutil.Exists(filepath.Join(srcDir, fname)) {
			missing = append(missing, fname)
		}
	}
	if len(missing) > 0 {
		return errf("%s is missing required file(s): %v", srcDir, missing)
	}

	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	dstDir := s.ProviderDir(a.ID, name)
	if fsutil.Exists(dstDir) {
		return errf("provider %q already exists at %s (remove it first)", name, dstDir)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return errf("%v", err)
	}
	for fname := range a.Files {
		src := filepath.Join(srcDir, fname)
		dst := filepath.Join(dstDir, fname)
		if err := fsutil.AtomicCopy(src, dst); err != nil {
			return errf("copy %s: %v", fname, err)
		}
	}
	fmt.Printf("added %s provider %q at %s\n", a.ID, name, dstDir)
	return 0
}

func runRemove(args []string) int {
	if len(args) != 2 {
		return errf("usage: ccsw rm <app> <name>")
	}
	a, err := app.Get(args[0])
	if err != nil {
		return errf("%v", err)
	}
	name := args[1]

	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	dir := s.ProviderDir(a.ID, name)
	if !fsutil.Exists(dir) {
		return errf("provider %q not found", name)
	}
	if err := os.RemoveAll(dir); err != nil {
		return errf("%v", err)
	}

	// If it was the active provider, clear the state entry.
	if st, err := s.LoadState(); err == nil {
		if st.Current[a.ID] == name {
			delete(st.Current, a.ID)
			_ = s.SaveState(st)
		}
	}
	fmt.Printf("removed %s provider %q\n", a.ID, name)
	return 0
}

func runEdit(args []string) int {
	if len(args) != 2 {
		return errf("usage: ccsw edit <app> <name>")
	}
	a, err := app.Get(args[0])
	if err != nil {
		return errf("%v", err)
	}
	name := args[1]

	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	dir := s.ProviderDir(a.ID, name)
	if !fsutil.Exists(dir) {
		return errf("provider %q not found", name)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, dir)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return errf("editor: %v", err)
	}
	return 0
}

// rollback best-effort undoes a partial `use`: it removes any newly written
// target files and renames each <file>.bak back to <file> for targets that
// had a pre-existing config. Errors during rollback are logged but not
// propagated; the caller already has a primary error to report.
func rollback(written, backedUp []string) {
	for _, p := range written {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "ccsw: rollback: remove %s: %v\n", p, err)
		}
	}
	for _, p := range backedUp {
		bak := p + ".bak"
		if err := os.Rename(bak, p); err != nil {
			fmt.Fprintf(os.Stderr, "ccsw: rollback: restore %s from %s: %v\n", p, bak, err)
		}
	}
}

func runPath(args []string) int {
	if len(args) != 2 {
		return errf("usage: ccsw path <app> <name>")
	}
	a, err := app.Get(args[0])
	if err != nil {
		return errf("%v", err)
	}
	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	fmt.Println(s.ProviderDir(a.ID, args[1]))
	return 0
}

// runNew runs the interactive wizard to create a new provider from scratch.
//
// Usage:
//
//	ccsw new <app> <name> [--template <tpl>]
//
// For claude, --template <tpl> uses a built-in preset (only asks for the API
// key). Without --template, the user pastes the official config block.
// For codex, Q&A wizard generates config.toml + auth.json.
func runNew(args []string) int {
	// Parse --template / -t flag manually (no framework dependency).
	var templateID string
	filtered := args[:0]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--template", "-t":
			if i+1 >= len(args) {
				return errf("--template requires an argument")
			}
			i++
			templateID = args[i]
		default:
			filtered = append(filtered, args[i])
		}
	}
	args = filtered

	if len(args) != 2 {
		return errf("usage: ccsw new <app> <name> [--template <tpl>]")
	}
	a, err := app.Get(args[0])
	if err != nil {
		return errf("%v", err)
	}
	name := args[1]

	s, err := store.Default()
	if err != nil {
		return errf("%v", err)
	}
	dstDir := s.ProviderDir(a.ID, name)
	if fsutil.Exists(dstDir) {
		return errf("provider %q already exists at %s (remove it first)", name, dstDir)
	}

	stdin := bufio.NewReader(os.Stdin)
	var files map[string][]byte
	switch a.ID {
	case "claude":
		if templateID != "" {
			tpl, ok := app.ClaudeTemplateByID(templateID)
			if !ok {
				return errf("unknown template %q — run `ccsw templates` to see available templates", templateID)
			}
			files, err = wizardClaudeTemplate(stdin, tpl)
		} else {
			files, err = wizardClaude(stdin)
		}
	case "codex":
		files, err = wizardCodex(stdin)
	default:
		return errf("no wizard for app %q — use `ccsw add` instead", a.ID)
	}
	if err != nil {
		return errf("%v", err)
	}

	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return errf("%v", err)
	}
	for fname, data := range files {
		target := filepath.Join(dstDir, fname)
		if err := os.WriteFile(target, data, 0o600); err != nil {
			return errf("write %s: %v", target, err)
		}
	}

	fmt.Printf("\ncreated %s provider %q at %s\n", a.ID, name, dstDir)
	fmt.Printf("switch with: ccsw use %s %s\n", a.ID, name)
	return 0
}

// runTemplates lists built-in provider templates.
func runTemplates(_ []string) int {
	fmt.Println("Built-in Claude templates (use with: ccsw new claude <name> --template <id>):")
	fmt.Println()
	for _, t := range app.AllClaudeTemplates() {
		fmt.Printf("  %-12s  %s\n", t.ID, t.DisplayName)
		if t.WebsiteURL != "" {
			fmt.Printf("               %s\n", t.WebsiteURL)
		}
	}
	return 0
}

