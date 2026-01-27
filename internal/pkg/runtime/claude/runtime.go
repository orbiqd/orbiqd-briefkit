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
	return agent.RuntimeFeatures{}, nil
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

func (runtime *Runtime) RegisterMCPServer(ctx context.Context, serverName agent.RuntimeMCPServerName, server agent.RuntimeMCPServer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	name := strings.TrimSpace(string(serverName))
	if name == "" {
		return errors.New("missing mcp server name")
	}

	if server.STDIO == nil {
		return errors.New("missing mcp server stdio configuration")
	}

	command := strings.TrimSpace(server.STDIO.Command)
	if command == "" {
		return errors.New("missing mcp server stdio command")
	}

	path, err := locateExecutable(ctx)
	if err != nil {
		return err
	}

	// #nosec G204 - path comes from locateExecutable which validates the executable
	removeOutput, removeErr := exec.CommandContext(ctx, path, "mcp", "remove", "--scope", "user", name).CombinedOutput()
	if removeErr != nil {
		outputStr := strings.ToLower(strings.TrimSpace(string(removeOutput)))
		if !strings.Contains(outputStr, "no mcp server found") {
			return fmt.Errorf("claude mcp server removal: %s", strings.TrimSpace(string(removeOutput)))
		}
	}

	addArgs := []string{"mcp", "add", "--scope", "user", name, command}
	addArgs = append(addArgs, server.STDIO.Arguments...)

	// #nosec G204 - path comes from locateExecutable which validates the executable
	addOutput, addErr := exec.CommandContext(ctx, path, addArgs...).CombinedOutput()
	if addErr != nil {
		trimmedOutput := strings.TrimSpace(string(addOutput))
		if trimmedOutput == "" {
			return fmt.Errorf("claude mcp server registration: %w", addErr)
		}
		return fmt.Errorf("claude mcp server registration: %s", trimmedOutput)
	}

	return nil
}
