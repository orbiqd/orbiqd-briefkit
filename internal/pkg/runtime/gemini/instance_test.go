package gemini

import (
	"context"
	"os"
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstance_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	resetGeminiMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := newInstance(ctx, briefkit.ExecutionID("exec-1"), briefkit.ExecutionInput{}, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestNewInstance_WhenWorkingDirectoryProvided_ThenUsesProvidedDir(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	workingDir := t.TempDir()
	input := briefkit.ExecutionInput{
		WorkingDirectory: &workingDir,
		Prompt:           "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, workingDir, instance.cmd.Dir)
	waitForInstance(t, instance)
}

func TestNewInstance_WhenWorkingDirectoryMissing_ThenUsesCurrentDir(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	expectedDir, err := os.Getwd()
	require.NoError(t, err)
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	assert.Equal(t, expectedDir, instance.cmd.Dir)
	waitForInstance(t, instance)
}

func TestInstanceRun_WhenGeminiEmitsInitAndMessage_ThenCapturesConversationAndResponse(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	conversationID := briefkit.ConversationID("session-123")
	input := briefkit.ExecutionInput{
		Prompt:         "Hello Gemini",
		ConversationID: &conversationID,
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.NoError(t, err)
	assert.Equal(t, conversationID, result.ConversationID)
	assert.Contains(t, result.Response, "Mock response to: Hello Gemini")
	assert.Contains(t, result.Response, "Resumed: session-123")
}

func TestInstanceRun_WhenGeminiEmitsMultipleMessages_ThenConcatenatesResponse(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_MULTI_MESSAGE", "1")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "First part of responseSecond part of responseThird part of response", result.Response)
}

func TestInstanceRun_WhenGeminiSkipsInit_ThenConversationIDIsEmpty(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_NO_INIT", "1")
	conversationID := briefkit.ConversationID("session-123")
	input := briefkit.ExecutionInput{
		Prompt:         "Hello Gemini",
		ConversationID: &conversationID,
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.ConversationID)
	assert.Contains(t, result.Response, "Mock response to: Hello Gemini")
}

func TestInstanceRun_WhenGeminiOutputsMalformedJSON_ThenIgnoresAndCompletesWithoutError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_MALFORMED_JSON", "1")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.NoError(t, err)
	assert.Empty(t, result.Response)
	assert.Empty(t, result.ConversationID)
}

func TestInstanceRun_WhenGeminiPartialFail_ThenReturnsRuntimeErrorWithExitCodeAndStderr(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_PARTIAL_FAIL", "1")
	t.Setenv("MOCK_GEMINI_EXIT_CODE", "42")
	t.Setenv("MOCK_GEMINI_STDERR", "execution failed")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.Error(t, err)
	assert.Contains(t, result.Response, "Mock response to: Hello Gemini")

	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 42, *runtimeErr.ExitCode)
	assert.Equal(t, "execution failed", runtimeErr.Message)
}

func TestInstanceRun_WhenGeminiFailAfterSuccess_ThenReturnsRuntimeErrorWithExitCode(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_FAIL", "1")
	t.Setenv("MOCK_GEMINI_EXIT_CODE", "7")
	t.Setenv("MOCK_GEMINI_STDERR", "post-success failure")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	_, err = instance.Wait(context.Background())
	require.Error(t, err)

	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 7, *runtimeErr.ExitCode)
	assert.Equal(t, "post-success failure", runtimeErr.Message)
}

func TestInstanceWait_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	instance := &Instance{
		done: make(chan struct{}),
	}

	_, err := instance.Wait(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestInstance_Events_ReturnsEventsChannel(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), briefkit.ExecutionInput{Prompt: "Hello"}, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())
	require.NoError(t, err)

	events := instance.Events()

	assert.NotNil(t, events)
	waitForInstance(t, instance)
}

func TestInstanceRun_WhenGeminiEmitsEmptyLines_ThenSkipsWithoutError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_EMPTY_LINES", "1")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.NoError(t, err)
	assert.Contains(t, result.Response, "Mock response to: Hello Gemini")
}

func TestInstanceRun_WhenGeminiEmitsUnknownEvents_ThenIgnoresAndContinues(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_UNKNOWN_EVENTS", "1")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	result, err := instance.Wait(context.Background())
	require.NoError(t, err)
	assert.Contains(t, result.Response, "Mock response to: Hello Gemini")
}

func TestInstanceRun_WhenGeminiFailWithoutStderr_ThenUsesExitErrorMessage(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_FAIL", "1")
	t.Setenv("MOCK_GEMINI_EXIT_CODE", "5")
	input := briefkit.ExecutionInput{
		Prompt: "Hello Gemini",
	}

	instance, err := newInstance(context.Background(), briefkit.ExecutionID("exec-1"), input, RuntimeConfig{}, briefkit.RuntimeFeatures{}, t.TempDir())

	require.NoError(t, err)
	_, err = instance.Wait(context.Background())
	require.Error(t, err)

	var runtimeErr *briefkit.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Contains(t, runtimeErr.Message, "exit status")
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 5, *runtimeErr.ExitCode)
}

func waitForInstance(t *testing.T, instance *Instance) {
	t.Helper()

	_, err := instance.Wait(context.Background())
	require.NoError(t, err)
}
