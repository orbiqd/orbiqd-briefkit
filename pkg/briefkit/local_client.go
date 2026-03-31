package briefkit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

type LocalClientOpt func(*LocalClient)

type LocalClient struct {
	runner              Runner
	executionRepository ExecutionRepository
	configRepository    ConfigRepository

	defaultExecutionTimeout time.Duration
}

func WithLocalClientDefaultExecutionTimeout(timeout time.Duration) LocalClientOpt {
	return func(client *LocalClient) {
		client.defaultExecutionTimeout = timeout
	}
}

func NewLocalClient(runner Runner, executionRepository ExecutionRepository, configRepository ConfigRepository) *LocalClient {
	return &LocalClient{
		runner:              runner,
		executionRepository: executionRepository,
		configRepository:    configRepository,

		defaultExecutionTimeout: 5 * time.Minute,
	}
}

func (client *LocalClient) Ask(ctx context.Context, agentID AgentID, prompt string, opts ...AskOption) (*AskResult, error) {
	options := AskOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	agentExists, err := client.configRepository.Exists(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("agent config exists check: %w", err)
	}

	if !agentExists {
		return nil, fmt.Errorf("agent config does not exist: %s", agentID)
	}

	agentConfig, err := client.configRepository.Get(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("get agent config: %w", err)
	}

	slog.Debug("Found agent config.", slog.String("runtimeKind", string(agentConfig.Runtime.Kind)))

	executionTimeout := client.defaultExecutionTimeout
	if agentConfig.Timeout != nil && time.Duration(*agentConfig.Timeout) > 0 {
		executionTimeout = time.Duration(*agentConfig.Timeout)
	}
	if options.Timeout != nil {
		executionTimeout = *options.Timeout
	}

	executionInput := ExecutionInput{
		WorkingDirectory: nil,
		Timeout:          utils.Duration(executionTimeout),
		Prompt:           prompt,
		Model:            options.Model,
		ConversationID:   options.ConversationID,
	}

	executionId, err := client.executionRepository.Create(ctx, executionInput, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("create execution: %w", err)
	}

	slog.Info("Created execution.", slog.String("executionId", string(executionId)))

	if err = client.runner.Spawn(ctx, executionId); err != nil {
		return nil, fmt.Errorf("spawn runner: %w", err)
	}

	if err = client.runner.Wait(ctx, executionId); err != nil {
		return nil, fmt.Errorf("wait for runner: %w", err)
	}

	executionHandle, err := client.executionRepository.Get(ctx, executionId)
	if err != nil {
		return nil, fmt.Errorf("get execution handle: %w", err)
	}

	executionResult, err := executionHandle.GetResult(ctx)
	if err != nil {
		return nil, fmt.Errorf("get execution result: %w", err)
	}

	return &AskResult{
		Response:       executionResult.Response,
		ConversationID: executionResult.ConversationID,
	}, nil
}
