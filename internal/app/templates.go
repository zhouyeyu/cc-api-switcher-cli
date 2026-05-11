package app

import "sort"

// ClaudeTemplate is a built-in provider preset for Claude Code.
// It pre-fills all env vars except the API key, so the user only needs to
// supply their key when creating a new provider.
type ClaudeTemplate struct {
	// ID is the short identifier used on the command line, e.g. "deepseek".
	ID string
	// DisplayName is shown in listings.
	DisplayName string
	// APIKeyField is the env var name that holds the API key.
	// The user will be prompted to supply this value.
	APIKeyField string
	// Env holds all pre-filled env vars. APIKeyField is present with an empty
	// value as a placeholder; it will be replaced with the user-supplied key.
	Env map[string]string
	// WebsiteURL is shown to help users find their API key.
	WebsiteURL string
}

// claudeTemplates is the built-in preset registry for Claude Code providers.
var claudeTemplates = []*ClaudeTemplate{
	{
		ID:          "official",
		DisplayName: "Claude Official (Anthropic)",
		APIKeyField: "ANTHROPIC_AUTH_TOKEN",
		Env:         map[string]string{},
		WebsiteURL:  "https://console.anthropic.com/settings/keys",
	},
	{
		ID:          "deepseek",
		DisplayName: "DeepSeek",
		APIKeyField: "ANTHROPIC_AUTH_TOKEN",
		Env: map[string]string{
			"ANTHROPIC_BASE_URL":             "https://api.deepseek.com/anthropic",
			"ANTHROPIC_MODEL":                "deepseek-v4-pro",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "deepseek-v4-flash",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "deepseek-v4-pro",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   "deepseek-v4-pro",
		},
		WebsiteURL: "https://platform.deepseek.com",
	},
	{
		ID:          "kimi",
		DisplayName: "Kimi (Moonshot)",
		APIKeyField: "ANTHROPIC_AUTH_TOKEN",
		Env: map[string]string{
			"ANTHROPIC_BASE_URL":             "https://api.moonshot.cn/anthropic",
			"ANTHROPIC_MODEL":                "kimi-k2.6",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "kimi-k2.6",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "kimi-k2.6",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   "kimi-k2.6",
		},
		WebsiteURL: "https://platform.moonshot.cn/console",
	},
	{
		ID:          "glm",
		DisplayName: "Zhipu GLM",
		APIKeyField: "ANTHROPIC_AUTH_TOKEN",
		Env: map[string]string{
			"ANTHROPIC_BASE_URL":             "https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_MODEL":                "glm-5",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "glm-5",
			"ANTHROPIC_DEFAULT_SONNET_MODEL": "glm-5",
			"ANTHROPIC_DEFAULT_OPUS_MODEL":   "glm-5",
		},
		WebsiteURL: "https://open.bigmodel.cn",
	},
}

// AllClaudeTemplates returns the built-in Claude templates in a stable order.
func AllClaudeTemplates() []*ClaudeTemplate {
	out := make([]*ClaudeTemplate, len(claudeTemplates))
	copy(out, claudeTemplates)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ClaudeTemplateByID returns the template with the given ID, or false if not found.
func ClaudeTemplateByID(id string) (*ClaudeTemplate, bool) {
	for _, t := range claudeTemplates {
		if t.ID == id {
			return t, true
		}
	}
	return nil, false
}
