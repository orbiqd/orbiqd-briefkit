# OrbiqD BriefKit

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](go.mod)

## Overview

OrbiqD BriefKit is a local orchestration tool that exposes an **MCP server** to drive **subscription-based coding CLIs** (Codex / Claude Code / Gemini) — **no APIs, no API keys required**.

Built for workflows where you want to:
- Run multiple coding agents from a unified interface
- Maintain clean, explicit execution logs you can inspect and analyze
- Integrate agents into your own tools via the MCP protocol
- Automate multi-agent workflows with command-line scripting

## Key Features

- **Multiple Agent Support** - Work with Claude Code, Codex, and Gemini from one interface
- **MCP Server** - Integrate agents into Claude Desktop or any MCP-compatible client
- **CLI Tool** - Direct command-line access for scripting and automation
- **Execution State Management** - Track and inspect all agent interactions
- **Runtime Configuration** - Control features like web search and network access per agent
- **Conversation Continuity** - Resume conversations across multiple executions
- **Local-First** - All state stored locally on your filesystem, no external dependencies

## Architecture

BriefKit consists of three components:

- **`briefkit-ctl`** - Main CLI for direct agent interaction and management
- **`briefkit-mcp`** - MCP server exposing agents as tools
- **`briefkit-runner`** - Internal execution orchestrator (spawned automatically)

## Installation

### Prerequisites

