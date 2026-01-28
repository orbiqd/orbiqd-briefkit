package claude

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocateExecutable_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := locateExecutable(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func TestLocateExecutable_WhenPathLookupSucceeds_ThenReturnsPath(t *testing.T) {
	tempDir := t.TempDir()
	executable := filepath.Join(tempDir, "claude")
	require.NoError(t, os.WriteFile(executable, []byte("binary"), 0o600))

	t.Setenv("CLAUDE_EXECUTABLE", "")
	t.Setenv("PATH", "")
	t.Setenv(envExecutablePath, executable)

	path, err := locateExecutable(context.Background())

	require.NoError(t, err)
	require.Equal(t, executable, path)
}
