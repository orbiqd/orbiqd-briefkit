# OrbiqD BriefKit

BriefKit runs your local, subscription-based agent CLIs (no APIs or API keys) and ships as a single, self-contained
install (no Python MCP), providing both a CLI and an MCP server.

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](go.mod)
![Coverage](https://img.shields.io/codecov/c/github/orbiqd/orbiqd-briefkit)

## Overview

OrbiqD BriefKit is a local orchestration tool that runs **your existing agent CLIs** directly in your current working
directory. Agents see your repository the same way you do, with no uploads and no remote context copying.

BriefKit is built for workflows where you want multiple LLMs to collaborate in the same workspace, with clean execution
logs and local state storage.

Supported local agent CLIs:

- Claude (`claude`)
- Codex (`codex`)
- Gemini (`gemini`)

## Shared Workspace Collaboration

BriefKit launches every agent inside the same working directory, so they can read and modify the same codebase. That
enables workflows like:

### Agent-to-agent context handoff

Let Codex analyze a module, then ask Gemini to validate or expand the findings from the same repo.

```bash
briefkit-ctl ask codex "Map the auth flow and list risk points"
briefkit-ctl ask gemini "Validate Codex findings and propose improvements"
```

### Agent challenge

Run the same prompt across agents and compare viewpoints.

```bash
briefkit-ctl ask claude "Review this PR for design issues"
briefkit-ctl ask codex "Review this PR for design issues"
briefkit-ctl ask gemini "Review this PR for design issues"
```

### Cross-agent review

Have one agent review work produced by another.

```bash
briefkit-ctl ask codex "Refactor the parser to reduce allocations"
briefkit-ctl ask claude "Review the parser refactor for correctness and style"
```

## Key Features

- Local-first execution in your current working directory
- Multi-agent orchestration with shared workspace access
- CLI for scripting and automation
- MCP server for integrations with MCP-compatible clients (e.g. run Codex from Claude)
- No API keys required (uses your local agent subscriptions)

## Installation

At least one agent CLI installed and available on your `PATH`:
- `claude`
- `codex`
- `gemini`

Install provides `briefkit-ctl`, `briefkit-mcp`, and `briefkit-runner`.

### Linux

Replace `VERSION` with the latest release version.

Deb package:

```bash
VERSION="1.2.3"
wget "https://github.com/orbiqd/orbiqd-briefkit/releases/download/v${VERSION}/briefkit_${VERSION}_linux_amd64.deb"
sudo dpkg -i "briefkit_${VERSION}_linux_amd64.deb"
```

RPM package:

```bash
VERSION="1.2.3"
wget "https://github.com/orbiqd/orbiqd-briefkit/releases/download/v${VERSION}/briefkit_${VERSION}_linux_amd64.rpm"
sudo rpm -i "briefkit_${VERSION}_linux_amd64.rpm"
```

Tarball install:

```bash
VERSION="1.2.3"
wget "https://github.com/orbiqd/orbiqd-briefkit/releases/download/v${VERSION}/briefkit_${VERSION}_linux_amd64.tar.gz"
tar -xzf "briefkit_${VERSION}_linux_amd64.tar.gz"
sudo install -m 0755 briefkit-ctl /usr/local/bin/briefkit-ctl
sudo install -m 0755 briefkit-mcp /usr/local/bin/briefkit-mcp
sudo install -m 0755 briefkit-runner /usr/local/bin/briefkit-runner
```

### Homebrew (macOS)

```bash
brew tap orbiqd/briefkit
brew install briefkit
```

### Agent discovery and MCP setup

Run setup to discover local agent CLIs, create configs, and register the MCP server with supported runtimes. This lets you ask Claude Code to run Codex, for example.

```bash
briefkit-ctl setup
```

Ask your first prompt.

```bash
briefkit-ctl ask codex "Explain the architecture of this repo."
```

Setup options you may need:

- `--runtime-kind claude|codex|gemini` to limit configuration to specific runtimes
- `--setup-agent-mcp=false` to skip MCP server registration
- `--force` to overwrite existing agent configs
- `--enable-sandbox=true|false` to override runtime sandbox defaults

## Go Library Usage

BriefKit can be embedded in Go apps by providing implementations of `Runner`, `ExecutionRepository`, and `ConfigRepository`.

```go
package main

import (
	"context"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

func main() {
	ctx := context.Background()

	var runner briefkit.Runner
	var executionRepo briefkit.ExecutionRepository
	var configRepo briefkit.ConfigRepository

	client := briefkit.NewLocalClient(runner, executionRepo, configRepo)

	result, err := client.Ask(
		ctx,
		"codex",
		"Summarize this workspace",
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

By default, `briefkit-ctl setup` registers the MCP server with supported runtimes. Use `--setup-agent-mcp=false` to skip registration.

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
