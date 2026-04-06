package codex

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

const envExecutablePath = "CODEX_EXECUTABLE"

var defaultExecutableCandidates = []string{"codex"}

func locateExecutable(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if envPath := os.Getenv(envExecutablePath); envPath != "" {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", fmt.Errorf("resolve %s path: %w", envExecutablePath, err)
		}

		if _, err := os.Stat(absPath); err != nil { //nolint:gosec // Path is user-supplied via env var and already resolved with filepath.Abs.
			return "", fmt.Errorf("executable from %s not found: %w", envExecutablePath, err)
		}

		return absPath, nil
	}

	path, err := utils.LookupExecutable(ctx, defaultExecutableCandidates)
	if err != nil {
		return "", fmt.Errorf("lookup codex executable: %w", err)
	}

	return path, nil
}
