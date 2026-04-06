package briefkit_mcp

import (
	"context"
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

type Command struct {
	Log     cli.LogConfig   `embed:"" prefix:"log-"`
	Store   cli.StoreConfig `embed:"" prefix:"store-"`
	Version VersionFlag     `name:"version" help:"Print version information."`
}

type VersionFlag bool

func (v VersionFlag) BeforeApply(app *kong.Kong) error {
	fmt.Println(app.Model.Vars()["version"])
	app.Exit(0)
	return nil
}

func (command *Command) Run(ctx context.Context, agentConfigRepository briefkit.ConfigRepository, client briefkit.Client, workspaceManager *workspace.Manager) error {
	agentIds, err := agentConfigRepository.List(ctx)
	if err != nil {
		return fmt.Errorf("list agent ids: %w", err)
	}

	workspaceSchemes := workspaceManager.Schemes()

	var agentExecTools []mcpserver.ServerTool

	for _, agentId := range agentIds {
		_, err := agentConfigRepository.Get(ctx, agentId)
		if err != nil {
			return fmt.Errorf("get agent config: %s: %w", agentId, err)
		}

		agentExecTool := createExecTool(agentId, client, workspaceSchemes)

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
