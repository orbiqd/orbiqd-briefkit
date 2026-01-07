package fs

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigRepository(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"

	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, basePath, repo.basePath)
	assert.Equal(t, memFs, repo.fs)

	exists, err := afero.DirExists(memFs, basePath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestConfigRepository_UpdateGet(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	id := agent.AgentID("codex-1")
	config := agent.Config{
		Runtime: struct {
			Kind    agent.RuntimeKind     `json:"kind"`
			Config  agent.RuntimeConfig   `json:"config"`
			Feature agent.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: agent.RuntimeKind("codex"),
			Config: map[string]any{
				"path": "/bin/codex",
			},
		},
	}

	err = repo.Update(ctx, id, config)
	require.NoError(t, err)

	filePath := filepath.Join(basePath, "codex-1.yaml")
	exists, err := afero.Exists(memFs, filePath)
	require.NoError(t, err)
	assert.True(t, exists)

	loaded, err := repo.Get(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, config.Runtime.Kind, loaded.Runtime.Kind)
	assert.Equal(t, config.Runtime.Config, loaded.Runtime.Config)
}

func TestConfigRepository_Exists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	id := agent.AgentID("codex")
	config := agent.Config{
		Runtime: struct {
			Kind    agent.RuntimeKind     `json:"kind"`
			Config  agent.RuntimeConfig   `json:"config"`
			Feature agent.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: agent.RuntimeKind("codex"),
			Config: map[string]any{
				"path": "/bin/codex",
			},
		},
	}

	err = repo.Update(ctx, id, config)
	require.NoError(t, err)

	t.Run("existing config", func(t *testing.T) {
		exists, err := repo.Exists(ctx, id)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("missing config", func(t *testing.T) {
		exists, err := repo.Exists(ctx, agent.AgentID("codex-2"))
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("invalid id", func(t *testing.T) {
		exists, err := repo.Exists(ctx, agent.AgentID("Codex"))
		assert.ErrorIs(t, err, agent.ErrAgentIDInvalid)
		assert.False(t, exists)
	})
}

func TestConfigRepository_Get(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("missing config", func(t *testing.T) {
		_, err := repo.Get(ctx, agent.AgentID("codex"))
		assert.ErrorIs(t, err, agent.ErrAgentConfigNotFound)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := repo.Get(ctx, agent.AgentID("Codex"))
		assert.ErrorIs(t, err, agent.ErrAgentIDInvalid)
	})
}

func TestConfigRepository_List(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	ids, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, ids)

	require.NoError(t, repo.Update(ctx, agent.AgentID("codex"), agent.Config{
		Runtime: struct {
			Kind    agent.RuntimeKind     `json:"kind"`
			Config  agent.RuntimeConfig   `json:"config"`
			Feature agent.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: agent.RuntimeKind("codex"),
		},
	}))
	require.NoError(t, repo.Update(ctx, agent.AgentID("claude-code"), agent.Config{
		Runtime: struct {
			Kind    agent.RuntimeKind     `json:"kind"`
			Config  agent.RuntimeConfig   `json:"config"`
			Feature agent.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: agent.RuntimeKind("claude-code"),
		},
	}))
	require.NoError(t, afero.WriteFile(memFs, filepath.Join(basePath, "readme.txt"), []byte("ignore"), 0644))
	require.NoError(t, afero.WriteFile(memFs, filepath.Join(basePath, "Bad.yaml"), []byte("kind: codex"), 0644))

	ids, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, []agent.AgentID{"claude-code", "codex"}, ids)
}
