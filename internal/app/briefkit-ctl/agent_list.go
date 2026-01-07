package briefkitctl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

// AgentListOutputItem represents a single agent entry in the list output.
type AgentListOutputItem struct {
	ID          agent.AgentID     `json:"id"`
	RuntimeKind agent.RuntimeKind `json:"runtimeKind"`
}

// AgentListOutput captures the list output payload for agents.
type AgentListOutput struct {
	Items []AgentListOutputItem `json:"items"`
	Count int                   `json:"count"`
}

// AgentListCmd lists configured agents and their runtime kinds.
type AgentListCmd struct{}

// Run executes the agent list command.
func (a *AgentListCmd) Run(ctx context.Context, repository agent.ConfigRepository) error {
	ids, err := repository.List(ctx)
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	items := make([]AgentListOutputItem, 0, len(ids))
	for _, id := range ids {
		config, err := repository.Get(ctx, id)
		if err != nil {
			slog.Warn(
				"Failed to load agent config",
				slog.String("id", string(id)),
				slog.String("error", err.Error()),
			)
			continue
		}

		items = append(items, AgentListOutputItem{
			ID:          id,
			RuntimeKind: config.Runtime.Kind,
		})
	}

	output := AgentListOutput{
		Items: items,
		Count: len(items),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encode agent list output: %w", err)
	}

	return nil
}
