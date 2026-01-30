package codex

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance_New_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	setupCodexMock(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := newInstance(ctx, agent.ExecutionID("exec-1"), agent.ExecutionInput{Prompt: "hello"}, RuntimeConfig{}, agent.RuntimeFeatures{}, t.TempDir())

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

func TestInstance_New_WhenWorkingDirectoryInvalid_ThenReturnsError(t *testing.T) {
	setupCodexMock(t)

	workingDir := filepath.Join(t.TempDir(), "missing")
	input := agent.ExecutionInput{Prompt: "hello", WorkingDirectory: &workingDir}

	_, err := newInstance(context.Background(), agent.ExecutionID("exec-1"), input, RuntimeConfig{}, agent.RuntimeFeatures{}, t.TempDir())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "start codex")
}

func TestInstance_New_WhenWorkingDirectoryNil_ThenUsesCurrentDir(t *testing.T) {
	setupCodexMock(t)

	instance, err := newInstance(
		context.Background(),
		agent.ExecutionID("exec-1"),
		agent.ExecutionInput{Prompt: "test"},
		RuntimeConfig{RequireWorkspaceRepository: false},
		agent.RuntimeFeatures{},
		t.TempDir(),
	)

	require.NoError(t, err)
	require.NotNil(t, instance)

	cwd, _ := os.Getwd()
	assert.Equal(t, cwd, instance.cmd.Dir)

	_, _ = waitInstance(t, instance)
}

func TestInstance_New_WhenWorkingDirectoryEmpty_ThenUsesCurrentDir(t *testing.T) {
	setupCodexMock(t)

	emptyDir := "   "
	instance, err := newInstance(
		context.Background(),
		agent.ExecutionID("exec-1"),
		agent.ExecutionInput{Prompt: "test", WorkingDirectory: &emptyDir},
		RuntimeConfig{RequireWorkspaceRepository: false},
		agent.RuntimeFeatures{},
		t.TempDir(),
	)

	require.NoError(t, err)
	require.NotNil(t, instance)

	cwd, _ := os.Getwd()
	assert.Equal(t, cwd, instance.cmd.Dir)

	_, _ = waitInstance(t, instance)
}

func TestInstance_Run_WhenParsingSucceedsUnderVariations_ThenReturnsResult(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
	}{
		{name: "Base", env: nil},
		{name: "EmptyLines", env: map[string]string{"MOCK_CODEX_EMPTY_LINES": "1"}},
		{name: "WhitespaceVariations", env: map[string]string{"MOCK_CODEX_WHITESPACE_VARIATIONS": "1"}},
		{name: "UnknownEvents", env: map[string]string{"MOCK_CODEX_UNKNOWN_EVENTS": "1"}},
		{name: "MalformedJSON", env: map[string]string{"MOCK_CODEX_MALFORMED_JSON": "1"}},
		{name: "MixedOutput", env: map[string]string{"MOCK_CODEX_MIXED_OUTPUT": "1"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupCodexMock(t)
			for key, value := range tc.env {
				t.Setenv(key, value)
			}

			result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

			require.NoError(t, err)
			assert.NotEmpty(t, result.ConversationID)
			assert.Contains(t, result.Response, "hello")
		})
	}
}

func TestInstance_Run_WhenOtherItemTypeThenAgentMessage_ThenUsesAgentMessage(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_OTHER_ITEM_TYPE", "1")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.NoError(t, err)
	assert.Contains(t, result.Response, "Mock response to: hello")
}

func TestInstance_Run_WhenMultiItemCompleted_ThenLastWins(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_MULTI_ITEM", "1")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.NoError(t, err)
	assert.Equal(t, "Second response", result.Response)
}

func TestInstance_Run_WhenEmptyText_ThenResponseEmpty(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_EMPTY_TEXT", "1")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.NoError(t, err)
	assert.Empty(t, result.Response)
}

func TestInstance_Run_WhenNoItemCompleted_ThenResponseEmpty(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_NO_RESULT", "1")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.NoError(t, err)
	assert.NotEmpty(t, result.ConversationID)
	assert.Empty(t, result.Response)
}

func TestInstance_Run_WhenNoThreadStarted_ThenConversationIDEmpty(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_NO_THREAD_STARTED", "1")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.NoError(t, err)
	assert.Equal(t, agent.ConversationID(""), result.ConversationID)
	assert.Contains(t, result.Response, "hello")
}

func TestInstance_Run_WhenEmptyStdinAllowed_ThenResponseUsesEmptyPrompt(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_EMPTY_STDIN", "1")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: ""})

	require.NoError(t, err)
	assert.Equal(t, "Mock response to: ", result.Response)
}

func TestInstance_Run_WhenResumeConversation_ThenUsesSessionID(t *testing.T) {
	setupCodexMock(t)
	sessionID := agent.ConversationID("session-123")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello", ConversationID: &sessionID})

	require.NoError(t, err)
	assert.Equal(t, sessionID, result.ConversationID)
	assert.Contains(t, result.Response, "Resumed: session-123")
}

