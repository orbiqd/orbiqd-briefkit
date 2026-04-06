package briefkitctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

// StateExecutionCreateCmd creates a new execution.
type StateExecutionCreateCmd struct {
	AgentID   string  `required:"" help:"Agent ID"`
	Prompt    string  `arg:"" required:"" help:"Question or prompt"`
	Workspace *string `short:"w" help:"Workspace URI for isolated execution (e.g. dir:///path/to/project)."`
	Timeout   string  `short:"t" default:"5m" help:"Execution timeout"`
}

type executionCreateOutput struct {
	ID briefkit.ExecutionID `json:"id"`
}

func (e *StateExecutionCreateCmd) Run(ctx context.Context, repository briefkit.ExecutionRepository, configRepository briefkit.ConfigRepository) error {
	config, err := configRepository.Get(ctx, briefkit.AgentID(e.AgentID))
	if err != nil {
		return fmt.Errorf("load agent config: %w", err)
	}

	timeout, err := time.ParseDuration(e.Timeout)
	if err != nil {
		return fmt.Errorf("parse timeout: %w", err)
	}

	input := briefkit.ExecutionInput{
		Workspace: e.Workspace,
		Timeout:   utils.Duration(timeout),
		Prompt:    e.Prompt,
	}

	id, err := repository.Create(ctx, input, config)
	if err != nil {
		return fmt.Errorf("create execution: %w", err)
	}

	output := executionCreateOutput{ID: id}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("encode execution create output: %w", err)
	}

	return nil
}
