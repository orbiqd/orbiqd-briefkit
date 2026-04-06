package fs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
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

	id := briefkit.AgentID("codex-1")
	config := briefkit.Config{
		Runtime: struct {
			Kind    briefkit.RuntimeKind     `json:"kind"`
			Config  briefkit.RuntimeConfig   `json:"config"`
			Feature briefkit.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: briefkit.RuntimeKind("codex"),
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

func TestConfigRepository_UpdateGet_WhenConfigHasTimeout_ThenRoundTripsCorrectly(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	id := briefkit.AgentID("codex-timeout")
	configTimeout := utils.Duration(10 * time.Minute)
	config := briefkit.Config{
		Runtime: struct {
			Kind    briefkit.RuntimeKind     `json:"kind"`
			Config  briefkit.RuntimeConfig   `json:"config"`
			Feature briefkit.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: briefkit.RuntimeKind("codex"),
		},
		Timeout: &configTimeout,
	}

	err = repo.Update(ctx, id, config)
	require.NoError(t, err)

	loaded, err := repo.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, loaded.Timeout)
	assert.Equal(t, 10*time.Minute, time.Duration(*loaded.Timeout))
}

func TestConfigRepository_UpdateGet_WhenConfigWithoutTimeout_ThenTimeoutIsNil(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	id := briefkit.AgentID("codex-no-timeout")
	config := briefkit.Config{
		Runtime: struct {
			Kind    briefkit.RuntimeKind     `json:"kind"`
			Config  briefkit.RuntimeConfig   `json:"config"`
			Feature briefkit.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: briefkit.RuntimeKind("codex"),
		},
	}

	err = repo.Update(ctx, id, config)
	require.NoError(t, err)

	loaded, err := repo.Get(ctx, id)
	require.NoError(t, err)
	assert.Nil(t, loaded.Timeout)
}

func TestConfigRepository_Exists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	id := briefkit.AgentID("codex")
	config := briefkit.Config{
		Runtime: struct {
			Kind    briefkit.RuntimeKind     `json:"kind"`
			Config  briefkit.RuntimeConfig   `json:"config"`
			Feature briefkit.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: briefkit.RuntimeKind("codex"),
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
		exists, err := repo.Exists(ctx, briefkit.AgentID("codex-2"))
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("invalid id", func(t *testing.T) {
		exists, err := repo.Exists(ctx, briefkit.AgentID("Codex"))
		require.ErrorIs(t, err, briefkit.ErrAgentIDInvalid)
		assert.False(t, exists)
	})
}

func TestConfigRepository_Update(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("invalid id", func(t *testing.T) {
		err := repo.Update(ctx, briefkit.AgentID("Invalid"), briefkit.Config{})
		assert.ErrorIs(t, err, briefkit.ErrAgentIDInvalid)
	})
}

func TestConfigRepository_Get(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-agents"
	repo, err := NewConfigRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("missing config", func(t *testing.T) {
		_, err := repo.Get(ctx, briefkit.AgentID("codex"))
		assert.ErrorIs(t, err, briefkit.ErrAgentConfigNotFound)
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := repo.Get(ctx, briefkit.AgentID("Codex"))
		assert.ErrorIs(t, err, briefkit.ErrAgentIDInvalid)
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

	require.NoError(t, repo.Update(ctx, briefkit.AgentID("codex"), briefkit.Config{
		Runtime: struct {
			Kind    briefkit.RuntimeKind     `json:"kind"`
			Config  briefkit.RuntimeConfig   `json:"config"`
			Feature briefkit.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: briefkit.RuntimeKind("codex"),
		},
	}))
	require.NoError(t, repo.Update(ctx, briefkit.AgentID("claude-code"), briefkit.Config{
		Runtime: struct {
			Kind    briefkit.RuntimeKind     `json:"kind"`
			Config  briefkit.RuntimeConfig   `json:"config"`
			Feature briefkit.RuntimeFeatures `json:"feature,omitempty"`
		}{
			Kind: briefkit.RuntimeKind("claude-code"),
		},
	}))
	require.NoError(t, afero.WriteFile(memFs, filepath.Join(basePath, "readme.txt"), []byte("ignore"), 0644))
	require.NoError(t, afero.WriteFile(memFs, filepath.Join(basePath, "Bad.yaml"), []byte("kind: codex"), 0644))
	require.NoError(t, memFs.MkdirAll(filepath.Join(basePath, "subdir.yaml"), 0755))

	ids, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, []briefkit.AgentID{"claude-code", "codex"}, ids)
}

func TestNewConfigRepository_WhenMkdirAllFails_ThenReturnsError(t *testing.T) {
	readOnlyFs := afero.NewReadOnlyFs(afero.NewMemMapFs())

	repo, err := NewConfigRepository("/tmp/new-path", readOnlyFs)

	require.Error(t, err)
	require.ErrorContains(t, err, "create agent config path")
	assert.Nil(t, repo)
}

func TestConfigRepository_Update_WhenWriteFails_ThenReturnsError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	repo, err := NewConfigRepository("/tmp/test-agents", memFs)
	require.NoError(t, err)
	repo.fs = afero.NewReadOnlyFs(memFs)
	ctx := context.Background()

	err = repo.Update(ctx, briefkit.AgentID("codex"), briefkit.Config{})

	require.Error(t, err)
}

func TestConfigRepository_Get_WhenYAMLCorrupt_ThenReturnsError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	repo, err := NewConfigRepository("/tmp/test-agents", memFs)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, afero.WriteFile(memFs, "/tmp/test-agents/codex.yaml", []byte(": invalid: yaml: {{"), 0o600))

	_, err = repo.Get(ctx, briefkit.AgentID("codex"))

	require.Error(t, err)
}