- **Go 1.22 or later**
- **One or more supported agent CLIs installed:**
  - [Claude Code](https://claude.ai/download) (`claude` binary)
  - [Codex](https://codex.anthropic.com) (`codex` binary)
  - [Gemini](https://ai.google.dev) (`gemini` binary)

### Build from Source

```bash
git clone https://github.com/orbiqd/orbiqd-briefkit
cd orbiqd-briefkit
make build
```

Binaries will be available in `./bin/`:
- `./bin/briefkit-ctl`
- `./bin/briefkit-mcp`
- `./bin/briefkit-runner`

### Add to PATH (Optional)

```bash
export PATH="$PATH:/path/to/orbiqd-briefkit/bin"
```

Add this line to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to make it permanent.

## Quick Start

### 1. Discover Available Agents

BriefKit can auto-discover installed agent CLIs on your system:

```bash
briefkit-ctl agent discovery
```

This will scan your `$PATH` for supported agent executables.

### 2. Generate Default Configuration

Create configuration files for discovered agents:

```bash
briefkit-ctl agent discovery --write-default-config
```

This creates `~/.orbiqd/briefkit/agents/{agent-name}.yaml` for each discovered agent.

### 3. List Configured Agents

Verify your agent configurations:

```bash
briefkit-ctl agent list
```

### 4. Execute Your First Prompt

Run a prompt with one of your configured agents:

```bash
briefkit-ctl exec --agent-id claude-code "What is the capital of France?"
```

The agent ID comes from the filename in `~/.orbiqd/briefkit/agents/` (e.g., `claude-code.yaml` → `claude-code`).

### 5. Resume a Conversation

Continue a previous conversation by using the conversation ID from the output:

```bash
briefkit-ctl exec --agent-id claude-code \
  --conversation-id <conversation-id-from-previous-run> \
  "Tell me more about its history"
```

## Configuration

### Configuration Directory

BriefKit stores all configuration and state in `~/.orbiqd/briefkit/`:

```
~/.orbiqd/briefkit/
├── config.yaml           # Global configuration (optional)
├── agents/               # Agent definitions
│   ├── claude-code.yaml
│   ├── codex.yaml
│   └── gemini.yaml
├── state/                # Execution state
│   ├── executions/
│   ├── turns/
│   └── sessions/
└── logs/                 # Runtime logs
    └── runtime/
```

### Agent Configuration

Each agent is configured via a YAML file in `~/.orbiqd/briefkit/agents/`. The agent ID is derived from the filename (e.g., `codex.yaml` → agent ID `codex`).

#### Minimal Configuration

The simplest configuration specifies only the runtime kind:

```yaml
runtime:
  kind: claude-code  # or: codex, gemini
  config: {}
```

#### Full Configuration with Features

```yaml
runtime:
  kind: claude-code
  config:
    # Runtime-specific configuration (see below)
  feature:
    enableWebSearch: true      # Allow web search tool (where supported)
    enableNetworkAccess: true  # Allow network access (where supported)
```

### Runtime-Specific Configuration

#### Claude Code

```yaml
runtime:
  kind: claude-code
  config: {}  # No specific config options currently required
  feature:
    enableWebSearch: true  # Controls availability of web search tool
```

**Note:** `enableNetworkAccess` is not currently implemented for Claude Code.

#### Codex

```yaml
runtime:
  kind: codex
  config:
    requireWorkspaceRepository: true  # Enforce git repository (default: true)
  feature:
    enableWebSearch: true       # Controls web_search_request feature
    enableNetworkAccess: true   # Controls sandbox network access
```

#### Gemini

```yaml
runtime:
  kind: gemini
  config: {}  # No specific config options currently required
  feature:
    enableNetworkAccess: true  # When false, runs in sandboxed mode
```

**Note:** `enableWebSearch` is not currently implemented for Gemini.

### Environment Variables

- **`BRIEFKIT_RUNTIME_LOG_DIR`** - Override the runtime log directory (default: `~/.orbiqd/briefkit/logs/runtime/`)

## CLI Reference

### Agent Management

#### List Agents

```bash
briefkit-ctl agent list
```

Lists all configured agents from `~/.orbiqd/briefkit/agents/`.

**Output includes:**
- Agent ID
- Runtime kind (claude-code, codex, gemini)
- Runtime version (if available)

#### Discover Agents

```bash
briefkit-ctl agent discovery [--write-default-config] [--runtime-kind <kind>]
```

Discovers installed agent CLIs on your system by scanning `$PATH`.

**Options:**
- `--write-default-config` - Automatically generate configuration files for discovered agents
- `--runtime-kind <kind>` - Filter discovery to specific runtime(s) (can specify multiple)

**Examples:**

```bash
# Discover all supported agents
briefkit-ctl agent discovery

# Discover and create configs automatically
briefkit-ctl agent discovery --write-default-config

# Discover only Claude and Codex
briefkit-ctl agent discovery --runtime-kind claude-code --runtime-kind codex
```

### Execution

#### Execute Prompt

```bash
briefkit-ctl exec --agent-id <id> [options] <prompt>
```

Execute a prompt with the specified agent.

**Required:**
- `--agent-id <id>` - Agent identifier (from `briefkit-ctl agent list`)
- `<prompt>` - Prompt text to execute

**Options:**
- `--model <model>` - Override the default model for this execution
- `--conversation-id <id>` - Resume an existing conversation
- `--timeout <duration>` - Execution timeout (default: `5m`)
- `--auto` - Enable automatic mode (if supported by the agent)

**Examples:**

```bash
# Simple execution
briefkit-ctl exec --agent-id codex "Analyze this codebase structure"

# With model override
briefkit-ctl exec --agent-id claude-code --model claude-opus-4 "Review this pull request"

# Resume conversation
briefkit-ctl exec --agent-id gemini --conversation-id abc123 "Continue from where we left off"

# Custom timeout
briefkit-ctl exec --agent-id codex --timeout 10m "Perform a comprehensive security audit"
```

**Output:**
- Execution ID
- Conversation ID (for resuming)
- Agent response

### State Management

#### List Executions

```bash
briefkit-ctl state execution list
```

Lists all execution records stored in `~/.orbiqd/briefkit/state/executions/`.

**Output includes:**
- Execution ID
- Agent ID
- Status (pending, running, succeeded, failed)
- Created timestamp

#### Show Execution Details

```bash
briefkit-ctl state execution show <execution-id>
```

Display detailed information about a specific execution, including:
- Full input (prompt, model, configuration)
- Current status
- Result (if completed)
- Error details (if failed)

#### Create Execution

```bash
briefkit-ctl state execution create [options]
```

Manually create an execution record. This is an advanced command primarily used for automation and integration scenarios.

### Global Options

All commands support these global options:

- `--log-level <level>` - Set logging level (`debug`, `info`, `warn`, `error`)
- `--store-dir <path>` - Override state directory (default: `~/.orbiqd/briefkit`)

## MCP Server Usage

BriefKit exposes all configured agents as MCP tools, enabling integration with Claude Desktop and other MCP-compatible clients.

### Starting the MCP Server

```bash
./bin/briefkit-mcp
```

The server will run continuously and communicate via standard input/output using the MCP protocol.

### Claude Desktop Setup

#### 1. Locate Claude Desktop Configuration

The configuration file location depends on your operating system:

- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

#### 2. Add BriefKit MCP Server

Add the following to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "briefkit": {
      "command": "/absolute/path/to/briefkit-mcp"
    }
  }
}
```

**Important:** Use the absolute path to your `briefkit-mcp` binary.

#### 3. Restart Claude Desktop

Completely quit and restart Claude Desktop for the configuration to take effect.

### Available MCP Tools

For each configured agent, BriefKit exposes a tool following the naming pattern `exec_<agent_id>` (in snake_case).

**Example tools:**
- `exec_claude_code` - Execute prompts with Claude Code
- `exec_codex` - Execute prompts with Codex
- `exec_gemini` - Execute prompts with Gemini

### Tool Parameters

Each tool accepts the following parameters:

- **`prompt`** (required) - The instruction to send to the agent
- **`model`** (optional) - Override the default model for this execution
- **`conversationId`** (optional) - Resume an existing conversation session

### Example Usage in Claude Desktop

```
User: Use the exec_codex tool to analyze the project structure and identify potential performance bottlenecks.
```

Claude Desktop will invoke the tool and display the results from Codex.

To continue a conversation:

```
User: Use exec_codex with the conversationId from the previous response to propose optimizations.
```

### Debugging MCP Server

Use the MCP Inspector for debugging the server:

```bash
make debug-briefkit-mcp
```

This requires Node.js and will open an interactive debugging interface for testing MCP tools.

## Use Cases

### CLI Automation

Script multi-agent workflows for automation:

```bash
#!/bin/bash

# Get Claude's architectural analysis
claude_result=$(briefkit-ctl exec --agent-id claude-code "Analyze the system architecture")

# Get Codex's security review
codex_result=$(briefkit-ctl exec --agent-id codex "Review for security vulnerabilities")

# Get Gemini's test coverage analysis
gemini_result=$(briefkit-ctl exec --agent-id gemini "Analyze test coverage")

# Combine results for comprehensive review
echo "=== Multi-Agent Analysis ===" > report.md
echo "$claude_result" >> report.md
echo "$codex_result" >> report.md
echo "$gemini_result" >> report.md
```

### MCP Integration

Seamlessly use multiple agents within Claude Desktop conversations:

1. Ask Claude Desktop a question
2. Have it consult Codex for code-specific analysis via `exec_codex`
3. Continue the conversation with context from both agents
4. No need to switch between different tools or terminals

### Execution Tracking

All agent interactions are logged and inspectable:

```bash
# View all execution history
briefkit-ctl state execution list

# Inspect a specific execution
briefkit-ctl state execution show <execution-id>

# Or directly access the filesystem
cat ~/.orbiqd/briefkit/state/executions/<execution-id>/result.json | jq
```

### Conversation Continuity

Build long-running conversations across multiple sessions:

```bash
# Start a conversation
RESULT=$(briefkit-ctl exec --agent-id claude-code "Let's design a new feature for user authentication")

# Extract conversation ID from result
CONV_ID=$(echo "$RESULT" | jq -r '.conversationId')

# Continue later (even after restart)
briefkit-ctl exec --agent-id claude-code --conversation-id "$CONV_ID" \
  "Now let's add tests for the authentication feature"

# Keep going
briefkit-ctl exec --agent-id claude-code --conversation-id "$CONV_ID" \
  "Deploy this to staging"
```

## State & Storage

BriefKit maintains all state on the local filesystem for transparency and debuggability.

### State Directory Structure

```
~/.orbiqd/briefkit/state/
├── executions/<execution-id>/
│   ├── input.json      # Execution request
│   ├── status.json     # Current execution status
│   └── result.json     # Final result (when complete)
├── turns/<turn-id>/
│   ├── request.json    # Turn request
│   ├── response.json   # Turn response
│   └── status.json     # Turn status
└── sessions/<session-id>/
    └── transcript.ndjson  # Session transcript
```

### Execution States

An execution can be in one of four states:

- **`pending`** - Execution created, waiting to start
- **`running`** - Currently executing
- **`succeeded`** - Completed successfully
- **`failed`** - Failed with error

### Inspecting State

You can inspect execution state through the CLI or directly via the filesystem:

```bash
# CLI method
briefkit-ctl state execution list
briefkit-ctl state execution show <execution-id>

# Direct filesystem access
cat ~/.orbiqd/briefkit/state/executions/<execution-id>/status.json | jq
cat ~/.orbiqd/briefkit/state/executions/<execution-id>/result.json | jq
```

### Runtime Logs

Execution logs are stored in:

```
~/.orbiqd/briefkit/logs/runtime/<execution-id>.log
```

Override the log directory with the `BRIEFKIT_RUNTIME_LOG_DIR` environment variable:

```bash
export BRIEFKIT_RUNTIME_LOG_DIR=/tmp/briefkit-logs
```

## Troubleshooting

### Agent Not Discovered

**Problem:** `briefkit-ctl agent discovery` doesn't find your installed agent.

**Solutions:**
- Ensure the agent CLI is in your `$PATH` (try `which claude` or `which codex`)
- Verify the agent binary is executable: `ls -la $(which claude)`
- Try discovery with explicit runtime: `--runtime-kind claude-code`
- Check that you're using the correct binary name (must be `claude`, `codex`, or `gemini`)

### Agent Config Not Found

**Problem:** `briefkit-ctl exec` fails with "agent config not found" or similar error.

**Solutions:**
- Run `briefkit-ctl agent list` to see available configured agents
- Generate configs automatically: `briefkit-ctl agent discovery --write-default-config`
- Manually create a config file in `~/.orbiqd/briefkit/agents/<agent-id>.yaml`
- Verify the agent ID matches the filename without the `.yaml` extension

### Execution Timeout

**Problem:** Execution times out before the agent finishes responding.

**Solution:**

Increase the timeout with the `--timeout` flag:

```bash
briefkit-ctl exec --agent-id codex --timeout 10m "Complex analysis task"
```

Default timeout is 5 minutes. Use values like `30s`, `5m`, `1h`.

### MCP Tools Not Appearing in Claude Desktop

**Problem:** BriefKit tools don't show up in Claude Desktop.

**Solutions:**
- Verify the config file path (see "Claude Desktop Setup" above)
- Use an **absolute path** to `briefkit-mcp` in the configuration
- Check JSON syntax is valid (use a JSON validator)
- Completely quit and restart Claude Desktop (⌘Q on macOS, not just close window)
- Check Claude Desktop logs for errors:
  - macOS: `~/Library/Logs/Claude/mcp*.log`
  - Check for connection errors or path issues

### View Runtime Logs

**Problem:** Need to debug what's happening during agent execution.

**Solution:**

Runtime logs are stored per execution:

```bash
# Find your execution ID
briefkit-ctl state execution list

# View the log
tail -f ~/.orbiqd/briefkit/logs/runtime/<execution-id>.log

# Or set a custom log directory
export BRIEFKIT_RUNTIME_LOG_DIR=/tmp/briefkit-logs
briefkit-ctl exec --agent-id claude-code "test"
tail -f /tmp/briefkit-logs/*.log
```

## Roadmap

Features planned for future releases:

- **Multi-Agent Collaboration** - Shared sessions where multiple agents can work together on the same problem with structured turn-taking and context sharing
- **Interactive Agent Add** - CLI command to add new agents interactively with guided configuration
- **Web Dashboard** - Web UI for monitoring executions, browsing history, and managing configurations
- **Execution Search & Filtering** - Query past executions by agent, status, date, or content
- **Custom Runtime Plugins** - Support for additional agent CLIs beyond the built-in three

See [GitHub Issues](https://github.com/orbiqd/orbiqd-briefkit/issues) for detailed discussion and progress tracking.

## Development

### Development Setup

```bash
git clone https://github.com/orbiqd/orbiqd-briefkit
cd orbiqd-briefkit
go mod tidy
make build
go test ./...
```

### Documentation

For detailed development information, see:

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development guide, architecture, core concepts, and contribution guidelines
- **[CLAUDE.md](CLAUDE.md)** - Instructions for Claude Code when working on this project
- **[GEMINI.md](GEMINI.md)** - Instructions for Gemini when working on this project
- **[AGENTS.md](AGENTS.md)** - Instructions for Codex when working on this project

## Contributing

Contributions are welcome! We appreciate:

- Bug reports and feature requests via [GitHub Issues](https://github.com/orbiqd/orbiqd-briefkit/issues)
- Documentation improvements
- Pull requests for bug fixes and features
- Sharing your use cases and workflows

### Before Contributing

1. Read [DEVELOPMENT.md](DEVELOPMENT.md) for coding standards and architecture
2. Check existing issues to avoid duplicates
3. Run tests: `go test ./...`
4. Format code: `gofmt -w .`

### Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Make your changes following the coding standards
4. Run tests and ensure they pass
5. Commit with conventional commit messages
6. Push to your fork and create a pull request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) for MCP server implementation
- Uses [alecthomas/kong](https://github.com/alecthomas/kong) for CLI parsing
- Uses [spf13/afero](https://github.com/spf13/afero) for filesystem abstraction
- Inspired by the need for local-first AI agent orchestration
