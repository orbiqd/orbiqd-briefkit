package briefkitctl

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
)

// TODO: Refine log texts to be more clear and well polished
// TODO: Refine help texts to be more descriptive and clear for users

type SetupCmd struct {
	RuntimeKind []agent.RuntimeKind `help:"limit setup to the this runtime kidds, when empty all kinds will be used"`

	SetupAgentConfig bool `default:"true" help:""`
	SetupAgentMCP    bool `default:"true" help:""`

	EnableSandbox *bool `help:"if setted, ovverride default agent config and enable sandbox for all tasks"`

	Force bool `default:"false" help:"allow to override existing mcp or agent config"`
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

		logger.Debug("Discovering runtime.")

		runtime, err := runtimeRegistry.Get(ctx, runtimeKind)
		if err != nil {
			return map[agent.RuntimeKind]agent.Runtime{}, fmt.Errorf("get %s runtime: %w", runtimeKind, err)
		}

		runtimeFound, err := runtime.Discovery(ctx)
		if err != nil {
			return map[agent.RuntimeKind]agent.Runtime{}, fmt.Errorf("discover %s runtime: %w", runtimeKind, err)
		}

		if !runtimeFound {
			slog.Warn("Runtime not found.")
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
		logger.Warn("Skippign setup agent configuratoin.")
	}

	if command.SetupAgentMCP {
		err := command.setupRuntimeAgentMCP(ctx, runtimeKind, runtime)
		if err != nil {
			return fmt.Errorf("agent mcp: %w", err)
		}
	} else {
		logger.Warn("Skipping setup agent MCP.")
	}

	return nil
}

func (command *SetupCmd) setupRuntimeAgentConfig(ctx context.Context, runtimeKind agent.RuntimeKind, runtime agent.Runtime, configRepository agent.ConfigRepository) error {
	agentId := agent.AgentID(runtimeKind)

	logger := slog.With(slog.String("runtimeKind", string(runtimeKind)), slog.String("agentId", string(agentId)))

	logger.Debug("Setting up agent config.")

	runtimeConfig, err := runtime.GetDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("get default config: %w", err)
	}

	runtimeFeatures, err := runtime.GetDefaultFeatures(ctx)
	if err != nil {
		return fmt.Errorf("get default features")
	}

	if command.EnableSandbox != nil {
		if *command.EnableSandbox {
			logger.Info("Enablging sandbox for runtime.")
		} else {
			logger.Warn("Overriding disabling sandbox for runtime.")
		}

		runtimeFeatures.EnableSandbox = command.EnableSandbox
	}

	config := agent.Config{}
	config.Runtime.Kind = runtimeKind
	config.Runtime.Config = runtimeConfig
	config.Runtime.Feature = runtimeFeatures

	hasConfig, err := configRepository.Exists(ctx, agentId)
	if hasConfig && !command.Force {
		return fmt.Errorf("agent config %s already exists", agentId)
	}

	err = configRepository.Update(ctx, agentId, config)
	if err != nil {
		return fmt.Errorf("update agent config: %w", err)
	}

	logger.Info("Agent config updated.")

	return nil
}

func (command *SetupCmd) setupRuntimeAgentMCP(ctx context.Context, runtimeKind agent.RuntimeKind, runtime agent.Runtime) error {
	mcpServerName := agent.RuntimeMCPServerName("briefkit")

	logger := slog.With(
		slog.String("runtimeKind", string(runtimeKind)),
		slog.String("mcpServerName", string(mcpServerName)),
	)

	logger.Debug("Setting up agent MCP.")

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
