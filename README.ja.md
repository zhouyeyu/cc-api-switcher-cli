# ccsw — AI CLI API スイッチャー

[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md)

1コマンドで AI コマンドラインツール（Claude Code、Codex）の API プロバイダーを切り替える。
ゼロ依存。単一スタティックバイナリ。スクリプト対応。

```bash
ccsw use claude deepseek
```

## インストール

```bash
# Go 1.21+ が必要
go install github.com/zhouyeyu/cc-api-switcher-cli/cmd/ccsw@latest

# PATH に追加（初回のみ）
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

またはリリースページからプリビルドバイナリをダウンロードして `$PATH` に配置。

## クイックスタート

```bash
# 1. 初期化
ccsw init

# 2a. 最速：組み込みテンプレートで API キーだけ入力
ccsw new claude deepseek --template deepseek

# 2b. またはプロバイダーのドキュメントから設定ブロックをペースト
ccsw new claude myprovider

# 3. 切り替え
ccsw use claude deepseek

# 4. 状態確認
ccsw ls
ccsw current
```

## 組み込みテンプレート

```bash
ccsw templates        # 一覧表示
```

| テンプレート ID | プロバイダー |
| --- | --- |
| `official` | Claude 公式（Anthropic） |
| `deepseek` | DeepSeek |
| `kimi` | Kimi（Moonshot） |
| `glm` | Zhipu GLM |

テンプレートを使うと URL とモデル名は自動入力済み — **API キーを入力するだけ**：

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

## ペーストウィザード（`--template` なしの `ccsw new`）

プロバイダーのドキュメントから設定をそのままペースト — **3つの形式に対応**：

```bash
# Shell export 形式
export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
export ANTHROPIC_AUTH_TOKEN=sk-...

# dotenv 形式
ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic

# JSON 形式（settings.json の env ブロックから直接コピー）
"ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic",
"ANTHROPIC_AUTH_TOKEN": "sk-...",
"API_TIMEOUT_MS": "600000",
```

形式は混在してもOK。シークレット系フィールドはプレビュー時にマスク表示。空行で入力終了。

`ccsw new codex <name>` は Q&A 形式で `config.toml` と `auth.json` を自動生成。

## ストレージ構造

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

各プロバイダーはネイティブ設定ファイルを含む普通のディレクトリ。Git 管理、マシン間同期、任意のエディタでの編集が可能。

## コマンドリファレンス

| コマンド | 説明 |
| --- | --- |
| `ccsw init` | `~/.ccsw` ディレクトリを作成 |
| `ccsw new <app> <name> [--template <t>]` | ウィザードまたはテンプレートでプロバイダーを作成 |
| `ccsw templates` | 組み込みテンプレートを一覧表示 |
| `ccsw add <app> <name> <src-dir>` | 既存ディレクトリからプロバイダーをインポート |
| `ccsw use <app> <name>` | プロバイダーに切り替え |
| `ccsw ls [app]` | プロバイダーを一覧表示 |
| `ccsw current` | 現在アクティブなプロバイダーを表示 |
| `ccsw rm <app> <name>` | プロバイダーを削除 |
| `ccsw edit <app> <name>` | `$EDITOR` でプロバイダーディレクトリを開く |
| `ccsw path <app> <name>` | プロバイダーディレクトリのパスを表示 |
| `ccsw version` | バージョンを表示 |

対応 `<app>`：`claude`、`codex`。

## 安全性

- **アトミック書き込み**：一時ファイル + rename で半書き込み状態を防止。
- **自動バックアップ**：上書き前に既存ファイルを `.bak` にリネーム（1世代保持）。
- **マルチファイルロールバック**：切り替え中に失敗した場合、書き込み済みファイルを自動的に `.bak` から復元。
- **スコープ分離**：`~/.ccsw/` と各アプリの宣言済みターゲットパスのみ書き込み。

## License

MIT
