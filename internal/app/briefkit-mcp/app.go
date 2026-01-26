package briefkit_mcp

import (
	"context"
	"fmt"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"

	//"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type Command struct {
	Log   cli.LogConfig   `embed:"" prefix:"log-"`
	Store cli.StoreConfig `embed:"" prefix:"store-"`
}

func (command *Command) Run(ctx context.Context, agentConfigRepository agent.ConfigRepository, executionRepository agent.ExecutionRepository) error {
	agentIds, err := agentConfigRepository.List(ctx)
	if err != nil {
		return fmt.Errorf("list agent ids: %w", err)
	}

	var agentExecTools []mcpserver.ServerTool

	for _, agentId := range agentIds {
		agentConfig, err := agentConfigRepository.Get(ctx, agentId)
		if err != nil {
			return fmt.Errorf("get agent config: %s: %w", agentId, err)
		}

		agentExecTool := createExecTool(agentId, agentConfig, executionRepository)

		agentExecTools = append(agentExecTools, agentExecTool)
	}

	server := mcpserver.NewMCPServer(
		"briefkit-mcp",
		"1.0.0",
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithRecovery(),
	)

	server.AddTools(agentExecTools...)

	if err := mcpserver.ServeStdio(server); err != nil {
		return fmt.Errorf("server MCP: %w", err)
	}

	return nil
}
