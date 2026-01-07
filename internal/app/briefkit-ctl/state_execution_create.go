package briefkitctl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

// StateExecutionCreateCmd creates a new execution.
type StateExecutionCreateCmd struct {
	AgentID    string `required:"" help:"Agent ID"`
	Prompt     string `arg:"" required:"" help:"Question or prompt"`
	WorkingDir string `short:"w" default:"." help:"Working directory"`
	Timeout    string `short:"t" default:"5m" help:"Execution timeout"`
}

type executionCreateOutput struct {
	ID agent.ExecutionID `json:"id"`
}

func (e *StateExecutionCreateCmd) Run(ctx context.Context, repository agent.ExecutionRepository, configRepository agent.ConfigRepository) error {
	config, err := configRepository.Get(ctx, agent.AgentID(e.AgentID))
	if err != nil {
		return fmt.Errorf("load agent config: %w", err)
	}

	timeout, err := time.ParseDuration(e.Timeout)
	if err != nil {
		return fmt.Errorf("parse timeout: %w", err)
	}

	expandedWorkingDir, err := homedir.Expand(e.WorkingDir)
	if err != nil {
		return fmt.Errorf("expand working directory: %w", err)
	}

	workingDir, err := filepath.Abs(expandedWorkingDir)
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	input := agent.ExecutionInput{
		WorkingDirectory: &workingDir,
		Timeout:          utils.Duration(timeout),
		Prompt:           e.Prompt,
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
