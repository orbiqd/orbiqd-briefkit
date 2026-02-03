package main

import (
	"context"
	"log/slog"

	"github.com/alecthomas/kong"
	briefkit_ctl "github.com/orbiqd/orbiqd-briefkit/internal/app/briefkit-ctl"
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
	var command briefkit_ctl.Command

	ctx := kong.Parse(&command,
		kong.Name("briefkit-ctl"),
		kong.Description("OrbiqD BriefKit CLI - Manage agent collaborations"),
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
	ctx.BindTo(executionRepository, (*briefkit.ExecutionRepository)(nil))

	configRepository, err := cli.CreateConfigRepositoryFromConfig(command.Store)
	if err != nil {
		ctx.FatalIfErrorf(err)
	}
	ctx.BindTo(configRepository, (*briefkit.ConfigRepository)(nil))

	runner := runtime.NewRunner(executionRepository)

	localClient := briefkit.NewLocalClient(runner, executionRepository, configRepository)
	ctx.BindTo(localClient, (*briefkit.Client)(nil))

	runtimeRegistry := runtime.NewRegistry()
	ctx.BindTo(runtimeRegistry, (*briefkit.RuntimeRegistry)(nil))

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
