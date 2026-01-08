package claude

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

var semverPattern = regexp.MustCompile(`\d+\.\d+\.\d+`)

const Claude = agent.RuntimeKind("claude")

type RuntimeConfig struct {
}

type Runtime struct {
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func (runtime *Runtime) Execute(ctx context.Context, executionId agent.ExecutionID, executionInput agent.ExecutionInput, agentConfig agent.Config) (agent.RuntimeInstance, error) {
	logDir, err := cli.ResolveRuntimeLogDir()
	if err != nil {
		return nil, err
	}

	runtimeConfig, err := utils.AnyToStruct[RuntimeConfig](agentConfig.Runtime.Config)
	if err != nil {
		return nil, fmt.Errorf("convert runtime config: %w", err)
	}

	instance, err := newInstance(ctx, executionId, executionInput, *runtimeConfig, agentConfig.Runtime.Feature, logDir)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (runtime *Runtime) Discovery(ctx context.Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	_, err := locateExecutable(ctx)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, exec.ErrNotFound) {
		return false, nil
	}

	return false, err
}

func (runtime *Runtime) GetDefaultConfig(ctx context.Context) (agent.RuntimeConfig, error) {
	return RuntimeConfig{}, nil
}

func (runtime *Runtime) GetDefaultFeatures(ctx context.Context) (agent.RuntimeFeatures, error) {
	return agent.RuntimeFeatures{
		EnableWebSearch:     nil,
		EnableNetworkAccess: nil,
	}, nil
}

func (runtime *Runtime) GetInfo(ctx context.Context) (agent.RuntimeInfo, error) {
	if err := ctx.Err(); err != nil {
		return agent.RuntimeInfo{}, err
	}

	path, err := locateExecutable(ctx)
	if err != nil {
		return agent.RuntimeInfo{}, err
	}

	output, err := exec.CommandContext(ctx, path, "--version").CombinedOutput()
	if err != nil {
		return agent.RuntimeInfo{}, fmt.Errorf("read claude version: %w", err)
	}

	version := semverPattern.FindString(string(output))
	if version == "" {
		return agent.RuntimeInfo{}, fmt.Errorf("parse claude version from output: %s", strings.TrimSpace(string(output)))
	}

	return agent.RuntimeInfo{Version: version}, nil
}

func (runtime *Runtime) AddMCPServer(ctx context.Context, mcpServerName agent.RuntimeMCPServerName, mcpServer agent.RuntimeMCPServer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	config, err := readClaudeConfig()
	if err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	claudeServer, err := toClaudeMCPServerConfig(mcpServer)
	if err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	config.MCPServers[string(mcpServerName)] = claudeServer

	if err := writeClaudeConfig(config); err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	return nil
}

func (runtime *Runtime) ListMCPServers(ctx context.Context) (map[agent.RuntimeMCPServerName]agent.RuntimeMCPServer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	config, err := readClaudeConfig()
	if err != nil {
		return nil, fmt.Errorf("list mcp servers: %w", err)
	}

	result := make(map[agent.RuntimeMCPServerName]agent.RuntimeMCPServer)

	for name, claudeServer := range config.MCPServers {
		runtimeServer, err := toRuntimeMCPServer(claudeServer)
		if err != nil {
			return nil, fmt.Errorf("list mcp servers: convert server %s: %w", name, err)
		}
		result[agent.RuntimeMCPServerName(name)] = runtimeServer
	}

	return result, nil
}

func (runtime *Runtime) RemoveMCPServer(ctx context.Context, mcpServerName agent.RuntimeMCPServerName) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	config, err := readClaudeConfig()
	if err != nil {
		return fmt.Errorf("remove mcp server: %w", err)
	}

	if _, exists := config.MCPServers[string(mcpServerName)]; !exists {
		return fmt.Errorf("remove mcp server: %w", ErrMCPServerNotFound)
	}

	delete(config.MCPServers, string(mcpServerName))

	if err := writeClaudeConfig(config); err != nil {
		return fmt.Errorf("remove mcp server: %w", err)
	}

	return nil
}
