package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateExecutionRepositoryFromConfig_AbsolutePath_Success(t *testing.T) {
	tmpDir := t.TempDir()
	config := StoreConfig{
		StatePath: tmpDir,
	}

	repo, err := CreateExecutionRepositoryFromConfig(config)

	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestCreateExecutionRepositoryFromConfig_TildePath_Success(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	tmpDir := t.TempDir()
	relPath, err := filepath.Rel(homeDir, tmpDir)
	if err != nil || filepath.IsAbs(relPath) || relPath[:2] == ".." {
		t.Skip("temp dir not under home directory")
	}

	config := StoreConfig{
		StatePath: "~/" + relPath,
	}

	repo, err := CreateExecutionRepositoryFromConfig(config)

	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestCreateExecutionRepositoryFromConfig_RelativePath_ReturnsError(t *testing.T) {
	config := StoreConfig{
		StatePath: "relative/path",
	}

	repo, err := CreateExecutionRepositoryFromConfig(config)

	require.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "state path must be absolute")
}

func TestCreateConfigRepositoryFromConfig_AbsolutePath_Success(t *testing.T) {
	tmpDir := t.TempDir()
	config := StoreConfig{
		AgentConfigPath: tmpDir,
	}

	repo, err := CreateConfigRepositoryFromConfig(config)

	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestCreateConfigRepositoryFromConfig_TildePath_Success(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	tmpDir := t.TempDir()
	relPath, err := filepath.Rel(homeDir, tmpDir)
	if err != nil || filepath.IsAbs(relPath) || relPath[:2] == ".." {
		t.Skip("temp dir not under home directory")
	}

	config := StoreConfig{
		AgentConfigPath: "~/" + relPath,
	}

	repo, err := CreateConfigRepositoryFromConfig(config)

	require.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestCreateConfigRepositoryFromConfig_RelativePath_ReturnsError(t *testing.T) {
	config := StoreConfig{
		AgentConfigPath: "relative/path",
	}

	repo, err := CreateConfigRepositoryFromConfig(config)

	require.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "agent config path must be absolute")
}
