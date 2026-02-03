package briefkitctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

// ExecutionShowOutput captures the output payload for execution show.
type ExecutionShowOutput struct {
	Status briefkit.ExecutionStatus  `json:"status"`
	Input  briefkit.ExecutionInput   `json:"input"`
	Result *briefkit.ExecutionResult `json:"result,omitempty"`
}

// StateExecutionShowCmd shows details for a single execution.
type StateExecutionShowCmd struct {
	ID string `arg:"" required:"" help:"Execution ID"`
}

// Run executes the execution show command.
func (e *StateExecutionShowCmd) Run(ctx context.Context, repository briefkit.ExecutionRepository) error {
	id := briefkit.ExecutionID(e.ID)
	if err := id.Validate(); err != nil {
		return fmt.Errorf("validate execution id: %w", err)
	}

	execution, err := repository.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("load execution: %w", err)
	}

	status, err := execution.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("load execution status: %w", err)
	}

	input, err := execution.GetInput(ctx)
	if err != nil {
		return fmt.Errorf("load execution input: %w", err)
	}

	var result *briefkit.ExecutionResult
	executionResult, err := execution.GetResult(ctx)
	if err != nil {
		if !errors.Is(err, briefkit.ErrExecutionNoResult) {
			return fmt.Errorf("load execution result: %w", err)
		}
	} else {
		result = &executionResult
	}

	output := ExecutionShowOutput{
		Status: status,
		Input:  input,
		Result: result,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encode execution show output: %w", err)
	}

	return nil
}
