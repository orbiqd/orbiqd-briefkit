package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/mitchellh/go-homedir"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/process"
)

var ErrExecutableNotFound = errors.New("executable not found")

const (
	ExecutableCtl    = "briefkit-ctl"
	ExecutableMCP    = "briefkit-mcp"
	ExecutableRunner = "briefkit-runner"
)

func ResolveRuntimeLogDir() (string, error) {
	dir := os.Getenv("BRIEFKIT_RUNTIME_LOG_DIR")
	if dir == "" {
		dir = "~/.orbiqd/briefkit/logs/runtime/"
	}

	expanded, err := homedir.Expand(dir)
	if err != nil {
		return "", fmt.Errorf("expand runtime log dir: %w", err)
	}

	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolve absolute runtime log dir: %w", err)
	}

	return abs, nil
}

func ResolveExecutable(ctx context.Context, executableName string) (string, error) {
	envVarName := "BRIEFKIT_" + strcase.ToScreamingSnake(executableName) + "_PATH"

	if envPath, ok := os.LookupEnv(envVarName); ok {
		if _, err := os.Stat(envPath); err != nil {
			return "", fmt.Errorf("%w: executable from %s: %w", ErrExecutableNotFound, envVarName, err)
		}
		return envPath, nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get executable path: %w", err)
	}

	evalPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", fmt.Errorf("eval symlinks: %w", err)
	}

	execDir := filepath.Dir(evalPath)
	executablePath := filepath.Join(execDir, executableName)

	if _, err := os.Stat(executablePath); err == nil {
		return executablePath, nil
	}

	pathExecutable, err := process.LookupExecutable(ctx, []string{executableName})
	if err != nil {
		return "", ErrExecutableNotFound
	}

	return pathExecutable, nil
}
