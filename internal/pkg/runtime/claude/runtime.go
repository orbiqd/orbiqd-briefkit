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
	"github.com/spf13/afero"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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

	// #nosec G204 - path comes from locateExecutable which validates the executable
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

	// Only stdio servers supported
	if mcpServer.STDIO == nil {
		return errors.New("add mcp server: only STDIO servers are supported")
	}

	fs := afero.NewOsFs()

	// Read current config
	data, err := readClaudeConfig(fs)
	if err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	// Use sjson to add server fields
	serverPath := fmt.Sprintf("mcpServers.%s", mcpServerName)

	data, err = sjson.SetBytes(data, serverPath+".type", "stdio")
	if err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	data, err = sjson.SetBytes(data, serverPath+".command", mcpServer.STDIO.Command)
	if err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	if len(mcpServer.STDIO.Arguments) > 0 {
		data, err = sjson.SetBytes(data, serverPath+".args", mcpServer.STDIO.Arguments)
		if err != nil {
			return fmt.Errorf("add mcp server: %w", err)
		}
	}

	// Write back
	if err := writeClaudeConfig(fs, data); err != nil {
		return fmt.Errorf("add mcp server: %w", err)
	}

	return nil
}

func (runtime *Runtime) ListMCPServers(ctx context.Context) (map[agent.RuntimeMCPServerName]agent.RuntimeMCPServer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fs := afero.NewOsFs()
	data, err := readClaudeConfig(fs)
	if err != nil {
		return nil, fmt.Errorf("list mcp servers: %w", err)
	}

	result := make(map[agent.RuntimeMCPServerName]agent.RuntimeMCPServer)

	// Use gjson to iterate mcpServers
	mcpServersJSON := gjson.GetBytes(data, "mcpServers")
	if !mcpServersJSON.Exists() {
		return result, nil
	}

	mcpServersJSON.ForEach(func(serverName, serverData gjson.Result) bool {
		serverType := serverData.Get("type").String()

		// CRITICAL: Skip non-stdio servers (SSE, websocket)
		if serverType != "stdio" {
			return true // continue
		}

		// Build RuntimeMCPServer directly from gjson
		args := []string{}
		argsResult := serverData.Get("args")
		if argsResult.Exists() && argsResult.IsArray() {
			argsResult.ForEach(func(_, arg gjson.Result) bool {
				args = append(args, arg.String())
				return true
			})
		}

		runtimeServer := agent.RuntimeMCPServer{
			STDIO: &agent.RuntimeSTDIOMCPServer{
				Command:   serverData.Get("command").String(),
				Arguments: args,
			},
		}

		result[agent.RuntimeMCPServerName(serverName.String())] = runtimeServer
		return true // continue
	})

	return result, nil
}

func (runtime *Runtime) RemoveMCPServer(ctx context.Context, mcpServerName agent.RuntimeMCPServerName) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	fs := afero.NewOsFs()

	// Read config
	data, err := readClaudeConfig(fs)
	if err != nil {
		return fmt.Errorf("remove mcp server: %w", err)
	}

	// Check if server exists
	serverPath := fmt.Sprintf("mcpServers.%s", mcpServerName)
	if !gjson.GetBytes(data, serverPath).Exists() {
		return fmt.Errorf("remove mcp server: %w", ErrMCPServerNotFound)
	}

	// Delete using sjson
	data, err = sjson.DeleteBytes(data, serverPath)
	if err != nil {
		return fmt.Errorf("remove mcp server: %w", err)
	}

	// Write back
	if err := writeClaudeConfig(fs, data); err != nil {
		return fmt.Errorf("remove mcp server: %w", err)
	}

	return nil
}
