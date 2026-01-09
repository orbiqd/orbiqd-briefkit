package briefkit_runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
)

type RunnerCommand struct {
	Log   cli.LogConfig   `embed:"" prefix:"log-"`
	Store cli.StoreConfig `embed:"" prefix:"store-"`

	ExecutionID agent.ExecutionID `arg:"" required:"" help:"Execution ID to run."`
	Retry       bool              `help:"Allow rerunning finished executions."`
}

func (command *RunnerCommand) Run(ctx context.Context, executionRepository agent.ExecutionRepository, runtimeRegistry agent.RuntimeRegistry) error {
	slog.Info("Starting BriefKIT agent runner.", slog.String("executionID", string(command.ExecutionID)))

	execution, err := executionRepository.Get(ctx, command.ExecutionID)
	if err != nil {
		return fmt.Errorf("get execution: %w", err)
	}
	slog.Debug("Found execution.", slog.String("executionID", string(command.ExecutionID)))

	executionInput, err := execution.GetInput(ctx)
	if err != nil {
		return fmt.Errorf("get execution input: %w", err)
	}

	executionStatus, err := execution.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("get execution status: %w", err)
	}
	slog.Debug("Got execution status.", slog.String("executionID", string(command.ExecutionID)), slog.String("executionState", string(executionStatus.State)))

	if executionStatus.State != agent.ExecutionCreated {
		if !command.Retry {
			return fmt.Errorf("execution state is %s", executionStatus.State)
		}

		if executionStatus.State != agent.ExecutionFailed && executionStatus.State != agent.ExecutionSucceeded {
			return errors.New("execution state must be created, failed, or succeeded to retry")
		}

		slog.Info("Retrying execution.", slog.String("executionID", string(command.ExecutionID)), slog.String("executionState", string(executionStatus.State)))
	}

	agentConfig, err := execution.GetAgentConfig(ctx)
	if err != nil {
		return fmt.Errorf("get agent config: %w", err)
	}
	slog.Debug("Got agent config.", slog.String("executionID", string(command.ExecutionID)), slog.String("runtimeKind", string(agentConfig.Runtime.Kind)))

	runtime, err := runtimeRegistry.Get(ctx, agentConfig.Runtime.Kind)
	if err != nil {
		return fmt.Errorf("get runtime: %w", err)
	}
	slog.Debug("Got runtime.", slog.String("executionID", string(command.ExecutionID)), slog.String("runtimeKind", string(agentConfig.Runtime.Kind)))

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(executionInput.Timeout))
	defer cancel()

	executionStatus.State = agent.ExecutionStarted
	executionStatus.Attempts++
	executionStatus.Error = nil
	executionStatus.ExitCode = nil
	if err := execution.UpdateStatus(ctx, executionStatus); err != nil {
		return fmt.Errorf("update execution status: %w", err)
	}

	instance, err := runtime.Execute(runCtx, command.ExecutionID, executionInput, agentConfig)
	if err != nil {
		if updateErr := command.finishExecutionWithError(ctx, execution, executionStatus, err); updateErr != nil {
			return updateErr
		}
		return fmt.Errorf("execute runtime: %w", err)
	}

	executionStatus.State = agent.ExecutionRunning
	if err := execution.UpdateStatus(ctx, executionStatus); err != nil {
		return fmt.Errorf("update execution status: %w", err)
	}

	events := instance.Events()
	go command.drainRuntimeEvents(runCtx, events)

	result, err := instance.Wait(runCtx)
	if err != nil {
		if updateErr := command.finishExecutionWithError(ctx, execution, executionStatus, err); updateErr != nil {
			return updateErr
		}

		return fmt.Errorf("wait for runtime: %w", err)
	}

	executionResult := agent.ExecutionResult{
		Response:       result.Response,
		ConversationID: result.ConversationID,
	}
	if err := execution.SetResult(ctx, executionResult); err != nil {
		return fmt.Errorf("set execution result: %w", err)
	}

	slog.Info("Execution succeeded.")

	return nil
}

func (command *RunnerCommand) finishExecutionWithError(ctx context.Context, execution agent.Execution, status agent.ExecutionStatus, err error) error {
	now := time.Now()
	status.State = agent.ExecutionFailed
	status.FinishedAt = &now

	message := err.Error()
	status.Error = &message
	status.ExitCode = nil

	var runtimeErr *agent.RuntimeExecutionError
	if errors.As(err, &runtimeErr) {
		status.ExitCode = runtimeErr.ExitCode
	}

	if updateErr := execution.UpdateStatus(ctx, status); updateErr != nil {
		return fmt.Errorf("update execution status: %w", updateErr)
	}

	return nil
}

func (command *RunnerCommand) drainRuntimeEvents(ctx context.Context, events <-chan agent.RuntimeEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-events:
			if !ok {
				return
			}
		}
	}
}
