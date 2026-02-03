//go:build !coverage

package briefkit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLocalClient_Ask_WhenConfigExistsCheckFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, _, _, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	expectedErr := errors.New("exists failed")

	configRepository.EXPECT().Exists(ctx, agentID).Return(false, expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "agent config exists check")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenConfigDoesNotExist_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, _, _, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"

	configRepository.EXPECT().Exists(ctx, agentID).Return(false, nil)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.EqualError(t, err, "agent config does not exist: "+string(agentID))
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenConfigGetFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, _, _, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	expectedErr := errors.New("get failed")

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(Config{}, expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "get agent config")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenCreateExecutionFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, _, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	expectedErr := errors.New("create failed")

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).Return(EmptyExecutionID, expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "create execution")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenRunnerSpawnFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	expectedErr := errors.New("spawn failed")

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "spawn runner")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenRunnerWaitFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	expectedErr := errors.New("wait failed")

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(nil)
	runner.EXPECT().Wait(ctx, executionID).Return(expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "wait for runner")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenExecutionHandleGetFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	expectedErr := errors.New("get handle failed")

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(nil)
	runner.EXPECT().Wait(ctx, executionID).Return(nil)
	executionRepository.EXPECT().Get(ctx, executionID).Return(nil, expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "get execution handle")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenGetResultFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	expectedErr := errors.New("result failed")
	execution := NewMockExecution(t)

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(nil)
	runner.EXPECT().Wait(ctx, executionID).Return(nil)
	executionRepository.EXPECT().Get(ctx, executionID).Return(execution, nil)
	execution.EXPECT().GetResult(ctx).Return(ExecutionResult{}, expectedErr)

	result, err := client.Ask(ctx, agentID, prompt)

	require.Error(t, err)
	require.ErrorIs(t, err, expectedErr)
	require.ErrorContains(t, err, "get execution result")
	assert.Nil(t, result)
}

func TestLocalClient_Ask_WhenSuccessWithDefaultOptions_ThenReturnsResultAndUsesDefaultTimeout(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	result := ExecutionResult{Response: "done", ConversationID: ConversationID("conv-1")}
	execution := NewMockExecution(t)
	var capturedInput ExecutionInput

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).
		Run(func(ctx context.Context, input ExecutionInput, agentConfig Config) {
			capturedInput = input
		}).
		Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(nil)
	runner.EXPECT().Wait(ctx, executionID).Return(nil)
	executionRepository.EXPECT().Get(ctx, executionID).Return(execution, nil)
	execution.EXPECT().GetResult(ctx).Return(result, nil)

	askResult, err := client.Ask(ctx, agentID, prompt)

	require.NoError(t, err)
	require.NotNil(t, askResult)
	assert.Equal(t, result.Response, askResult.Response)
	assert.Equal(t, result.ConversationID, askResult.ConversationID)
	assert.Equal(t, prompt, capturedInput.Prompt)
	assert.Nil(t, capturedInput.WorkingDirectory)
	assert.Nil(t, capturedInput.Model)
	assert.Nil(t, capturedInput.ConversationID)
	assert.Equal(t, 5*time.Minute, time.Duration(capturedInput.Timeout))
}

func TestLocalClient_Ask_WhenSuccessWithOptions_ThenUsesProvidedModelConversationAndTimeout(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	result := ExecutionResult{Response: "done", ConversationID: ConversationID("conv-1")}
	execution := NewMockExecution(t)
	model := "gpt-4"
	conversationID := ConversationID("conv-2")
	timeout := 2 * time.Minute
	var capturedInput ExecutionInput

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).
		Run(func(ctx context.Context, input ExecutionInput, agentConfig Config) {
			capturedInput = input
		}).
		Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(nil)
	runner.EXPECT().Wait(ctx, executionID).Return(nil)
	executionRepository.EXPECT().Get(ctx, executionID).Return(execution, nil)
	execution.EXPECT().GetResult(ctx).Return(result, nil)

	askResult, err := client.Ask(
		ctx,
		agentID,
		prompt,
		AskWithModel(model),
		AskWithConversationID(conversationID),
		AskWithTimeout(timeout),
	)

	require.NoError(t, err)
	require.NotNil(t, askResult)
	assert.Equal(t, result.Response, askResult.Response)
	assert.Equal(t, result.ConversationID, askResult.ConversationID)
	assert.Equal(t, prompt, capturedInput.Prompt)
	require.NotNil(t, capturedInput.Model)
	assert.Equal(t, model, *capturedInput.Model)
	require.NotNil(t, capturedInput.ConversationID)
	assert.Equal(t, conversationID, *capturedInput.ConversationID)
	assert.Equal(t, timeout, time.Duration(capturedInput.Timeout))
}

func TestLocalClientDefaultExecutionTimeout_WhenAsk_ThenUsesCustomTimeout(t *testing.T) {
	ctx := context.Background()
	client, runner, executionRepository, configRepository := newLocalClientFixture(t)
	agentID := AgentID("codex")
	prompt := "hello"
	config := testConfig()
	executionID := ExecutionID("execution-1")
	result := ExecutionResult{Response: "done", ConversationID: ConversationID("conv-1")}
	execution := NewMockExecution(t)
	customTimeout := 2 * time.Minute
	var capturedInput ExecutionInput

	WithLocalClientDefaultExecutionTimeout(customTimeout)(client)

	configRepository.EXPECT().Exists(ctx, agentID).Return(true, nil)
	configRepository.EXPECT().Get(ctx, agentID).Return(config, nil)
	executionRepository.EXPECT().Create(ctx, mock.Anything, config).
		Run(func(ctx context.Context, input ExecutionInput, agentConfig Config) {
			capturedInput = input
		}).
		Return(executionID, nil)
	runner.EXPECT().Spawn(ctx, executionID).Return(nil)
	runner.EXPECT().Wait(ctx, executionID).Return(nil)
	executionRepository.EXPECT().Get(ctx, executionID).Return(execution, nil)
	execution.EXPECT().GetResult(ctx).Return(result, nil)

	askResult, err := client.Ask(ctx, agentID, prompt)

	require.NoError(t, err)
	require.NotNil(t, askResult)
	assert.Equal(t, customTimeout, time.Duration(capturedInput.Timeout))
}

func newLocalClientFixture(t *testing.T) (*LocalClient, *MockRunner, *MockExecutionRepository, *MockConfigRepository) {
	t.Helper()

	runner := NewMockRunner(t)
	executionRepository := NewMockExecutionRepository(t)
	configRepository := NewMockConfigRepository(t)

	return NewLocalClient(runner, executionRepository, configRepository), runner, executionRepository, configRepository
}

func testConfig() Config {
	var config Config
	config.Runtime.Kind = RuntimeKind("codex")
	return config
}
