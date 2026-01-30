package codex

import (
	"context"
	"os"
	"os/exec"
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

func TestLocateExecutable_WhenEnvRelativePath_ThenReturnsAbsPath(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	executableName := "codex-mock"
	executablePath := filepath.Join(tempDir, executableName)
	require.NoError(t, os.WriteFile(executablePath, []byte("binary"), 0o600))

	t.Setenv(envExecutablePath, executableName)
	t.Setenv("PATH", "")

	path, err := locateExecutable(context.Background())

	require.NoError(t, err)
	expectedPath, err := filepath.EvalSymlinks(executablePath)
	require.NoError(t, err)

	actualPath, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)

	require.Equal(t, expectedPath, actualPath)
}

func TestLocateExecutable_WhenEnvPathMissing_ThenReturnsError(t *testing.T) {
	t.Setenv(envExecutablePath, filepath.Join(t.TempDir(), "missing"))
	t.Setenv("PATH", "")

	_, err := locateExecutable(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "executable from "+envExecutablePath+" not found")
}

func TestLocateExecutable_WhenPathLookupSucceeds_ThenReturnsPath(t *testing.T) {
	tempDir := t.TempDir()
	executablePath := createExecutable(t, tempDir, "codex")

	t.Setenv(envExecutablePath, "")
	t.Setenv("PATH", tempDir)

	path, err := locateExecutable(context.Background())

	require.NoError(t, err)
	require.Equal(t, executablePath, path)
}

func TestLocateExecutable_WhenPathLookupFails_ThenReturnsError(t *testing.T) {
	t.Setenv(envExecutablePath, "")
	t.Setenv("PATH", "")

	_, err := locateExecutable(context.Background())

	require.Error(t, err)
	require.ErrorContains(t, err, "lookup codex executable")
	require.ErrorIs(t, err, exec.ErrNotFound)
}

func createExecutable(t *testing.T, dir string, name string) string {
	t.Helper()

	target, err := os.Executable()
	require.NoError(t, err)

	path := filepath.Join(dir, name)
	require.NoError(t, os.Link(target, path))

	return path
}
