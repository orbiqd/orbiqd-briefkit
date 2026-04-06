package briefkit_mcp

import (
	"context"
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func createExecTool(agentId briefkit.AgentID, client briefkit.Client) mcpserver.ServerTool {
	toolName := "ask_" + strcase.ToSnake(string(agentId))

	tool := mcp.NewTool(toolName,

		mcp.WithDescription(fmt.Sprintf("Ask an %s agent anything. Returns a result and a 'conversationId'. To continue a session, you MUST pass the returned 'conversationId' in subsequent calls.", agentId)),
		mcp.WithString("prompt",
			mcp.Description("The comprehensive instruction or message for the agent."),
			mcp.Required(),
		),
		mcp.WithString("model",
			mcp.Description("Optional model override for the execution."),
		),
		mcp.WithString("conversationId",
			mcp.Description("Pass the 'conversationId' received from a previous execution to continue that specific session. Leave empty for new conversations."),
		),
		mcp.WithString("reasoningEffort",
			mcp.Description("Optional reasoning effort override for runtimes that support it (e.g. low/medium/high/xhigh for Codex, low/medium/high/max for Claude). Not supported by Gemini."),
		),
		mcp.WithString("workspace",
			mcp.Description("Optional workspace URI. When provided, the agent works on an isolated copy of the source. Supported schemes: dir:// (e.g. dir:///path/to/project)."),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prompt, err := request.RequireString("prompt")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var opts []briefkit.AskOption
		model := request.GetString("model", "")
		if model != "" {
			opts = append(opts, briefkit.AskWithModel(model))
		}

		conversationId := request.GetString("conversationId", "")
		if conversationId != "" {
			opts = append(opts, briefkit.AskWithConversationID(briefkit.ConversationID(conversationId)))
		}

		reasoningEffort := request.GetString("reasoningEffort", "")
		if reasoningEffort != "" {
			opts = append(opts, briefkit.AskWithReasoningEffort(reasoningEffort))
		}

		workspace := request.GetString("workspace", "")
		if workspace != "" {
			opts = append(opts, briefkit.AskWithWorkspace(workspace))
		}

		result, err := client.Ask(ctx, agentId, prompt, opts...)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		executionResult := briefkit.ExecutionResult{
			ConversationID: result.ConversationID,
			Response:       result.Response,
		}
		return mcp.NewToolResultStructured(executionResult, executionResult.Response), nil
	}

	return mcpserver.ServerTool{
		Tool:    tool,
		Handler: handler,
	}
}
