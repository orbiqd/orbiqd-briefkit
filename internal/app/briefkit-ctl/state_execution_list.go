package briefkitctl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

// ExecutionListOutputItem represents a single execution entry in the list output.
type ExecutionListOutputItem struct {
	Id     briefkit.ExecutionID     `json:"id"`
	Status briefkit.ExecutionStatus `json:"status"`
}

// ExecutionListOutput captures the list output payload for executions.
type ExecutionListOutput struct {
	Items []ExecutionListOutputItem `json:"items"`
	Count int                       `json:"count"`
}

// StateExecutionListCmd lists stored executions and their status.
type StateExecutionListCmd struct{}

func (e *StateExecutionListCmd) Run(ctx context.Context, repository briefkit.ExecutionRepository) error {
	ids, err := repository.Find(ctx)
	if err != nil {
		return fmt.Errorf("list executions: %w", err)
	}

	items := make([]ExecutionListOutputItem, 0, len(ids))
	for _, id := range ids {
		execution, err := repository.Get(ctx, id)
		if err != nil {
			slog.Warn(
				"Failed to load execution",
				slog.String("id", string(id)),
				slog.String("error", err.Error()),
			)
			continue
		}

		status, err := execution.GetStatus(ctx)
		if err != nil {
			slog.Warn(
				"Failed to load execution status",
				slog.String("id", string(id)),
				slog.String("error", err.Error()),
			)
			continue
		}

		items = append(items, ExecutionListOutputItem{
			Id:     id,
			Status: status,
		})
	}

	output := ExecutionListOutput{
		Items: items,
		Count: len(items),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encode execution list output: %w", err)
	}

	return nil
}
