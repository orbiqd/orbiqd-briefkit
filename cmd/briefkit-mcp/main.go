package main

import (
	"context"

	"github.com/alecthomas/kong"
	briefkit_mcp "github.com/orbiqd/orbiqd-briefkit/internal/app/briefkit-mcp"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var command briefkit_mcp.Command

	ctx := kong.Parse(&command,
		kong.Name("briefkit-mcp"),
		kong.Description("OrbiqD BriefKit MCP Server - Model Context Protocol stdio server"),
		kong.UsageOnError(),
		kong.Vars{"version": version + " (" + commit + ", " + date + ")"},
	)

	cliCtx := context.Background()
	err := ctx.BindToProvider(func() (context.Context, error) {
		return cliCtx, nil
	})
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	executionRepository, err := cli.CreateExecutionRepositoryFromConfig(command.Store)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	ctx.BindTo(executionRepository, (*briefkit.ExecutionRepository)(nil))

	configRepository, err := cli.CreateConfigRepositoryFromConfig(command.Store)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	ctx.BindTo(configRepository, (*briefkit.ConfigRepository)(nil))

	workspaceManager, err := cli.CreateWorkspaceManagerFromConfig(command.Store)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	ctx.Bind(workspaceManager)

	runner := runtime.NewRunner(executionRepository)

	localClient := briefkit.NewLocalClient(runner, executionRepository, configRepository)
	ctx.BindTo(localClient, (*briefkit.Client)(nil))

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
