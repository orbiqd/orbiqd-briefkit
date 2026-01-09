package claude

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/process"
)

const (
	envExecutablePath = "CLAUDE_EXECUTABLE"
	envConfigPath     = "CLAUDE_CONFIG_PATH"

	defaultConfigPath = "~/.claude.json"
)

var defaultExecutableCandidates = []string{"claude", "claude-code"}

func locateExecutable(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if envPath := os.Getenv(envExecutablePath); envPath != "" {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", fmt.Errorf("resolve %s path: %w", envExecutablePath, err)
		}

		if _, err := os.Stat(absPath); err != nil {
			return "", fmt.Errorf("executable from %s not found: %w", envExecutablePath, err)
		}

		return absPath, nil
	}

	path, err := process.LookupExecutable(ctx, defaultExecutableCandidates)
	if err != nil {
		return "", fmt.Errorf("lookup claude executable: %w", err)
	}

	return path, nil
}

func locateConfigPath() (string, error) {
	if envPath := os.Getenv(envConfigPath); envPath != "" {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", fmt.Errorf("resolve %s path: %w", envConfigPath, err)
		}
		return absPath, nil
	}

	expanded, err := homedir.Expand(defaultConfigPath)
	if err != nil {
		return "", fmt.Errorf("expand home directory: %w", err)
	}

	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	return absPath, nil
}
