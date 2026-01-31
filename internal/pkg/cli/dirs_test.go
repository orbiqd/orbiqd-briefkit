package cli

import (
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
