package briefkitctl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAskCmd_Run_WhenNoOptionsProvided_ThenCallsAskWithoutOptions(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	cmd := &AskCmd{
		AgentID: "codex",
		Prompt:  "test prompt",
	}
	result := &briefkit.AskResult{
		ConversationID: "conv-1",
		Response:       "ok",
	}

	var gotOptions []briefkit.AskOption
	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt).Run(func(ctx context.Context, agentID briefkit.AgentID, prompt string, opts ...briefkit.AskOption) {
		gotOptions = append([]briefkit.AskOption(nil), opts...)
	}).Return(result, nil)

	err := cmd.Run(ctx, client)

	require.NoError(t, err)
	assert.Empty(t, gotOptions)
}

func TestAskCmd_Run_WhenTimeoutProvided_ThenPassesTimeoutOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	timeout := 2 * time.Minute
	cmd := &AskCmd{
		AgentID: "codex",
		Prompt:  "test prompt",
		Timeout: &timeout,
	}
	result := &briefkit.AskResult{
		ConversationID: "conv-1",
		Response:       "ok",
	}

	var gotOptions []briefkit.AskOption
	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt, mock.Anything).Run(func(ctx context.Context, agentID briefkit.AgentID, prompt string, opts ...briefkit.AskOption) {
		gotOptions = append([]briefkit.AskOption(nil), opts...)
	}).Return(result, nil)

	err := cmd.Run(ctx, client)

	require.NoError(t, err)
	require.Len(t, gotOptions, 1)
	got := applyAskOptions(gotOptions)
	require.NotNil(t, got.Timeout)
	assert.Equal(t, timeout, *got.Timeout)
	assert.Nil(t, got.ConversationID)
	assert.Nil(t, got.Model)
}

func TestAskCmd_Run_WhenConversationIDProvided_ThenPassesConversationIDOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	conversationID := briefkit.ConversationID("conv-2")
	cmd := &AskCmd{
		AgentID:        "codex",
		Prompt:         "test prompt",
		ConversationID: &conversationID,
	}
	result := &briefkit.AskResult{
		ConversationID: "conv-1",
		Response:       "ok",
	}

	var gotOptions []briefkit.AskOption
	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt, mock.Anything).Run(func(ctx context.Context, agentID briefkit.AgentID, prompt string, opts ...briefkit.AskOption) {
		gotOptions = append([]briefkit.AskOption(nil), opts...)
	}).Return(result, nil)

	err := cmd.Run(ctx, client)

	require.NoError(t, err)
	require.Len(t, gotOptions, 1)
	got := applyAskOptions(gotOptions)
	assert.Nil(t, got.Timeout)
	require.NotNil(t, got.ConversationID)
	assert.Equal(t, conversationID, *got.ConversationID)
	assert.Nil(t, got.Model)
}

func TestAskCmd_Run_WhenModelProvided_ThenPassesModelOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	model := "gpt-4.1"
	cmd := &AskCmd{
		AgentID: "codex",
		Prompt:  "test prompt",
		Model:   &model,
	}
	result := &briefkit.AskResult{
		ConversationID: "conv-1",
		Response:       "ok",
	}

	var gotOptions []briefkit.AskOption
	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt, mock.Anything).Run(func(ctx context.Context, agentID briefkit.AgentID, prompt string, opts ...briefkit.AskOption) {
		gotOptions = append([]briefkit.AskOption(nil), opts...)
	}).Return(result, nil)

	err := cmd.Run(ctx, client)

	require.NoError(t, err)
	require.Len(t, gotOptions, 1)
	got := applyAskOptions(gotOptions)
	assert.Nil(t, got.Timeout)
	assert.Nil(t, got.ConversationID)
	require.NotNil(t, got.Model)
	assert.Equal(t, model, *got.Model)
}

func TestAskCmd_Run_WhenAllOptionsProvided_ThenPassesAllOptions(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	timeout := 3 * time.Minute
	conversationID := briefkit.ConversationID("conv-3")
	model := "gpt-4.1-mini"
	cmd := &AskCmd{
		AgentID:        "codex",
		Prompt:         "test prompt",
		Timeout:        &timeout,
		ConversationID: &conversationID,
		Model:          &model,
	}
	result := &briefkit.AskResult{
		ConversationID: "conv-1",
		Response:       "ok",
	}

	var gotOptions []briefkit.AskOption
	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt, mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, agentID briefkit.AgentID, prompt string, opts ...briefkit.AskOption) {
		gotOptions = append([]briefkit.AskOption(nil), opts...)
	}).Return(result, nil)

	err := cmd.Run(ctx, client)

	require.NoError(t, err)
	require.Len(t, gotOptions, 3)
	got := applyAskOptions(gotOptions)
	require.NotNil(t, got.Timeout)
	assert.Equal(t, timeout, *got.Timeout)
	require.NotNil(t, got.ConversationID)
	assert.Equal(t, conversationID, *got.ConversationID)
	require.NotNil(t, got.Model)
	assert.Equal(t, model, *got.Model)
}

func TestAskCmd_Run_WhenReasoningEffortProvided_ThenPassesReasoningEffortOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	effort := "high"
	cmd := &AskCmd{
		AgentID:         "codex",
		Prompt:          "test prompt",
		ReasoningEffort: &effort,
	}
	result := &briefkit.AskResult{
		ConversationID: "conv-1",
		Response:       "ok",
	}

	var gotOptions []briefkit.AskOption
	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt, mock.Anything).Run(func(ctx context.Context, agentID briefkit.AgentID, prompt string, opts ...briefkit.AskOption) {
		gotOptions = append([]briefkit.AskOption(nil), opts...)
	}).Return(result, nil)

	err := cmd.Run(ctx, client)

	require.NoError(t, err)
	require.Len(t, gotOptions, 1)
	got := applyAskOptions(gotOptions)
	assert.Nil(t, got.Timeout)
	assert.Nil(t, got.ConversationID)
	assert.Nil(t, got.Model)
	require.NotNil(t, got.ReasoningEffort)
	assert.Equal(t, effort, *got.ReasoningEffort)
}

func TestAskCmd_Run_WhenClientReturnsError_ThenReturnsWrappedError(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	cmd := &AskCmd{
		AgentID: "codex",
		Prompt:  "test prompt",
	}
	expectedErr := errors.New("boom")

	client.EXPECT().Ask(ctx, cmd.AgentID, cmd.Prompt).Return(nil, expectedErr)

	err := cmd.Run(ctx, client)

	require.Error(t, err)
	require.EqualError(t, err, "ask: boom")
	require.ErrorIs(t, err, expectedErr)
}

func applyAskOptions(opts []briefkit.AskOption) briefkit.AskOptions {
	var got briefkit.AskOptions
	for _, opt := range opts {
		opt(&got)
	}
	return got
}