func TestInstance_Run_WhenProcessFailsWithStderr_ThenRuntimeErrorUsesStderr(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_FAIL", "1")
	t.Setenv("MOCK_CODEX_STDERR", "boom")
	t.Setenv("MOCK_CODEX_EXIT_CODE", "7")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.Error(t, err)
	var runtimeErr *agent.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Equal(t, "boom", runtimeErr.Message)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 7, *runtimeErr.ExitCode)
	assert.Contains(t, result.Response, "hello")
}

func TestInstance_Run_WhenProcessFailsWithoutStderr_ThenRuntimeErrorUsesErr(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_FAIL", "1")

	_, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.Error(t, err)
	var runtimeErr *agent.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Contains(t, runtimeErr.Message, "exit status")
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 1, *runtimeErr.ExitCode)
}

func TestInstance_Run_WhenPartialFail_ThenReturnsErrorAndResponseEmpty(t *testing.T) {
	setupCodexMock(t)
	t.Setenv("MOCK_CODEX_PARTIAL_FAIL", "1")
	t.Setenv("MOCK_CODEX_STDERR", "partial failure")
	t.Setenv("MOCK_CODEX_EXIT_CODE", "9")

	result, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

	require.Error(t, err)
	var runtimeErr *agent.RuntimeExecutionError
	require.ErrorAs(t, err, &runtimeErr)
	assert.Equal(t, "partial failure", runtimeErr.Message)
	require.NotNil(t, runtimeErr.ExitCode)
	assert.Equal(t, 9, *runtimeErr.ExitCode)
	assert.Empty(t, result.Response)
}

func TestInstance_Run_WhenSignalReceived_ThenExitCodeMatchesSignal(t *testing.T) {
	cases := []struct {
		name          string
		signal        string
		expectedCodes []int
	}{
		{name: "SIGINT", signal: "SIGINT", expectedCodes: []int{130, -1}},
		{name: "SIGTERM", signal: "SIGTERM", expectedCodes: []int{143, -1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupCodexMock(t)
			t.Setenv("MOCK_CODEX_SIGNAL", tc.signal)

			_, err := runInstance(t, agent.ExecutionInput{Prompt: "hello"})

			require.Error(t, err)
			var runtimeErr *agent.RuntimeExecutionError
			require.ErrorAs(t, err, &runtimeErr)
			require.NotNil(t, runtimeErr.ExitCode)
			assert.Contains(t, tc.expectedCodes, *runtimeErr.ExitCode)
		})
	}
}

func TestInstance_Events_WhenStartedAndFinished_ThenEmitsRuntimeEvents(t *testing.T) {
	setupCodexMock(t)

	instance := newTestInstance(t, agent.ExecutionInput{Prompt: "hello"})
	_, err := waitInstance(t, instance)

	require.NoError(t, err)

	var kinds []agent.RuntimeEventKind
	for event := range instance.Events() {
		kinds = append(kinds, event.Kind())
	}

	require.Len(t, kinds, 2)
	assert.Equal(t, agent.RuntimeEventStarted, kinds[0])
	assert.Equal(t, agent.RuntimeEventFinished, kinds[1])
}

func TestInstance_Wait_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	setupCodexMock(t)

	instance := newTestInstance(t, agent.ExecutionInput{Prompt: "hello"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := instance.Wait(ctx)

	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)

	_, err = waitInstance(t, instance)
	require.NoError(t, err)
}

func TestInstance_EmitRuntimeEvent_WhenChannelNil_ThenReturnsEarly(t *testing.T) {
	instance := &Instance{events: nil}

	instance.emitRuntimeEvent(agent.RuntimeStartedEvent{})
}

func TestInstance_EmitRuntimeEvent_WhenChannelFull_ThenDropsEvent(t *testing.T) {
	instance := &Instance{
		events: make(chan agent.RuntimeEvent, 1),
	}
	instance.events <- agent.RuntimeStartedEvent{}

	instance.emitRuntimeEvent(agent.RuntimeFinishedEvent{})

	assert.Len(t, instance.events, 1)
}

func setupCodexMock(t *testing.T) {
	t.Helper()
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
}

func newTestInstance(t *testing.T, input agent.ExecutionInput) *Instance {
	t.Helper()

	instance, err := newInstance(
		context.Background(),
		agent.ExecutionID("exec-1"),
		input,
		RuntimeConfig{RequireWorkspaceRepository: false},
		agent.RuntimeFeatures{},
		t.TempDir(),
	)

	require.NoError(t, err)
	return instance
}

func waitInstance(t *testing.T, instance *Instance) (agent.RuntimeResult, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return instance.Wait(ctx)
}

func runInstance(t *testing.T, input agent.ExecutionInput) (agent.RuntimeResult, error) {
	t.Helper()

	instance := newTestInstance(t, input)
	return waitInstance(t, instance)
}
