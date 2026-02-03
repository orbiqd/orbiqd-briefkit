package runtime

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"

	"github.com/coder/quartz"
)

var ErrExecutionFailed = errors.New("execution failed")

type Runner struct {
	executionRepository briefkit.ExecutionRepository
	pollInterval        time.Duration
	clock               quartz.Clock
	executableResolver  func(ctx context.Context, name string) (string, error)
}

type RunnerOpt func(*Runner)

func WithPollInterval(interval time.Duration) RunnerOpt {
	return func(r *Runner) {
		r.pollInterval = interval
	}
}

func WithClock(clock quartz.Clock) RunnerOpt {
	return func(r *Runner) {
		r.clock = clock
	}
}

func WithExecutableResolver(resolver func(context.Context, string) (string, error)) RunnerOpt {
	return func(r *Runner) {
		r.executableResolver = resolver
	}
}

func NewRunner(executionRepository briefkit.ExecutionRepository, opts ...RunnerOpt) *Runner {
	runner := &Runner{
		executionRepository: executionRepository,
		pollInterval:        500 * time.Millisecond,
		clock:               quartz.NewReal(),
		executableResolver:  cli.ResolveExecutable,
	}

	for _, opt := range opts {
		opt(runner)
	}

	return runner
}

func (runner *Runner) Spawn(ctx context.Context, executionId briefkit.ExecutionID) error {
	executablePath, err := runner.executableResolver(ctx, cli.ExecutableRunner)
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

func (runner *Runner) Wait(ctx context.Context, executionId briefkit.ExecutionID) error {
	execution, err := runner.executionRepository.Get(ctx, executionId)
	if err != nil {
		return fmt.Errorf("get execution: %w", err)
	}

	ticker := runner.clock.NewTicker(runner.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for completion: %w", ctx.Err())
		case <-ticker.C:
			status, err := execution.GetStatus(ctx)
			if err != nil {
				return fmt.Errorf("get execution status: %w", err)
			}

			if status.State.IsFinished() {
				if status.State == briefkit.ExecutionFailed {
					var parts []string

					if status.Error != nil {
						parts = append(parts, *status.Error)
					}

					if status.ExitCode != nil {
						parts = append(parts, fmt.Sprintf("Exit code is %d.", *status.ExitCode))
					}

					if len(parts) == 0 {
						return ErrExecutionFailed
					}

					return fmt.Errorf("%w: %s", ErrExecutionFailed, strings.Join(parts, " "))
				}
				return nil
			}
		}
	}
}
