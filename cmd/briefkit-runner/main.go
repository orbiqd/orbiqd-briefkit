package main

import (
	"context"
	"log/slog"

	"github.com/alecthomas/kong"
	briefkit_runner "github.com/orbiqd/orbiqd-briefkit/internal/app/briefkit-runner"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var command briefkit_runner.RunnerCommand

	ctx := kong.Parse(&command,
		kong.Name("briefkit-runner"),
		kong.Description("OrbiqD BriefKit Runner - Execute agent instances"),
		kong.UsageOnError(),
		kong.Vars{"version": version + " (" + commit + ", " + date + ")"},
	)

	logger, err := cli.CreateLoggerFromConfig(command.Log)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	slog.SetDefault(logger)

	executionRepository, err := cli.CreateExecutionRepositoryFromConfig(command.Store)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	ctx.BindTo(executionRepository, (*agent.ExecutionRepository)(nil))

	ctx.BindTo(runtime.NewRegistry(), (*agent.RuntimeRegistry)(nil))

	cliCtx := context.Background()
	err = ctx.BindToProvider(func() (context.Context, error) {
		return cliCtx, nil
	})
	if err != nil {
		ctx.FatalIfErrorf(err)
	}

	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
