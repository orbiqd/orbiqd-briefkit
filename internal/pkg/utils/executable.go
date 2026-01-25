package utils

import (
	"context"
	"errors"
	"os/exec"

	"github.com/cli/safeexec"
)

// LookupExecutable returns the first executable path found for the given candidates.
func LookupExecutable(ctx context.Context, candidates []string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		path, err := safeexec.LookPath(candidate)
		if err == nil || errors.Is(err, exec.ErrDot) {
			if path == "" {
				return "", err
			}
			return path, nil
		}

		if errors.Is(err, exec.ErrNotFound) {
			continue
		}

		return "", err
	}

	return "", exec.ErrNotFound
}
