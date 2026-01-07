package briefkitctl

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

type AgentDiscoveryCmd struct {
	RuntimeKind        []agent.RuntimeKind `help:"Discovery specific runtime kind only."`
	WriteDefaultConfig bool                `help:"Write default config."`
}

func (command *AgentDiscoveryCmd) Run(ctx context.Context, runtimeRegistry agent.RuntimeRegistry, configRepository agent.ConfigRepository) error {
	kinds, err := runtimeRegistry.List(ctx)
	if err != nil {
		return fmt.Errorf("list runtimes: %w", err)
	}

	if len(command.RuntimeKind) > 0 {
		kinds = slices.DeleteFunc(kinds, func(runtimeKind agent.RuntimeKind) bool {
			return !slices.Contains(command.RuntimeKind, runtimeKind)
		})
	}

	for _, runtimeKind := range kinds {
		runtime, err := runtimeRegistry.Get(ctx, runtimeKind)
		if err != nil {
			return fmt.Errorf("get runtime %s: %w", runtimeKind, err)
		}

		runtimeFound, err := runtime.Discovery(ctx)
		if err != nil {
			return fmt.Errorf("discover runtime %s: %w", runtimeKind, err)
		}
		if !runtimeFound {
			continue
		}

		runtimeInfo, err := runtime.GetInfo(ctx)
		if err != nil {
			return fmt.Errorf("get runtime info %s: %w", runtimeKind, err)
		}

		slog.Info("Agent runtime discovered.",
			slog.String("runtimeKind", string(runtimeKind)),
			slog.String("runtimeVersion", runtimeInfo.Version),
		)

		if command.WriteDefaultConfig {
			agentConfig := agent.Config{}

			agentConfig.Runtime.Config, err = runtime.GetDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("get runtime default config: %w", err)
			}

			agentConfig.Runtime.Feature, err = runtime.GetDefaultFeatures(ctx)
			if err != nil {
				return fmt.Errorf("get runtime default features: %w", err)
			}

			agentConfig.Runtime.Kind = runtimeKind

			agentId := agent.AgentID(runtimeKind)

			err = configRepository.Update(ctx, agentId, agentConfig)
			if err != nil {
				return fmt.Errorf("update agent config %s: %w", agentId, err)
			}

			slog.Info("Default agent configuration saved.", slog.String("agentId", string(agentId)))
		}
	}

	return nil
}
