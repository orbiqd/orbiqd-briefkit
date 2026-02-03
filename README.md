# OrbiqD BriefKit
BriefKit runs your local, subscription-based agent CLIs (no APIs or API keys) and ships as a single, self-contained install (no Python MCP), providing both a CLI and an MCP server.

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](go.mod)
![Coverage](https://img.shields.io/codecov/c/github/orbiqd/orbiqd-briefkit)

## Overview

OrbiqD BriefKit is a local orchestration tool that runs **your existing agent CLIs** directly in your current working directory. Agents see your repository the same way you do, with no uploads and no remote context copying.

BriefKit is built for workflows where you want multiple LLMs to collaborate in the same workspace, with clean execution logs and local state storage.

Supported local agent CLIs:
- Claude (`claude`)
- Codex (`codex`)
- Gemini (`gemini`)

## Shared Workspace Collaboration

BriefKit launches every agent inside the same working directory, so they can read and modify the same codebase. That enables workflows like:

1. Agent-to-agent context handoff: let Codex analyze a module, then ask Gemini to validate or expand the findings from the same repo.

```bash
briefkit-ctl ask codex "Map the auth flow and list risk points"
briefkit-ctl ask gemini "Validate Codex findings and propose improvements"
```

2. Agent challenge: run the same prompt across agents and compare viewpoints.

```bash
briefkit-ctl ask claude "Review this PR for design issues"
briefkit-ctl ask codex "Review this PR for design issues"
briefkit-ctl ask gemini "Review this PR for design issues"
```

3. Cross-agent review: have one agent review work produced by another.

```bash
briefkit-ctl ask codex "Refactor the parser to reduce allocations"
briefkit-ctl ask claude "Review the parser refactor for correctness and style"
```

## Key Features

- Local-first execution in your current working directory
- Multi-agent orchestration with shared workspace access
- CLI for scripting and automation
- MCP server for integrations with MCP-compatible clients
- Clear execution state and JSON outputs
- No API keys required (uses your local agent subscriptions)

## Installation

Prerequisites:
- Go is not required when using prebuilt binaries
- One or more agent CLIs installed and on your PATH: `claude`, `codex`, `gemini`
- Install provides `briefkit-ctl`, `briefkit-mcp`, and `briefkit-runner`

### Automated install (script)

This installs the latest release to `~/.local/bin` (override with `INSTALL_DIR`).
Ensure the install directory is on your `PATH`.

```bash
set -euo pipefail

repo="orbiqd/orbiqd-briefkit"
version="$(curl -fsSL "https://api.github.com/repos/${repo}/releases/latest" | grep -Eo '"tag_name":\s*"[^"]+"' | cut -d'"' -f4)"
version_no_v="${version#v}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "Unsupported arch: $arch"; exit 1 ;;
 esac

case "$os" in
  darwin) ext="zip" ;;
  linux) ext="tar.gz" ;;
  *) echo "Unsupported OS: $os"; exit 1 ;;
 esac

asset="briefkit_${version_no_v}_${os}_${arch}.${ext}"
url="https://github.com/${repo}/releases/download/${version}/${asset}"

install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$install_dir"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

curl -fsSL "$url" -o "$tmp_dir/briefkit.${ext}"

if [ "$ext" = "zip" ]; then
  unzip -q "$tmp_dir/briefkit.${ext}" -d "$tmp_dir"
else
  tar -xzf "$tmp_dir/briefkit.${ext}" -C "$tmp_dir"
fi

install -m 0755 "$tmp_dir/briefkit-ctl" "$install_dir/briefkit-ctl"
install -m 0755 "$tmp_dir/briefkit-mcp" "$install_dir/briefkit-mcp"
install -m 0755 "$tmp_dir/briefkit-runner" "$install_dir/briefkit-runner"

echo "Installed to $install_dir"
```

### Homebrew (macOS)

```bash
brew tap orbiqd/briefkit
brew install briefkit
```

### Manual install (.deb on Linux)

Replace `VERSION` with the latest release version.

```bash
VERSION="1.2.3"
wget "https://github.com/orbiqd/orbiqd-briefkit/releases/download/v${VERSION}/briefkit_${VERSION}_linux_amd64.deb"
sudo dpkg -i "briefkit_${VERSION}_linux_amd64.deb"
```

## Setup

1. Run setup to discover local agent CLIs and create configs.

```bash
briefkit-ctl setup
```

2. Verify configured agents.

```bash
briefkit-ctl agent list
```

3. Ask your first prompt.

```bash
briefkit-ctl ask codex "Explain the architecture of this repo"
```

4. Continue a conversation.

```bash
briefkit-ctl ask codex --conversation-id <conversation-id> "Continue with deeper details"
```

Setup options you may need:
- `--runtime-kind claude|codex|gemini` to limit configuration to specific runtimes
- `--setup-agent-mcp=false` to skip MCP server registration
- `--force` to overwrite existing agent configs
- `--enable-sandbox=true|false` to override runtime sandbox defaults

## Go Library Usage

BriefKit can be embedded in Go apps by implementing the repository and runner interfaces.

```go
package main

import (
  "context"
  "time"

  "github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

type executionRepo struct{}
// Implement briefkit.ExecutionRepository

type configRepo struct{}
// Implement briefkit.ConfigRepository

type runner struct{}
// Implement briefkit.Runner

func main() {
  ctx := context.Background()

  client := briefkit.NewLocalClient(&runner{}, &executionRepo{}, &configRepo{})

  result, err := client.Ask(
    ctx,
    "codex",
    "Summarize this workspace",
    briefkit.AskWithTimeout(5*time.Minute),
  )
  if err != nil {
    panic(err)
  }

  _ = result
}
```

## Configuration and Paths

Default paths:
- Agent configs: `~/.orbiqd/briefkit/agents`
- State store: `~/.orbiqd/briefkit/state`
- Runtime logs: `~/.orbiqd/briefkit/logs/runtime`

Override paths with environment variables:
- `BRIEFKIT_AGENT_CONFIG_PATH`
- `BRIEFKIT_STATE_PATH`
- `BRIEFKIT_RUNTIME_LOG_DIR`

Example agent config (`~/.orbiqd/briefkit/agents/codex.yaml`):

```yaml
runtime:
  kind: codex
  config:
    requireWorkspaceRepository: false
  feature:
    enableSandbox: false
```

Runtime kinds:
- `claude`
- `codex`
- `gemini`

Configs are created automatically by `briefkit-ctl setup` as `~/.orbiqd/briefkit/agents/<agent-id>.yaml`.

## MCP Server

Start the MCP server:

```bash
briefkit-mcp
```

To use BriefKit from an MCP client, point the client to the `briefkit-mcp` executable. Example Claude Desktop config:

- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\\Claude\\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "briefkit": {
      "command": "/absolute/path/to/briefkit-mcp"
    }
  }
}
```

## CLI Manual

### briefkit-ctl

Usage:

```bash
briefkit-ctl [global flags] <command>
```

Global flags:
- `-v`, `--log-level` (`debug|info|warn|error`, default: `info`)
- `--log-format` (`text-color|text-no-color|json`, default: `text-color`)
- `--log-quiet` (disable logging)
- `-s`, `--store-state-path` (default: `~/.orbiqd/briefkit/state`)
- `--store-agent-config-path` (default: `~/.orbiqd/briefkit/agents`)
- `--version` (print version)

#### briefkit-ctl agent list

Lists configured agents (JSON output).

```bash
briefkit-ctl agent list
```

#### briefkit-ctl agent add

Adds a new agent entry. Note: currently not implemented.

```bash
briefkit-ctl agent add <id> <kind> <path>
```

Kind values: `claude`, `codex`, `gemini`.

#### briefkit-ctl ask

Runs a prompt with a configured agent.

```bash
briefkit-ctl ask [flags] <agent-id> <prompt>
```

Flags:
- `--timeout` (duration, e.g. `30s`, `5m`)
- `--model` (override model)
- `--conversation-id` (continue a conversation)

#### briefkit-ctl setup

Auto-discovers local agent CLIs, creates configs, and optionally registers the MCP server with supported runtimes.

```bash
briefkit-ctl setup [flags]
```

Flags:
- `--runtime-kind` (repeatable; one of `claude`, `codex`, `gemini`)
- `--setup-agent-config` (default: `true`)
- `--setup-agent-mcp` (default: `true`)
- `--enable-sandbox` (set `true` or `false` to override defaults)
- `--force` (overwrite existing configs)

#### briefkit-ctl state execution list

Lists stored executions (JSON output).

```bash
briefkit-ctl state execution list
```

#### briefkit-ctl state execution show

Shows details for a single execution (JSON output).

```bash
briefkit-ctl state execution show <execution-id>
```

#### briefkit-ctl state execution create

Creates a new execution record (advanced use; JSON output with execution ID).

```bash
briefkit-ctl state execution create --agent-id <id> [flags] <prompt>
```

Flags:
- `-w`, `--working-dir` (default: `.`)
- `-t`, `--timeout` (default: `5m`)

### briefkit-mcp

Usage:

```bash
briefkit-mcp [global flags]
```

Global flags:
- `-v`, `--log-level` (`debug|info|warn|error`, default: `info`)
- `--log-format` (`text-color|text-no-color|json`, default: `text-color`)
- `--log-quiet` (disable logging)
- `-s`, `--store-state-path` (default: `~/.orbiqd/briefkit/state`)
- `--store-agent-config-path` (default: `~/.orbiqd/briefkit/agents`)
- `--version` (print version)

### briefkit-runner

Executes a single stored execution (used internally by BriefKit, also available for advanced usage).

```bash
briefkit-runner [global flags] <execution-id>
```

Flags:
- `--retry` (allow rerunning finished executions)
- `--version` (print version)

Global flags:
- `-v`, `--log-level` (`debug|info|warn|error`, default: `info`)
- `--log-format` (`text-color|text-no-color|json`, default: `text-color`)
- `--log-quiet` (disable logging)
- `-s`, `--store-state-path` (default: `~/.orbiqd/briefkit/state`)
- `--store-agent-config-path` (default: `~/.orbiqd/briefkit/agents`)

## Troubleshooting

Problem: `briefkit-ctl setup` does not find your agent CLI.
- Verify the agent binary is installed and in your PATH (`which claude`, `which codex`, `which gemini`).

Problem: `briefkit-ctl ask` fails with missing agent config.
- Run `briefkit-ctl setup` and confirm the agent appears in `briefkit-ctl agent list`.

Problem: MCP tools do not appear in your client.
- Ensure the MCP client points to the absolute path of `briefkit-mcp`.
- Restart the MCP client after config changes.

## License

MIT License. See [LICENSE](LICENSE).
