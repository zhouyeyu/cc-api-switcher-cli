package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/cc-api-switcher-cli/ccsw/internal/app"
)

// envLineRE matches `export KEY=VAL`, `KEY=VAL`, with optional whitespace.
// KEY is a standard POSIX-ish identifier (letters, digits, underscore,
// starting with a non-digit).
var envLineRE = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*?)\s*$`)

// jsonKVRE matches a JSON-style `"KEY": "VAL"` line (with optional trailing
// comma), as found in settings.json excerpts or provider docs that hand you
// the env block already in JSON format.
var jsonKVRE = regexp.MustCompile(`^\s*"([^"]+)"\s*:\s*"((?:[^"\\]|\\.)*)"\s*,?\s*$`)

// parseEnvBlock reads `export KEY=VAL` / `KEY=VAL` lines from r and returns
// the parsed map. Reading stops at the first blank line (treated as a
// terminator so an interactive caller can keep reading further input from the
// same stream afterwards) or at EOF, whichever comes first. Lines that don't
// match the pattern are returned via `skipped` so the caller can warn about
// them. Surrounding single or double quotes are stripped from the value.
// Later occurrences of the same key win.
//
// The reader is consumed line-by-line via ReadString so that any remaining
// bytes (e.g. subsequent prompt answers) stay available on the same reader.
func parseEnvBlock(r *bufio.Reader) (env map[string]string, skipped []string, err error) {
	env = map[string]string{}
	for {
		line, readErr := r.ReadString('\n')
		// Trim trailing newline before inspecting content.
		trimmed := strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(trimmed) == "" {
			if readErr == io.EOF || len(line) == 0 {
				return env, skipped, nil
			}
			// Blank line terminates the block.
			return env, skipped, nil
		}
		if m := envLineRE.FindStringSubmatch(trimmed); m != nil {
			env[m[1]] = stripQuotes(m[2])
		} else if m := jsonKVRE.FindStringSubmatch(trimmed); m != nil {
			// JSON escape sequences in the value (e.g. \n, \") are decoded so
			// the stored string is the actual intended value.
			val := m[2]
			var decoded string
			if err := json.Unmarshal([]byte(`"`+val+`"`), &decoded); err == nil {
				val = decoded
			}
			env[m[1]] = val
		} else {
			skipped = append(skipped, trimmed)
		}
		if readErr == io.EOF {
			return env, skipped, nil
		}
		if readErr != nil {
			return nil, nil, readErr
		}
	}
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// buildClaudeSettings renders the env map into the Claude Code `settings.json`
// shape: {"env": {...}}. Keys are written in sorted order for deterministic
// output (nicer for git diffs of stored providers).
func buildClaudeSettings(env map[string]string) ([]byte, error) {
	// Use a manual ordered marshal to keep key order stable.
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	buf.WriteString("{\n  \"env\": {")
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("\n    ")
		kb, _ := json.Marshal(k)
		vb, _ := json.Marshal(env[k])
		buf.Write(kb)
		buf.WriteString(": ")
		buf.Write(vb)
	}
	if len(keys) > 0 {
		buf.WriteString("\n  ")
	}
	buf.WriteString("}\n}\n")
	return []byte(buf.String()), nil
}

// CodexParams captures the answers collected from the codex wizard.
type CodexParams struct {
	BaseURL      string
	APIKey       string
	Model        string
	ProviderKey  string // identifier used inside the TOML table name
	ProviderName string // display name
	WireAPI      string
}

// buildCodexConfigTOML renders a minimal ~/.codex/config.toml compatible with
// what Codex expects. We emit it by hand to avoid pulling a toml dependency.
func buildCodexConfigTOML(p CodexParams) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "model = %q\n", p.Model)
	fmt.Fprintf(&b, "model_provider = %q\n\n", p.ProviderKey)
	fmt.Fprintf(&b, "[model_providers.%s]\n", p.ProviderKey)
	fmt.Fprintf(&b, "name = %q\n", p.ProviderName)
	fmt.Fprintf(&b, "base_url = %q\n", p.BaseURL)
	fmt.Fprintf(&b, "wire_api = %q\n", p.WireAPI)
	return []byte(b.String())
}

// buildCodexAuthJSON produces the auth.json that Codex reads for its API key.
func buildCodexAuthJSON(apiKey string) []byte {
	data, _ := json.MarshalIndent(map[string]string{"OPENAI_API_KEY": apiKey}, "", "  ")
	return append(data, '\n')
}

// maskSecret hides most of a secret, leaving only a short prefix/suffix so
// the user can still eyeball that the value they pasted is what they meant.
func maskSecret(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// looksSecret returns true for env keys that conventionally hold credentials,
// so we can mask them in the preview.
func looksSecret(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "token") || strings.Contains(k, "key") || strings.Contains(k, "secret") || strings.Contains(k, "password")
}

// ---- Interactive runners (stdin/stdout bound; not unit-tested) ----------

func prompt(stdin *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, err := stdin.ReadString('\n')
	if err != nil && line == "" {
		return def
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return def
	}
	return line
}

func confirm(stdin *bufio.Reader, label string, defYes bool) bool {
	suffix := "[Y/n]"
	if !defYes {
		suffix = "[y/N]"
	}
	fmt.Printf("%s %s ", label, suffix)
	line, _ := stdin.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defYes
	}
	return line == "y" || line == "yes"
}

// wizardClaude runs the interactive paste-based importer. It reads from the
// shared *bufio.Reader (so subsequent prompts see the rest of stdin) and
// returns the filename → contents map to write into the provider dir.
func wizardClaude(stdin *bufio.Reader) (map[string][]byte, error) {
	fmt.Println("Paste the official setup block. Accepted formats:")
	fmt.Println(`  export KEY=VAL   (shell)`)
	fmt.Println(`  KEY=VAL          (dotenv)`)
	fmt.Println(`  "KEY": "VAL",    (JSON)`)
	fmt.Println("End with a blank line.")
	fmt.Println()

	env, skipped, err := parseEnvBlock(stdin)
	if err != nil {
		return nil, err
	}
	for _, line := range skipped {
		fmt.Fprintf(os.Stderr, "  (ignored) %s\n", line)
	}
	if len(env) == 0 {
		return nil, fmt.Errorf("no KEY=VAL lines parsed")
	}

	fmt.Println("\nParsed env vars:")
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := env[k]
		if looksSecret(k) {
			v = maskSecret(v)
		}
		fmt.Printf("  %s=%s\n", k, v)
	}
	fmt.Println()

	if !confirm(stdin, "Save this provider?", true) {
		return nil, fmt.Errorf("cancelled")
	}

	data, err := buildClaudeSettings(env)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"settings.json": data}, nil
}

// wizardCodex runs the Q&A-based codex wizard.
func wizardCodex(stdin *bufio.Reader) (map[string][]byte, error) {
	p := CodexParams{
		BaseURL:     prompt(stdin, "Base URL", "https://api.openai.com/v1"),
		APIKey:      prompt(stdin, "API key (required)", ""),
		Model:       prompt(stdin, "Model", "gpt-4o"),
		ProviderKey: prompt(stdin, "Provider key (TOML identifier)", "openai"),
		WireAPI:     prompt(stdin, "wire_api (responses/chat)", "responses"),
	}
	if p.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	p.ProviderName = prompt(stdin, "Provider display name", upperFirst(p.ProviderKey))

	toml := buildCodexConfigTOML(p)
	auth := buildCodexAuthJSON(p.APIKey)

	fmt.Println("\nconfig.toml:")
	fmt.Print(string(toml))
	fmt.Printf("\nauth.json:\n{\n  \"OPENAI_API_KEY\": %q\n}\n\n", maskSecret(p.APIKey))

	if !confirm(stdin, "Save this provider?", true) {
		return nil, fmt.Errorf("cancelled")
	}
	return map[string][]byte{
		"config.toml": toml,
		"auth.json":   auth,
	}, nil
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// wizardClaudeTemplate creates a provider using a built-in template.
// The user only needs to supply the API key; all other env vars are pre-filled.
func wizardClaudeTemplate(stdin *bufio.Reader, tpl *app.ClaudeTemplate) (map[string][]byte, error) {
	fmt.Printf("Template: %s\n", tpl.DisplayName)
	if tpl.WebsiteURL != "" {
		fmt.Printf("Get your API key: %s\n", tpl.WebsiteURL)
	}
	fmt.Println()

	apiKey := prompt(stdin, tpl.APIKeyField+" (required)", "")
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Merge template env + user-supplied key.
	env := make(map[string]string, len(tpl.Env)+1)
	for k, v := range tpl.Env {
		env[k] = v
	}
	env[tpl.APIKeyField] = apiKey

	// Preview.
	fmt.Println("\nProvider config:")
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := env[k]
		if looksSecret(k) {
			v = maskSecret(v)
		}
		fmt.Printf("  %s=%s\n", k, v)
	}
	fmt.Println()

	if !confirm(stdin, "Save this provider?", true) {
		return nil, fmt.Errorf("cancelled")
	}

	data, err := buildClaudeSettings(env)
	if err != nil {
		return nil, err
	}
	return map[string][]byte{"settings.json": data}, nil
}
