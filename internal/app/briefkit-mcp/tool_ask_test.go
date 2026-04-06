package briefkit_mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateExecTool_ToolName_WhenCamelCaseAgentID_ThenSnakeCaseName(t *testing.T) {
	client := briefkit.NewMockClient(t)

	st := createExecTool("myAgent", client, []string{"dir"})

	assert.Equal(t, "ask_my_agent", st.Tool.Name)
}

func TestCreateExecTool_ToolName_WhenSimpleAgentID_ThenPrefixedName(t *testing.T) {
	client := briefkit.NewMockClient(t)

	st := createExecTool("codex", client, []string{"dir"})

	assert.Equal(t, "ask_codex", st.Tool.Name)
}

func TestCreateExecTool_Handler_WhenOnlyPrompt_ThenCallsAskWithoutOptions(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"prompt": "hello"}

	var gotOpts []briefkit.AskOption
	client.EXPECT().Ask(ctx, agentID, "hello").
		Run(func(_ context.Context, _ briefkit.AgentID, _ string, opts ...briefkit.AskOption) {
			gotOpts = append([]briefkit.AskOption(nil), opts...)
		}).
		Return(&briefkit.AskResult{ConversationID: "conv-1", Response: "ok"}, nil)

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Empty(t, gotOpts)
}

func TestCreateExecTool_Handler_WhenModelProvided_ThenPassesModelOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"prompt": "hello",
		"model":  "gpt-4.1",
	}

	var gotOpts []briefkit.AskOption
	client.EXPECT().Ask(ctx, agentID, "hello", mock.Anything).
		Run(func(_ context.Context, _ briefkit.AgentID, _ string, opts ...briefkit.AskOption) {
			gotOpts = append([]briefkit.AskOption(nil), opts...)
		}).
		Return(&briefkit.AskResult{Response: "ok"}, nil)

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	got := applyMCPAskOptions(gotOpts)
	require.NotNil(t, got.Model)
	assert.Equal(t, "gpt-4.1", *got.Model)
}

func TestCreateExecTool_Handler_WhenConversationIDProvided_ThenPassesConversationIDOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"prompt":         "hello",
		"conversationId": "conv-abc",
	}

	var gotOpts []briefkit.AskOption
	client.EXPECT().Ask(ctx, agentID, "hello", mock.Anything).
		Run(func(_ context.Context, _ briefkit.AgentID, _ string, opts ...briefkit.AskOption) {
			gotOpts = append([]briefkit.AskOption(nil), opts...)
		}).
		Return(&briefkit.AskResult{Response: "ok"}, nil)

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	got := applyMCPAskOptions(gotOpts)
	require.NotNil(t, got.ConversationID)
	assert.Equal(t, briefkit.ConversationID("conv-abc"), *got.ConversationID)
}

func TestCreateExecTool_Handler_WhenReasoningEffortProvided_ThenPassesReasoningEffortOption(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"prompt":          "hello",
		"reasoningEffort": "high",
	}

	var gotOpts []briefkit.AskOption
	client.EXPECT().Ask(ctx, agentID, "hello", mock.Anything).
		Run(func(_ context.Context, _ briefkit.AgentID, _ string, opts ...briefkit.AskOption) {
			gotOpts = append([]briefkit.AskOption(nil), opts...)
		}).
		Return(&briefkit.AskResult{Response: "ok"}, nil)

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	got := applyMCPAskOptions(gotOpts)
	require.NotNil(t, got.ReasoningEffort)
	assert.Equal(t, "high", *got.ReasoningEffort)
}

func TestCreateExecTool_Handler_WhenAllParamsProvided_ThenPassesAllOptions(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"prompt":          "hello",
		"model":           "gpt-4.1",
		"conversationId":  "conv-xyz",
		"reasoningEffort": "xhigh",
	}

	var gotOpts []briefkit.AskOption
	client.EXPECT().Ask(ctx, agentID, "hello", mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ briefkit.AgentID, _ string, opts ...briefkit.AskOption) {
			gotOpts = append([]briefkit.AskOption(nil), opts...)
		}).
		Return(&briefkit.AskResult{ConversationID: "conv-xyz", Response: "done"}, nil)

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	require.Len(t, gotOpts, 3)
	got := applyMCPAskOptions(gotOpts)
	require.NotNil(t, got.Model)
	assert.Equal(t, "gpt-4.1", *got.Model)
	require.NotNil(t, got.ConversationID)
	assert.Equal(t, briefkit.ConversationID("conv-xyz"), *got.ConversationID)
	require.NotNil(t, got.ReasoningEffort)
	assert.Equal(t, "xhigh", *got.ReasoningEffort)
}

func TestCreateExecTool_Handler_WhenPromptMissing_ThenReturnsToolError(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	st := createExecTool("codex", client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestCreateExecTool_Handler_WhenClientReturnsError_ThenReturnsToolError(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"prompt": "hello"}

	client.EXPECT().Ask(ctx, agentID, "hello").
		Return(nil, errors.New("runtime failure"))

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
}

func TestCreateExecTool_Handler_WhenSuccess_ThenResultContainsResponseAndConversationID(t *testing.T) {
	ctx := context.Background()
	client := briefkit.NewMockClient(t)
	agentID := briefkit.AgentID("codex")
	st := createExecTool(agentID, client, []string{"dir"})

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"prompt": "hello"}

	client.EXPECT().Ask(ctx, agentID, "hello").
		Return(&briefkit.AskResult{ConversationID: "conv-1", Response: "the answer"}, nil)

	result, err := st.Handler(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
}

func applyMCPAskOptions(opts []briefkit.AskOption) briefkit.AskOptions {
	var o briefkit.AskOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
