package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkspaceManagerFromConfig_WhenValidStatePath_ThenReturnsDirScheme(t *testing.T) {
	config := StoreConfig{StatePath: t.TempDir()}

	manager, err := CreateWorkspaceManagerFromConfig(config)

	require.NoError(t, err)
	assert.Contains(t, manager.Schemes(), "dir")
}

func TestCreateWorkspaceManagerFromConfig_WhenStatePathNotAbsolute_ThenReturnsError(t *testing.T) {
	config := StoreConfig{StatePath: "relative/path"}

	_, err := CreateWorkspaceManagerFromConfig(config)

	require.Error(t, err)
}

func TestCreateWorkspaceManagerFromConfig_WhenGitAvailable_ThenRegistersGitSchemes(t *testing.T) {
	config := StoreConfig{StatePath: t.TempDir()}

	manager, err := CreateWorkspaceManagerFromConfig(config)

	require.NoError(t, err)
	schemes := manager.Schemes()
	assert.Contains(t, schemes, "git+https")
	assert.Contains(t, schemes, "git+ssh")
}
