package briefkit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func applyAskOptions(opts ...AskOption) AskOptions {
	var o AskOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func TestAskWithConversationID(t *testing.T) {
	t.Run("WhenCalled_ThenSetsConversationID", func(t *testing.T) {
		got := applyAskOptions(AskWithConversationID("conv-1"))

		require.NotNil(t, got.ConversationID)
		assert.Equal(t, ConversationID("conv-1"), *got.ConversationID)
	})

	t.Run("WhenCalledWithEmpty_ThenSetsEmptyConversationID", func(t *testing.T) {
		got := applyAskOptions(AskWithConversationID(""))

		require.NotNil(t, got.ConversationID)
		assert.Equal(t, ConversationID(""), *got.ConversationID)
	})

	t.Run("WhenNotApplied_ThenConversationIDIsNil", func(t *testing.T) {
		got := applyAskOptions()

		assert.Nil(t, got.ConversationID)
	})
}

func TestAskWithModel(t *testing.T) {
	t.Run("WhenCalled_ThenSetsModel", func(t *testing.T) {
		got := applyAskOptions(AskWithModel("gpt-4"))

		require.NotNil(t, got.Model)
		assert.Equal(t, "gpt-4", *got.Model)
	})

	t.Run("WhenCalledWithEmpty_ThenSetsEmptyModel", func(t *testing.T) {
		got := applyAskOptions(AskWithModel(""))

		require.NotNil(t, got.Model)
		assert.Empty(t, *got.Model)
	})

	t.Run("WhenNotApplied_ThenModelIsNil", func(t *testing.T) {
		got := applyAskOptions()

		assert.Nil(t, got.Model)
	})
}

func TestAskWithReasoningEffort(t *testing.T) {
	t.Run("WhenCalled_ThenSetsReasoningEffort", func(t *testing.T) {
		got := applyAskOptions(AskWithReasoningEffort("high"))

		require.NotNil(t, got.ReasoningEffort)
		assert.Equal(t, "high", *got.ReasoningEffort)
	})

	t.Run("WhenCalledWithEmpty_ThenSetsEmptyReasoningEffort", func(t *testing.T) {
		got := applyAskOptions(AskWithReasoningEffort(""))

		require.NotNil(t, got.ReasoningEffort)
		assert.Empty(t, *got.ReasoningEffort)
	})

	t.Run("WhenNotApplied_ThenReasoningEffortIsNil", func(t *testing.T) {
		got := applyAskOptions()

		assert.Nil(t, got.ReasoningEffort)
	})
}

func TestAskWithTimeout(t *testing.T) {
	t.Run("WhenCalled_ThenSetsTimeout", func(t *testing.T) {
		got := applyAskOptions(AskWithTimeout(2 * time.Minute))

		require.NotNil(t, got.Timeout)
		assert.Equal(t, 2*time.Minute, *got.Timeout)
	})

	t.Run("WhenCalledWithZero_ThenSetsZeroTimeout", func(t *testing.T) {
		got := applyAskOptions(AskWithTimeout(0))

		require.NotNil(t, got.Timeout)
		assert.Equal(t, time.Duration(0), *got.Timeout)
	})

	t.Run("WhenNotApplied_ThenTimeoutIsNil", func(t *testing.T) {
		got := applyAskOptions()

		assert.Nil(t, got.Timeout)
	})
}

func TestAskWithWorkspace(t *testing.T) {
	t.Run("WhenCalled_ThenSetsWorkspace", func(t *testing.T) {
		got := applyAskOptions(AskWithWorkspace("dir:///tmp/project"))

		require.NotNil(t, got.Workspace)
		assert.Equal(t, "dir:///tmp/project", *got.Workspace)
	})

	t.Run("WhenCalledWithEmpty_ThenSetsEmptyWorkspace", func(t *testing.T) {
		got := applyAskOptions(AskWithWorkspace(""))

		require.NotNil(t, got.Workspace)
		assert.Empty(t, *got.Workspace)
	})

	t.Run("WhenNotApplied_ThenWorkspaceIsNil", func(t *testing.T) {
		got := applyAskOptions()

		assert.Nil(t, got.Workspace)
	})
}

func TestAskOptions_WhenMultipleOptionsApplied_ThenAllFieldsSet(t *testing.T) {
	conversationID := ConversationID("conv-1")
	model := "gpt-4"
	effort := "high"
	timeout := 2 * time.Minute
	workspace := "dir:///tmp/project"

	got := applyAskOptions(
		AskWithConversationID(conversationID),
		AskWithModel(model),
		AskWithReasoningEffort(effort),
		AskWithTimeout(timeout),
		AskWithWorkspace(workspace),
	)

	require.NotNil(t, got.ConversationID)
	assert.Equal(t, conversationID, *got.ConversationID)
	require.NotNil(t, got.Model)
	assert.Equal(t, model, *got.Model)
	require.NotNil(t, got.ReasoningEffort)
	assert.Equal(t, effort, *got.ReasoningEffort)
	require.NotNil(t, got.Timeout)
	assert.Equal(t, timeout, *got.Timeout)
	require.NotNil(t, got.Workspace)
	assert.Equal(t, workspace, *got.Workspace)
}
