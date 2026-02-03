package claude

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance_Wait_WhenSingleAssistantAndResult_ThenReturnsResultResponse(t *testing.T) {
	resetClaudeMockEnv(t)
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	result, err := waitForInstance(t, instance)

	require.NoError(t, err)
	assert.Equal(t, "Mock response to: hello", result.Response)
}

func TestInstance_Wait_WhenMultipleAssistantEvents_ThenConcatenatesResponse(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_MULTI_ASSISTANT", "1")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	result, err := waitForInstance(t, instance)

	require.NoError(t, err)
	assert.Equal(t, "First part of responseSecond part of responseThird part of response", result.Response)
}

func TestInstance_Wait_WhenEmptyLines_ThenSkipsWithoutError(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_EMPTY_LINES", "1")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	result, err := waitForInstance(t, instance)

	require.NoError(t, err)
	assert.Equal(t, "Mock response to: hello", result.Response)
}

func TestInstance_Wait_WhenSystemInitHasSessionID_ThenSetsConversationID(t *testing.T) {
	resetClaudeMockEnv(t)
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	result, err := waitForInstance(t, instance)

	require.NoError(t, err)
	assert.Equal(t, briefkit.ConversationID("mock-session-id-12345"), result.ConversationID)
}

func TestInstance_Wait_WhenNoResultEvent_ThenReturnsAssistantResponse(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_NO_RESULT", "1")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	result, err := waitForInstance(t, instance)

	require.NoError(t, err)
	assert.Equal(t, "Mock response to: hello", result.Response)
}

func TestInstance_Wait_WhenResultErrorEvent_ThenReturnsAssistantResponse(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_RESULT_ERROR", "1")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	result, err := waitForInstance(t, instance)

	require.NoError(t, err)
	assert.Equal(t, "Mock response to: hello", result.Response)
}

func TestInstance_Wait_WhenFailWithStderr_ThenReturnsRuntimeExecutionError(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_FAIL", "1")
	t.Setenv("MOCK_CLAUDE_STDERR", "boom")
	t.Setenv("MOCK_CLAUDE_EXIT_CODE", "7")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	_, err := waitForInstance(t, instance)

	require.Error(t, err)
	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Equal(t, "boom", runtimeErr.Message)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 7, *runtimeErr.ExitCode)
}

func TestInstance_Wait_WhenFailWithoutStderr_ThenReturnsRuntimeExecutionError(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_FAIL", "1")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	_, err := waitForInstance(t, instance)

	require.Error(t, err)
	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Contains(t, runtimeErr.Message, "exit status")
	require.NotNil(t, runtimeErr.ExitCode)
}

func TestInstance_Wait_WhenPartialFail_ThenReturnsRuntimeExecutionError(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_PARTIAL_FAIL", "1")
	t.Setenv("MOCK_CLAUDE_STDERR", "partial failure")
	t.Setenv("MOCK_CLAUDE_EXIT_CODE", "3")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	_, err := waitForInstance(t, instance)

	require.Error(t, err)
	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Equal(t, "partial failure", runtimeErr.Message)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 3, *runtimeErr.ExitCode)
}

func TestInstance_Wait_WhenSignal_ThenReturnsRuntimeExecutionErrorWithExitCode(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("MOCK_CLAUDE_SIGNAL", "SIGINT")
	instance := newTestClaudeInstance(t, briefkit.ExecutionInput{Prompt: "hello"})

	_, err := waitForInstance(t, instance)

	require.Error(t, err)
	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Contains(t, []int{-1, 130}, *runtimeErr.ExitCode)
}

func newTestClaudeInstance(t *testing.T, input briefkit.ExecutionInput) *Instance {
	t.Helper()
	setClaudeMockExecutable(t)

	instance, err := newInstance(
		context.Background(),
		briefkit.ExecutionID("exec-123"),
		input,
		RuntimeConfig{},
		briefkit.RuntimeFeatures{},
		t.TempDir(),
	)
	require.NoError(t, err)
	return instance
}

func waitForInstance(t *testing.T, instance *Instance) (briefkit.RuntimeResult, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := instance.Wait(ctx)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("wait timeout: %v", err)
	}

	return result, err
}
