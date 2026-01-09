package briefkitctl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
)

type SetupCmd struct {
	RuntimeKind []agent.RuntimeKind `help:"Limit setup to these runtime kinds. When empty, all kinds will be configured."`

	SetupAgentConfig bool `default:"true" help:"Configure agent runtime settings (default: true)"`
	SetupAgentMCP    bool `default:"true" help:"Configure agent MCP server integration (default: true)"`

	EnableSandbox *bool `help:"If set, override default agent configuration and enable/disable sandbox for all tasks"`

	Force bool `default:"false" help:"Allow overriding existing MCP server or agent configuration"`
}

func (command *SetupCmd) Run(ctx context.Context, runtimeRegistry agent.RuntimeRegistry, configRepository agent.ConfigRepository) error {
	runtimes, err := command.detectRuntimes(ctx, runtimeRegistry)
	if err != nil {
		return fmt.Errorf("detect runtimes: %w", err)
	}

	if len(command.RuntimeKind) > 0 {
		maps.DeleteFunc(runtimes, func(runtimeKind agent.RuntimeKind, _ agent.Runtime) bool {
			return !slices.Contains(command.RuntimeKind, runtimeKind)
		})
	}

	for runtimeKind, runtime := range runtimes {
		err = command.setupRuntime(ctx, runtimeKind, runtime, configRepository)
		if err != nil {
			return fmt.Errorf("setup runtime: %w", err)
		}
	}

	return nil
}

func (command *SetupCmd) detectRuntimes(ctx context.Context, runtimeRegistry agent.RuntimeRegistry) (map[agent.RuntimeKind]agent.Runtime, error) {
	supportedRuntimes, err := runtimeRegistry.List(ctx)
	if err != nil {
		return map[agent.RuntimeKind]agent.Runtime{}, fmt.Errorf("list runtime: %w", err)
	}

	runtimes := map[agent.RuntimeKind]agent.Runtime{}

	for _, runtimeKind := range supportedRuntimes {
		logger := slog.With(slog.String("runtimeKind", string(runtimeKind)))

		logger.Debug("Discovering runtime on system.")

		runtime, err := runtimeRegistry.Get(ctx, runtimeKind)
		if err != nil {
			return map[agent.RuntimeKind]agent.Runtime{}, fmt.Errorf("get %s runtime: %w", runtimeKind, err)
		}

		runtimeFound, err := runtime.Discovery(ctx)
		if err != nil {
			return map[agent.RuntimeKind]agent.Runtime{}, fmt.Errorf("discover %s runtime: %w", runtimeKind, err)
		}

		if !runtimeFound {
			logger.Warn("Runtime not found on system.")
			continue
		}

		runtimeInfo, err := runtime.GetInfo(ctx)
		if err != nil {
			return map[agent.RuntimeKind]agent.Runtime{}, fmt.Errorf("get %s runtime info: %w", runtimeKind, err)
		}

		logger.Info("Runtime found.", slog.String("runtimeVersion", runtimeInfo.Version))

		runtimes[runtimeKind] = runtime
	}

	return runtimes, nil
}

func (command *SetupCmd) setupRuntime(ctx context.Context, runtimeKind agent.RuntimeKind, runtime agent.Runtime, configRepository agent.ConfigRepository) error {
	logger := slog.With(slog.String("runtimeKind", string(runtimeKind)))

	if command.SetupAgentConfig {
		err := command.setupRuntimeAgentConfig(ctx, runtimeKind, runtime, configRepository)
		if err != nil {
			return fmt.Errorf("agent config: %w", err)
		}
	} else {
		logger.Warn("Skipping agent configuration setup.")
	}

	if command.SetupAgentMCP {
		err := command.setupRuntimeAgentMCP(ctx, runtimeKind, runtime)
		if err != nil {
			return fmt.Errorf("agent mcp: %w", err)
		}
	} else {
		logger.Warn("Skipping agent MCP server setup.")
	}

	return nil
}

func (command *SetupCmd) setupRuntimeAgentConfig(ctx context.Context, runtimeKind agent.RuntimeKind, runtime agent.Runtime, configRepository agent.ConfigRepository) error {
	agentId := agent.AgentID(runtimeKind)

	logger := slog.With(slog.String("runtimeKind", string(runtimeKind)), slog.String("agentId", string(agentId)))

	logger.Debug("Configuring agent runtime settings.")

	runtimeConfig, err := runtime.GetDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("get default config: %w", err)
	}

	runtimeFeatures, err := runtime.GetDefaultFeatures(ctx)
	if err != nil {
		return errors.New("get default features")
	}

	if command.EnableSandbox != nil {
		if *command.EnableSandbox {
			logger.Info("Enabling sandbox for runtime.")
		} else {
			logger.Warn("Disabling sandbox for runtime.")
		}

		runtimeFeatures.EnableSandbox = command.EnableSandbox
	}

	config := agent.Config{}
	config.Runtime.Kind = runtimeKind
	config.Runtime.Config = runtimeConfig
	config.Runtime.Feature = runtimeFeatures

	hasConfig, err := configRepository.Exists(ctx, agentId)
	if err != nil {
		return fmt.Errorf("agent config exists: %w", err)
	}
	if hasConfig && !command.Force {
		return fmt.Errorf("agent config %s already exists", agentId)
	}

	err = configRepository.Update(ctx, agentId, config)
	if err != nil {
		return fmt.Errorf("update agent config: %w", err)
	}

	logger.Info("Agent configuration saved successfully.")

	return nil
}

func (command *SetupCmd) setupRuntimeAgentMCP(ctx context.Context, runtimeKind agent.RuntimeKind, runtime agent.Runtime) error {
	mcpServerName := agent.RuntimeMCPServerName("briefkit")

	logger := slog.With(
		slog.String("runtimeKind", string(runtimeKind)),
		slog.String("mcpServerName", string(mcpServerName)),
	)

	logger.Debug("Configuring agent MCP server.")

	mcpServers, err := runtime.ListMCPServers(ctx)
	if err != nil {
		return fmt.Errorf("list mcp servers: %w", err)
	}

	_, hasMcpServer := mcpServers[mcpServerName]
	if hasMcpServer && !command.Force {
		return fmt.Errorf("%s mcp server already exists in %s runtime", mcpServerName, runtimeKind)
	}

	if hasMcpServer {
		err = runtime.RemoveMCPServer(ctx, mcpServerName)
		if err != nil {
			return fmt.Errorf("remove %s mcp server from %s: %w", mcpServerName, runtimeKind, err)
		}
	}

	executablePath, err := cli.ResolveExecutable(ctx, cli.ExecutableMCP)
	if err != nil {
		return fmt.Errorf("resolve mcp-server executable: %w", err)
	}

	mcpServer := agent.RuntimeMCPServer{
		STDIO: &agent.RuntimeSTDIOMCPServer{
			Command: executablePath,
		},
	}

	err = runtime.AddMCPServer(ctx, mcpServerName, mcpServer)
	if err != nil {
		return fmt.Errorf("add %s mcp server to %s: %w", mcpServerName, runtimeKind, err)
	}

	return nil
}
