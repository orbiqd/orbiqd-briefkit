# OrbiqD BriefKit

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](go.mod)
![Coverage](https://img.shields.io/codecov/c/github/orbiqd/orbiqd-briefkit)

## Overview

BriefKit runs your local, subscription-based agent CLIs (no APIs or API keys) directly in your working directory. It
ships as a single, self-contained CLI and an MCP server, with no Python/JS and no extra runtimes. Agents see your
repository the same way you do, with no uploads and no remote context copying. BriefKit supports workflows where more
than one LLM collaborates in the same workspace.

Currently supported local agent CLIs are `claude`, `codex`, `gemini`.

## Usage

**Agent-to-agent context handoff**

Let Codex analyze a module, then ask Gemini to confirm or expand the findings from the same repo.

```bash
briefkit-ctl ask codex "Map the auth flow and list risk points"
briefkit-ctl ask gemini "Validate Codex findings and propose improvements"
```

**Agent challenge**

Run the same prompt across agents and compare viewpoints.

```bash
briefkit-ctl ask claude "Review this PR for design issues"
briefkit-ctl ask codex "Review this PR for design issues"
briefkit-ctl ask gemini "Review this PR for design issues"
```

**Cross-agent review**

Have one agent review work produced by another.

```bash
briefkit-ctl ask codex "Refactor the parser to reduce allocations"
briefkit-ctl ask claude "Review the parser refactor for correctness and style"
```

## Installation

### Prerequisites

At least one agent CLI installed and available on your `PATH` (`claude`, `codex`, `gemini`).

### Install binaries

<details>
<summary>Linux (.deb)</summary>

Requires `curl`, `jq`, and `dpkg`.

```bash
arch=$(uname -m); case "$arch" in x86_64|amd64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; *) echo "Unsupported arch: $arch" >&2; exit 1 ;; esac; \
url=$(curl -sL https://api.github.com/repos/orbiqd/orbiqd-briefkit/releases/latest | jq -r ".assets[] | select(.name | endswith(\"linux_${arch}.deb\")) | .browser_download_url" | head -n1); \
curl -L -o /tmp/briefkit.deb "$url" && sudo dpkg -i /tmp/briefkit.deb
```

</details>

<details>
<summary>Linux (.rpm)</summary>

Requires `curl`, `jq`, and `rpm`.

```bash
arch=$(uname -m); case "$arch" in x86_64|amd64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; *) echo "Unsupported arch: $arch" >&2; exit 1 ;; esac; \
url=$(curl -sL https://api.github.com/repos/orbiqd/orbiqd-briefkit/releases/latest | jq -r ".assets[] | select(.name | endswith(\"linux_${arch}.rpm\")) | .browser_download_url" | head -n1); \
curl -L -o /tmp/briefkit.rpm "$url" && sudo rpm -i /tmp/briefkit.rpm
```

</details>

<details>
<summary>Linux (tarball to /usr/bin)</summary>

Requires `curl`, `jq`, and `tar`.

```bash
arch=$(uname -m); case "$arch" in x86_64|amd64) arch=amd64 ;; aarch64|arm64) arch=arm64 ;; *) echo "Unsupported arch: $arch" >&2; exit 1 ;; esac; \
url=$(curl -sL https://api.github.com/repos/orbiqd/orbiqd-briefkit/releases/latest | jq -r ".assets[] | select(.name | endswith(\"linux_${arch}.tar.gz\")) | .browser_download_url" | head -n1); \
tmpdir=$(mktemp -d) && curl -L -o "$tmpdir/briefkit.tar.gz" "$url" && tar -xzf "$tmpdir/briefkit.tar.gz" -C "$tmpdir" && \
sudo install -m 0755 "$tmpdir/briefkit-ctl" /usr/bin/briefkit-ctl && \
sudo install -m 0755 "$tmpdir/briefkit-mcp" /usr/bin/briefkit-mcp && \
sudo install -m 0755 "$tmpdir/briefkit-runner" /usr/bin/briefkit-runner
```

</details>

<details>
<summary>Homebrew (macOS)</summary>

```bash
brew tap orbiqd/briefkit
brew install briefkit
```

</details>

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

`briefkit-ctl setup` creates configs at `~/.orbiqd/briefkit/agents/<agent-id>.yaml`.

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

<details>
<summary>briefkit-ctl agent list</summary>

Lists configured agents (JSON output).

```bash
briefkit-ctl agent list
```

</details>

<details>
<summary>briefkit-ctl agent add</summary>

Adds a new agent entry. Note: currently not implemented.

```bash
briefkit-ctl agent add <id> <kind> <path>
```

Kind values: `claude`, `codex`, `gemini`.

</details>

<details>
<summary>briefkit-ctl ask</summary>

Runs a prompt with a configured agent.

```bash
briefkit-ctl ask [flags] <agent-id> <prompt>
```

Flags:

- `--timeout` (duration, e.g. `30s`, `5m`)
- `--model` (override model)
- `--conversation-id` (continue a conversation)

</details>

<details>
<summary>briefkit-ctl setup</summary>

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

</details>

<details>
<summary>briefkit-ctl state execution list</summary>

Lists stored executions (JSON output).

```bash
briefkit-ctl state execution list
```

</details>

<details>
<summary>briefkit-ctl state execution show</summary>

Shows details for a single execution (JSON output).

```bash
briefkit-ctl state execution show <execution-id>
```

</details>

<details>
<summary>briefkit-ctl state execution create</summary>

Creates a new execution record (advanced use; JSON output with execution ID).

```bash
briefkit-ctl state execution create --agent-id <id> [flags] <prompt>
```

Flags:

- `-w`, `--working-dir` (default: `.`)
- `-t`, `--timeout` (default: `5m`)

</details>

## Troubleshooting

Problem: `briefkit-ctl setup` does not find your agent CLI.

- Verify the agent binary is in your PATH (`which claude`, `which codex`, `which gemini`).

Problem: `briefkit-ctl ask` fails with missing agent config.

- Run `briefkit-ctl setup` and confirm the agent appears in `briefkit-ctl agent list`.

Problem: MCP tools do not appear in your client.

- Ensure the MCP client points to the absolute path of `briefkit-mcp`.
- Restart the MCP client after config changes.

## License

MIT License. See [LICENSE](LICENSE).
