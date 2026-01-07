package briefkit_mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	briefkitrunner "github.com/orbiqd/orbiqd-briefkit/internal/app/briefkit-runner"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

func createExecTool(agentId agent.AgentID, agentConfig agent.Config, executionRepository agent.ExecutionRepository) (mcpserver.ServerTool, error) {
	toolName := fmt.Sprintf("exec_%s", strcase.ToSnake(string(agentId)))

	tool := mcp.NewTool(toolName,
		mcp.WithDescription("Runs a prompt on agent, optionally continuing a conversation or overriding the model."),
		mcp.WithString("prompt",
			mcp.Description("Prompt to send to the agent."),
			mcp.Required(),
		),
		mcp.WithString("model",
			mcp.Description("Optional model override for the execution."),
		),
		mcp.WithString("conversationId",
			mcp.Description("Conversation ID to continue an existing agent session."),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		prompt, err := request.RequireString("prompt")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		executionInput := agent.ExecutionInput{
			WorkingDirectory: nil,
			Timeout:          utils.Duration(time.Minute * 5),
			Prompt:           prompt,
		}

		model := request.GetString("model", "")
		if model != "" {
			executionInput.Model = &model
		}

		conversationId := request.GetString("conversationId", "")
		if conversationId != "" {
			executionInput.ConversationID = (*agent.ConversationID)(&conversationId)
		}

		executionId, err := executionRepository.Create(ctx, executionInput, agentConfig)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if err := briefkitrunner.Spawn(ctx, executionId); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		execution, err := executionRepository.Get(ctx, executionId)
		if err != nil {
			return mcp.NewToolResultError(fmt.Errorf("get execution: %w", err).Error()), nil
		}

		for {
			select {
			case <-ctx.Done():
				return mcp.NewToolResultError(fmt.Errorf("wait for completion: %w", err).Error()), nil
			case <-ticker.C:
				status, err := execution.GetStatus(ctx)
				if err != nil {
					return mcp.NewToolResultError(fmt.Errorf("failed to get execution status: %w", err).Error()), nil
				}

				switch status.State {
				case agent.ExecutionSucceeded:
					executionResult, err := execution.GetResult(ctx)
					if err != nil {
						return mcp.NewToolResultError(fmt.Errorf("failed to get execution result: %w", err).Error()), nil
					}

					return mcp.NewToolResultStructured(executionResult, executionResult.Response), nil
				case agent.ExecutionFailed:
					var errors []string
					if status.Error != nil {
						errors = append(errors, fmt.Sprintf("%s", *status.Error))
					}

					if status.ExitCode != nil {
						errors = append(errors, fmt.Sprintf("Exit code is %d.", *status.ExitCode))
					}

					return mcp.NewToolResultErrorf("Execution failed. %s", strings.Join(errors, " ")), nil
				}

			}
		}
	}

	return mcpserver.ServerTool{
		Tool:    tool,
		Handler: handler,
	}, nil
}
