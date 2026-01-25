package utils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupExecutable_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	t.Run("returns context error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := LookupExecutable(ctx, []string{"anything"})

		require.ErrorIs(t, err, context.Canceled)
		assert.Empty(t, result)
	})
}

func TestLookupExecutable_WhenCandidatesEmpty_ThenReturnsErrNotFound(t *testing.T) {
	t.Run("returns ErrNotFound", func(t *testing.T) {
		result, err := LookupExecutable(context.Background(), nil)

		require.ErrorIs(t, err, exec.ErrNotFound)
		assert.Empty(t, result)
	})
}

func TestLookupExecutable_WhenCandidateIsExistingPath_ThenReturnsPath(t *testing.T) {
	t.Run("returns executable path", func(t *testing.T) {
		executablePath := currentExecutablePath(t)

		result, err := LookupExecutable(context.Background(), []string{executablePath})

		require.NoError(t, err)
		assert.Equal(t, executablePath, result)
	})
}

func TestLookupExecutable_WhenFirstMissing_ThenReturnsNextFound(t *testing.T) {
	t.Run("skips missing and returns next", func(t *testing.T) {
		executablePath := currentExecutablePath(t)
		t.Setenv("PATH", t.TempDir())

		result, err := LookupExecutable(context.Background(), []string{"missing-tool", executablePath})

		require.NoError(t, err)
		assert.Equal(t, executablePath, result)
	})
}

func TestLookupExecutable_WhenFoundInPath_ThenReturnsPath(t *testing.T) {
	t.Run("returns path resolved via PATH", func(t *testing.T) {
		executablePath := currentExecutablePath(t)
		t.Setenv("PATH", filepath.Dir(executablePath))

		result, err := LookupExecutable(context.Background(), []string{filepath.Base(executablePath)})

		require.NoError(t, err)
		assert.Equal(t, executablePath, result)
	})
}

func currentExecutablePath(t *testing.T) string {
	t.Helper()

	path, err := os.Executable()
	require.NoError(t, err)

	return path
}
