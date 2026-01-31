package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRuntimeLogDir_EnvNotSet_UsesDefault(t *testing.T) {
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", "")

	dir, err := ResolveRuntimeLogDir()

	require.NoError(t, err)
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	expectedPath := filepath.Join(homeDir, ".orbiqd", "briefkit", "logs", "runtime")
	assert.Equal(t, expectedPath, dir)
}

func TestResolveRuntimeLogDir_EnvSet_UsesEnvValue(t *testing.T) {
	customDir := "/custom/log/dir"
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", customDir)

	dir, err := ResolveRuntimeLogDir()

	require.NoError(t, err)
	assert.Equal(t, customDir, dir)
}

func TestResolveRuntimeLogDir_EnvWithTilde_ExpandsPath(t *testing.T) {
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", "~/custom/logs")

	dir, err := ResolveRuntimeLogDir()

	require.NoError(t, err)
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	expectedPath := filepath.Join(homeDir, "custom", "logs")
	assert.Equal(t, expectedPath, dir)
}

func TestResolveRuntimeLogDir_RelativePath_ReturnsAbsolute(t *testing.T) {
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", "relative/path")

	dir, err := ResolveRuntimeLogDir()

	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(dir), "expected absolute path, got: %s", dir)
}

func TestResolveExecutable_WhenCalledViaSymlink_ThenPreservesSymlinkPath(t *testing.T) {
	tmpDir := t.TempDir()

	realDir := filepath.Join(tmpDir, "real")
	symlinkDir := filepath.Join(tmpDir, "symlink")

	require.NoError(t, os.MkdirAll(realDir, 0o750))

	realExecutable := filepath.Join(realDir, "briefkit-ctl")
	targetExecutable := filepath.Join(realDir, ExecutableMCP)

	require.NoError(t, os.WriteFile(realExecutable, []byte("#!/bin/sh\necho test"), 0o600))
	require.NoError(t, os.WriteFile(targetExecutable, []byte("#!/bin/sh\necho mcp"), 0o600))

	require.NoError(t, os.Symlink(realDir, symlinkDir))

	symlinkExecutable := filepath.Join(symlinkDir, "briefkit-ctl")

	originalOsExecutable := osExecutable
	osExecutable = func() (string, error) {
		return symlinkExecutable, nil
	}
	t.Cleanup(func() {
		osExecutable = originalOsExecutable
	})

	result, err := ResolveExecutable(context.Background(), ExecutableMCP)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(symlinkDir, ExecutableMCP), result,
		"should return path using symlink directory, not resolved real path")
}
