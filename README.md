# ccsw — AI CLI API Switcher

[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md)

Switch API providers for AI command-line tools (Claude Code, Codex) in one command.
Zero dependencies. Single static binary. Scriptable.

```bash
ccsw use claude deepseek
```

## Install

```bash
# Requires Go 1.21+
go install github.com/cc-api-switcher-cli/ccsw/cmd/ccsw@latest

# Add to PATH (one-time)
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

Or download a prebuilt binary from the Releases page and drop it into `$PATH`.

## Quick start

```bash
# 1. Initialize
ccsw init

# 2a. Built-in template — only enter your API key
ccsw new claude deepseek --template deepseek

# 2b. Or paste any config block from provider docs (shell/dotenv/JSON)
ccsw new claude myprovider

# 3. Switch
ccsw use claude deepseek

# 4. Check status
ccsw ls
ccsw current
```

## Built-in templates

```bash
ccsw templates        # list all
```

| Template | Provider |
| --- | --- |
| `official` | Claude Official (Anthropic) |
| `deepseek` | DeepSeek |
| `kimi` | Kimi (Moonshot) |
| `glm` | Zhipu GLM |

Using a template only prompts for your API key — URLs and model names are pre-filled:

```
$ ccsw new claude deepseek --template deepseek
Template: DeepSeek
Get your API key: https://platform.deepseek.com

ANTHROPIC_AUTH_TOKEN (required): sk-...

Provider config:
  ANTHROPIC_AUTH_TOKEN=sk-a***xyz
  ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
  ANTHROPIC_DEFAULT_HAIKU_MODEL=deepseek-v4-flash
  ANTHROPIC_MODEL=deepseek-v4-pro
  ...

Save this provider? [Y/n]
created claude provider "deepseek"
switch with: ccsw use claude deepseek
```

## Paste wizard (`ccsw new` without `--template`)

Accepts any format directly from provider docs — all three work:

```bash
# Shell export
export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
export ANTHROPIC_AUTH_TOKEN=sk-...

# dotenv
ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic

# JSON (copy straight from a settings.json env block)
"ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic",
"ANTHROPIC_AUTH_TOKEN": "sk-...",
"API_TIMEOUT_MS": "600000",
```

Formats can be mixed. Secrets are masked in the preview. End input with a blank line.

`ccsw new codex <name>` uses a short Q&A to generate `config.toml` + `auth.json`.

## Storage layout

```
~/.ccsw/
├── state.json
└── providers/
    ├── claude/
    │   ├── deepseek/
    │   │   └── settings.json    → ~/.claude/settings.json
    │   └── official/
    │       └── settings.json
    └── codex/
        └── openai/
            ├── config.toml      → ~/.codex/config.toml
            └── auth.json        → ~/.codex/auth.json
```

Each provider is a plain directory of native config files. Put it under version control,
sync across machines, edit with any tool.

## Commands

| Command | Description |
| --- | --- |
| `ccsw init` | Create `~/.ccsw` skeleton |
| `ccsw new <app> <name> [--template <t>]` | Create provider via wizard or template |
| `ccsw templates` | List built-in provider templates |
| `ccsw add <app> <name> <src-dir>` | Import provider from a directory |
| `ccsw use <app> <name>` | Switch to provider |
| `ccsw ls [app]` | List providers |
| `ccsw current` | Show active provider per app |
| `ccsw rm <app> <name>` | Delete a provider |
| `ccsw edit <app> <name>` | Open provider dir in `$EDITOR` |
| `ccsw path <app> <name>` | Print provider directory path |
| `ccsw version` | Print version |

Supported `<app>`: `claude`, `codex`.

## Safety

- Atomic writes via temp-file + rename — no half-written configs.
- Previous config renamed to `.bak` before overwrite (one generation kept).
- Multi-file rollback: if a switch fails mid-way, already-written files are
  restored from `.bak` automatically.
- Only writes to `~/.ccsw/` and each app's declared target paths. Nothing else.

## License

MIT
