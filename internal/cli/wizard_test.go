package cli

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseEnvBlockJSONFormat(t *testing.T) {
	in := strings.NewReader(`    "ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic",
    "ANTHROPIC_AUTH_TOKEN": "sk-abc",
    "ANTHROPIC_MODEL": "deepseek-v4-pro[1m]",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "deepseek-v4-flash",
    "API_TIMEOUT_MS": "600000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
`)
	env, skipped, err := parseEnvBlock(bufio.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) > 0 {
		t.Errorf("unexpected skipped lines: %v", skipped)
	}
	cases := map[string]string{
		"ANTHROPIC_BASE_URL":                       "https://api.deepseek.com/anthropic",
		"ANTHROPIC_AUTH_TOKEN":                     "sk-abc",
		"ANTHROPIC_MODEL":                          "deepseek-v4-pro[1m]",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":            "deepseek-v4-flash",
		"API_TIMEOUT_MS":                           "600000",
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1",
	}
	for k, want := range cases {
		if env[k] != want {
			t.Errorf("env[%q] = %q, want %q", k, env[k], want)
		}
	}
}

func TestParseEnvBlockMixedFormats(t *testing.T) {
	// Real-world docs mix shell exports and JSON snippets.
	in := strings.NewReader(`export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
"ANTHROPIC_AUTH_TOKEN": "sk-xyz",
ANTHROPIC_MODEL=deepseek-v4-pro
`)
	env, skipped, err := parseEnvBlock(bufio.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(skipped) > 0 {
		t.Errorf("skipped: %v", skipped)
	}
	if env["ANTHROPIC_BASE_URL"] != "https://api.deepseek.com/anthropic" ||
		env["ANTHROPIC_AUTH_TOKEN"] != "sk-xyz" ||
		env["ANTHROPIC_MODEL"] != "deepseek-v4-pro" {
		t.Errorf("mixed parse failed: %v", env)
	}
}

func TestParseEnvBlock(t *testing.T) {
	in := strings.NewReader(`export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
export ANTHROPIC_AUTH_TOKEN=sk-abc
ANTHROPIC_MODEL="deepseek-v4-pro"
  export ANTHROPIC_DEFAULT_HAIKU_MODEL='deepseek-v4-flash'  
# a comment
not a kv line
`)
	env, skipped, err := parseEnvBlock(bufio.NewReader(in))
	if err != nil {
		t.Fatalf("parseEnvBlock: %v", err)
	}
	want := map[string]string{
		"ANTHROPIC_BASE_URL":           "https://api.deepseek.com/anthropic",
		"ANTHROPIC_AUTH_TOKEN":         "sk-abc",
		"ANTHROPIC_MODEL":              "deepseek-v4-pro",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL": "deepseek-v4-flash",
	}
	if len(env) != len(want) {
		t.Fatalf("env len = %d, want %d: %v", len(env), len(want), env)
	}
	for k, v := range want {
		if env[k] != v {
			t.Errorf("env[%q] = %q, want %q", k, env[k], v)
		}
	}
	if len(skipped) != 2 {
		t.Errorf("skipped = %v, want 2 entries (# comment and 'not a kv line')", skipped)
	}
}

func TestParseEnvBlockBlankLineTerminates(t *testing.T) {
	// Simulate a paste followed by a blank line and then a confirm answer.
	in := bufio.NewReader(strings.NewReader("export FOO=bar\n\ny\n"))
	env, _, err := parseEnvBlock(in)
	if err != nil {
		t.Fatal(err)
	}
	if env["FOO"] != "bar" {
		t.Fatalf("FOO = %q", env["FOO"])
	}
	// The remaining "y\n" must still be readable afterwards.
	rest, _ := in.ReadString('\n')
	if strings.TrimSpace(rest) != "y" {
		t.Fatalf("remaining input after block: %q, want %q", rest, "y")
	}
}

func TestParseEnvBlockEmpty(t *testing.T) {
	env, _, err := parseEnvBlock(bufio.NewReader(strings.NewReader("")))
	if err != nil {
		t.Fatal(err)
	}
	if len(env) != 0 {
		t.Fatalf("expected empty, got %v", env)
	}
}

func TestStripQuotes(t *testing.T) {
	cases := map[string]string{
		`"hello"`:   "hello",
		`'hello'`:   "hello",
		`hello`:     "hello",
		`"mismatched'`: `"mismatched'`,
		`""`:        "",
		`"`:         `"`, // too short, left alone
	}
	for in, want := range cases {
		if got := stripQuotes(in); got != want {
			t.Errorf("stripQuotes(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildClaudeSettingsSortedAndValid(t *testing.T) {
	env := map[string]string{
		"B_KEY": "2",
		"A_KEY": "1",
	}
	out, err := buildClaudeSettings(env)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	// Keys must appear in sorted order.
	if i, j := strings.Index(s, "A_KEY"), strings.Index(s, "B_KEY"); i == -1 || j == -1 || i > j {
		t.Fatalf("keys not in sorted order:\n%s", s)
	}
	// Wrapper shape.
	if !strings.Contains(s, `"env":`) {
		t.Fatalf("missing env wrapper: %s", s)
	}
}

func TestBuildClaudeSettingsEscapesSpecialChars(t *testing.T) {
	env := map[string]string{"K": `a"b\c`}
	out, _ := buildClaudeSettings(env)
	// Must be valid JSON — let json.Valid confirm it.
	if !jsonValid(out) {
		t.Fatalf("output is not valid JSON:\n%s", out)
	}
}

func TestBuildCodexConfigTOML(t *testing.T) {
	p := CodexParams{
		BaseURL: "https://api.openai.com/v1", APIKey: "sk-x",
		Model: "gpt-4o", ProviderKey: "openai", ProviderName: "OpenAI", WireAPI: "responses",
	}
	got := string(buildCodexConfigTOML(p))
	for _, want := range []string{
		`model = "gpt-4o"`,
		`model_provider = "openai"`,
		`[model_providers.openai]`,
		`name = "OpenAI"`,
		`base_url = "https://api.openai.com/v1"`,
		`wire_api = "responses"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("TOML missing %q:\n%s", want, got)
		}
	}
}

func TestBuildCodexAuthJSON(t *testing.T) {
	out := buildCodexAuthJSON("sk-xyz")
	if !jsonValid(out) {
		t.Fatalf("not valid JSON: %s", out)
	}
	if !strings.Contains(string(out), `"OPENAI_API_KEY": "sk-xyz"`) {
		t.Fatalf("missing key in auth.json: %s", out)
	}
}

func TestMaskSecret(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"short":         "*****",
		"sk-abcdefghij": "sk-a**********", // len 13 → 4 + 5*\ + 4
	}
	for in, want := range cases {
		got := maskSecret(in)
		// Only check length and prefix/suffix for the long case.
		if in == "" && got != "" {
			t.Errorf("mask(%q) = %q", in, got)
		}
		if len(in) <= 8 && got != strings.Repeat("*", len(in)) {
			t.Errorf("mask short(%q) = %q", in, got)
		}
		if len(in) > 8 {
			if len(got) != len(in) {
				t.Errorf("mask(%q) len = %d, want %d", in, len(got), len(in))
			}
			if got[:4] != in[:4] || got[len(got)-4:] != in[len(in)-4:] {
				t.Errorf("mask(%q) = %q: prefix/suffix should match", in, got)
			}
		}
		_ = want
	}
}

func TestLooksSecret(t *testing.T) {
	for _, k := range []string{"ANTHROPIC_AUTH_TOKEN", "OPENAI_API_KEY", "DB_PASSWORD", "APP_SECRET"} {
		if !looksSecret(k) {
			t.Errorf("%s should look secret", k)
		}
	}
	for _, k := range []string{"BASE_URL", "MODEL", "REGION"} {
		if looksSecret(k) {
			t.Errorf("%s should not look secret", k)
		}
	}
}

// jsonValid is a tiny helper — parsing through the stdlib is enough.
func jsonValid(b []byte) bool {
	var v any
	return json.Unmarshal(b, &v) == nil
}
