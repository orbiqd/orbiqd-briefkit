package claude

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

const (
	envExecutablePath = "CLAUDE_EXECUTABLE"
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

	path, err := utils.LookupExecutable(ctx, defaultExecutableCandidates)
	if err != nil {
		return "", fmt.Errorf("lookup claude executable: %w", err)
	}

	return path, nil
}
