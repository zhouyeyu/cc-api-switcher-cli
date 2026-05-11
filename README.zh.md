# ccsw — AI CLI API 切换工具

[English](README.md) | [中文](README.zh.md) | [日本語](README.ja.md)

一条命令，在不同 API 供应商之间切换 AI 命令行工具（Claude Code、Codex）的配置。
零依赖，单一静态二进制，可脚本化。

```bash
ccsw use claude deepseek
```

## 安装

```bash
# 需要 Go 1.21+
go install github.com/cc-api-switcher-cli/ccsw/cmd/ccsw@latest

# 把 ~/go/bin 加入 PATH（一次性操作）
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc && source ~/.zshrc
```

或前往 Releases 页面下载对应平台的预编译二进制，放进 `$PATH` 即可。

## 快速开始

```bash
# 1. 初始化
ccsw init

# 2a. 最快：内置模板，只需输入 API Key
ccsw new claude deepseek --template deepseek

# 2b. 或粘贴官方文档的配置块（支持 shell/dotenv/JSON 三种格式）
ccsw new claude myprovider

# 3. 切换
ccsw use claude deepseek

# 4. 查看状态
ccsw ls
ccsw current
```

## 内置模板

```bash
ccsw templates        # 列出全部模板
```

| 模板 ID | 供应商 |
| --- | --- |
| `official` | Claude 官方（Anthropic） |
| `deepseek` | DeepSeek |
| `kimi` | Kimi（月之暗面） |
| `glm` | 智谱 GLM |

使用模板时，URL 和模型名已预填，**只需输入你的 API Key**：

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

## 粘贴向导（`ccsw new` 不带 `--template`）

直接把供应商文档里的配置粘进来，**三种格式均支持**：

```bash
# Shell export 格式
export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic
export ANTHROPIC_AUTH_TOKEN=sk-...

# dotenv 格式
ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic

# JSON 格式（直接从 settings.json 的 env 块复制）
"ANTHROPIC_BASE_URL": "https://api.deepseek.com/anthropic",
"ANTHROPIC_AUTH_TOKEN": "sk-...",
"ANTHROPIC_MODEL": "deepseek-v4-pro[1m]",
"API_TIMEOUT_MS": "600000",
```

三种格式可以混着粘，工具自动识别。token/key 字段预览时自动打码，空行结束输入。

`ccsw new codex <name>` 采用逐项问答，自动生成 `config.toml` 和 `auth.json`。

## 存储结构

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

每个 provider 是一个普通目录，存放目标 CLI 的原生格式配置文件。可以放进 git 管理，跨机器同步，用任何编辑器修改。

## 命令参考

| 命令 | 说明 |
| --- | --- |
| `ccsw init` | 创建 `~/.ccsw` 目录骨架 |
| `ccsw new <app> <name> [--template <t>]` | 向导或模板创建新 provider |
| `ccsw templates` | 列出内置模板 |
| `ccsw add <app> <name> <src-dir>` | 从现有目录导入 provider |
| `ccsw use <app> <name>` | 切换到指定 provider |
| `ccsw ls [app]` | 列出所有（或指定 app 的）provider |
| `ccsw current` | 显示当前各 app 激活的 provider |
| `ccsw rm <app> <name>` | 删除一个 provider |
| `ccsw edit <app> <name>` | 用 `$EDITOR` 打开 provider 目录 |
| `ccsw path <app> <name>` | 打印 provider 目录绝对路径 |
| `ccsw version` | 显示版本号 |

支持的 `<app>`：`claude`、`codex`。

## 安全机制

- **原子写入**：临时文件 + rename，绝不产生半写状态。
- **自动备份**：覆盖前把旧配置重命名为 `.bak`，保留一代。
- **多文件回滚**：切换中途失败时，已写入的文件自动从 `.bak` 还原。
- **作用域隔离**：只写 `~/.ccsw/` 和各 app 注册表声明的目标路径，不碰其他文件。

## License

MIT
