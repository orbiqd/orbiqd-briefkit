package briefkit_runner

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
)

func Spawn(ctx context.Context, executionId agent.ExecutionID) error {
	executablePath, err := cli.ResolveExecutable(ctx, cli.ExecutableRunner)
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}

	// #nosec G204 - executablePath comes from ResolveExecutable, executionId is controlled
	cmd := exec.CommandContext(ctx, executablePath, string(executionId))

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("release process: %w", err)
	}

	return nil
}
