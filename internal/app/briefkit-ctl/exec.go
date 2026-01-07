package briefkitctl

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	briefkitrunner "github.com/orbiqd/orbiqd-briefkit/internal/app/briefkit-runner"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

// ExecCmd runs a prompt with specified model and options.
type ExecCmd struct {
	AgentID        agent.AgentID         `help:"ID of the agent." required:"true"`
	Auto           bool                  `help:"Enable automatic mode"`
	Timeout        time.Duration         `default:"5m"`
	Model          *string               `help:"Select model for execution."`
	ConversationID *agent.ConversationID `help:"Conversation ID for execution."`

	Prompt string `arg:"" required:"" help:"Prompt to execute"`
}

func (command *ExecCmd) Run(ctx context.Context, executionRepository agent.ExecutionRepository, agentConfigRepository agent.ConfigRepository) error {
	agentExists, err := agentConfigRepository.Exists(ctx, command.AgentID)
	if err != nil {
		return fmt.Errorf("agent config exists: %w", err)
	}
	if !agentExists {
		return fmt.Errorf("agent config does not exist: %s", command.AgentID)
	}

	agentConfig, err := agentConfigRepository.Get(ctx, command.AgentID)
	if err != nil {
		return fmt.Errorf("get agent config: %w", err)
	}

	slog.Debug("Found agent config.", slog.String("runtimeKind", string(agentConfig.Runtime.Kind)))

	executionInput := agent.ExecutionInput{
		WorkingDirectory: nil,
		Timeout:          utils.Duration(command.Timeout),
		Prompt:           command.Prompt,
		Model:            command.Model,
		ConversationID:   command.ConversationID,
	}

	executionID, err := executionRepository.Create(ctx, executionInput, agentConfig)
	if err != nil {
		return fmt.Errorf("create execution: %w", err)
	}

	slog.Info("Created execution.", slog.String("executionId", string(executionID)))

	if err := briefkitrunner.Spawn(ctx, executionID); err != nil {
		return fmt.Errorf("spawn runner: %w", err)
	}

	// Use a separate context for the polling loop to respect the command timeout.
	// We add a buffer to ensure we can retrieve the final status even if the runner
	// uses the full execution timeout.
	pollCtx, cancel := context.WithTimeout(ctx, command.Timeout+30*time.Second)
	defer cancel()

	return command.waitForCompletion(pollCtx, executionRepository, executionID)
}

func (command *ExecCmd) waitForCompletion(ctx context.Context, repo agent.ExecutionRepository, id agent.ExecutionID) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Get the execution handle.
	executionHandle, err := repo.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("get execution handle: %w", err)
	}

	var lastState agent.ExecutionState

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for completion: %w", ctx.Err())
		case <-ticker.C:
			status, err := executionHandle.GetStatus(ctx)
			if err != nil {
				slog.Warn("Failed to get execution status", "error", err)
				continue
			}

			if status.State != lastState {
				slog.Info("Execution status changed.",
					slog.String("old", string(lastState)),
					slog.String("new", string(status.State)))
				lastState = status.State
			}

			if status.State.IsFinished() {
				if status.State == agent.ExecutionSucceeded {
					result, err := executionHandle.GetResult(ctx)
					if err != nil {
						return fmt.Errorf("get result: %w", err)
					}
					slog.Info("Execution finished successfully.", slog.String("conversationId", string(result.ConversationID)))
					fmt.Println()
					fmt.Println(result.Response)
					return nil
				}

				errMsg := "unknown error"
				if status.Error != nil {
					errMsg = *status.Error
				}
				return fmt.Errorf("execution failed: %s", errMsg)
			}
		}
	}
}
