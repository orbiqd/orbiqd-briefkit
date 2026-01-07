package fs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleAgentConfig = agent.Config{
	RuntimeKind: agent.RuntimeKind("codex"),
	RuntimeConfig: map[string]any{
		"path": "/bin/agent",
	},
}

func TestNewExecutionRepository(t *testing.T) {
	// Use an in-memory filesystem for testing
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-executions"

	repo, err := NewExecutionRepository(basePath, memFs)
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, basePath, repo.basePath)
	assert.Equal(t, memFs, repo.fs)

	exists, err := afero.DirExists(memFs, basePath)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRepository_Create(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-executions"
	repo, err := NewExecutionRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()
	workingDir := "/app"

	// Prepare a valid input
	input := agent.ExecutionInput{
		Prompt:           "test prompt",
		Timeout:          utils.Duration(5 * time.Minute),
		WorkingDirectory: &workingDir,
	}

	id, err := repo.Create(ctx, input, sampleAgentConfig)
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.NoError(t, id.Validate())

	executionPath := filepath.Join(basePath, string(id))
	dirExists, err := afero.DirExists(memFs, executionPath)
	require.NoError(t, err)
	assert.True(t, dirExists, "execution directory should exist")

	fileExists, err := afero.Exists(memFs, filepath.Join(executionPath, executionInputFileName))
	require.NoError(t, err)
	assert.True(t, fileExists, "input.json should exist")

	fileExists, err = afero.Exists(memFs, filepath.Join(executionPath, executionStatusFileName))
	require.NoError(t, err)
	assert.True(t, fileExists, "status.json should exist")

	fileExists, err = afero.Exists(memFs, filepath.Join(executionPath, executionAgentConfigFileName))
	require.NoError(t, err)
	assert.True(t, fileExists, "agent.json should exist")

	// Verify content of input.json
	retrievedInput, err := readJSON[agent.ExecutionInput](memFs, filepath.Join(executionPath, executionInputFileName))
	require.NoError(t, err)
	assert.Equal(t, input.Prompt, retrievedInput.Prompt)
	assert.Equal(t, input.Timeout, retrievedInput.Timeout)
	assert.Equal(t, input.WorkingDirectory, retrievedInput.WorkingDirectory)

	// Verify content of agent.json
	retrievedConfig, err := readJSON[agent.Config](memFs, filepath.Join(executionPath, executionAgentConfigFileName))
	require.NoError(t, err)
	assert.Equal(t, sampleAgentConfig, retrievedConfig)

	// Verify content of status.json
	retrievedStatus, err := readJSON[agent.ExecutionStatus](memFs, filepath.Join(executionPath, executionStatusFileName))
	require.NoError(t, err)
	assert.Equal(t, agent.ExecutionCreated, retrievedStatus.State)
	assert.WithinDuration(t, time.Now(), retrievedStatus.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), retrievedStatus.UpdatedAt, time.Second)
	assert.Nil(t, retrievedStatus.FinishedAt)
	assert.Equal(t, 0, retrievedStatus.Attempts)

	t.Run("invalid input", func(t *testing.T) {
		invalidInput := input
		invalidInput.Prompt = "" // Invalid prompt
		id, err := repo.Create(ctx, invalidInput, sampleAgentConfig)
		require.Error(t, err)
		assert.Equal(t, agent.EmptyExecutionID, id)
		assert.ErrorIs(t, err, agent.ErrExecutionPromptRequired)
	})
}

func TestRepository_Exists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-executions"
	repo, err := NewExecutionRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()
	workingDir := "/app"

	input := agent.ExecutionInput{
		Prompt:           "test prompt",
		Timeout:          utils.Duration(5 * time.Minute),
		WorkingDirectory: &workingDir,
	}

	id, err := repo.Create(ctx, input, sampleAgentConfig)
	require.NoError(t, err)

	t.Run("existing execution", func(t *testing.T) {
		exists, err := repo.Exists(ctx, id)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing execution", func(t *testing.T) {
		nonExistingID := agent.ExecutionID("non-existing-uuid") // Not a valid UUID, but Exists should handle it
		exists, err := repo.Exists(ctx, nonExistingID)
		require.Error(t, err) // Should return error because ID validation fails
		assert.False(t, exists)

		// Test with valid but non-existing UUID
		validNonExistingID := agent.NewExecutionID()
		exists, err = repo.Exists(ctx, validNonExistingID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("invalid ID", func(t *testing.T) {
		invalidID := agent.ExecutionID("invalid")
		exists, err := repo.Exists(ctx, invalidID)
		require.ErrorIs(t, err, agent.ErrExecutionIDInvalid)
		assert.False(t, exists)
	})
}

func TestRepository_Get(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-executions"
	repo, err := NewExecutionRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()
	workingDir := "/app"

	input := agent.ExecutionInput{
		Prompt:           "test prompt",
		Timeout:          utils.Duration(5 * time.Minute), // Using the correct Duration type
		WorkingDirectory: &workingDir,
	}

	id, err := repo.Create(ctx, input, sampleAgentConfig)
	require.NoError(t, err)

	t.Run("existing execution", func(t *testing.T) {
		exec, err := repo.Get(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, exec)

		// Verify GetInput
		retrievedInput, err := exec.GetInput(ctx)
		require.NoError(t, err)
		assert.Equal(t, input.Prompt, retrievedInput.Prompt)
		assert.Equal(t, input.Timeout, retrievedInput.Timeout)
		assert.Equal(t, input.WorkingDirectory, retrievedInput.WorkingDirectory)

		// Verify GetStatus
		retrievedStatus, err := exec.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, agent.ExecutionCreated, retrievedStatus.State)

		retrievedConfig, err := exec.GetAgentConfig(ctx)
		require.NoError(t, err)
		assert.Equal(t, sampleAgentConfig, retrievedConfig)
	})

	t.Run("non-existing execution", func(t *testing.T) {
		nonExistingID := agent.NewExecutionID()
		exec, err := repo.Get(ctx, nonExistingID)
		require.ErrorIs(t, err, agent.ErrExecutionNotFound)
		assert.Nil(t, exec)
	})

	t.Run("invalid ID", func(t *testing.T) {
		invalidID := agent.ExecutionID("invalid")
		exec, err := repo.Get(ctx, invalidID)
		require.ErrorIs(t, err, agent.ErrExecutionIDInvalid)
		assert.Nil(t, exec)
	})
}

func TestRepository_Find(t *testing.T) {
	t.Run("empty base path returns empty slice", func(t *testing.T) {
		memFs := afero.NewMemMapFs()
		basePath := "/tmp/test-executions"
		repo, err := NewExecutionRepository(basePath, memFs)
		require.NoError(t, err)

		ids, err := repo.Find(context.Background())
		require.NoError(t, err)
		assert.Empty(t, ids)
	})

	t.Run("returns sorted ids and ignores invalid entries", func(t *testing.T) {
		memFs := afero.NewMemMapFs()
		basePath := "/tmp/test-executions"
		repo, err := NewExecutionRepository(basePath, memFs)
		require.NoError(t, err)
		ctx := context.Background()

		validIDs := []agent.ExecutionID{
			agent.ExecutionID("00000000-0000-0000-0000-000000000002"),
			agent.ExecutionID("00000000-0000-0000-0000-000000000001"),
		}

		for _, id := range validIDs {
			err := memFs.MkdirAll(filepath.Join(basePath, string(id)), 0755)
			require.NoError(t, err)
		}

		err = memFs.MkdirAll(filepath.Join(basePath, "not-a-uuid"), 0755)
		require.NoError(t, err)

		err = afero.WriteFile(memFs, filepath.Join(basePath, "00000000-0000-0000-0000-000000000003"), []byte("x"), 0644)
		require.NoError(t, err)

		err = afero.WriteFile(memFs, filepath.Join(basePath, "random-file"), []byte("x"), 0644)
		require.NoError(t, err)

		ids, err := repo.Find(ctx)
		require.NoError(t, err)
		require.Len(t, ids, 2)
		assert.Equal(t, agent.ExecutionID("00000000-0000-0000-0000-000000000001"), ids[0])
		assert.Equal(t, agent.ExecutionID("00000000-0000-0000-0000-000000000002"), ids[1])
	})
}

func TestExecution_GetSetResult(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-executions"
	repo, err := NewExecutionRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()
	workingDir := "/app"

	input := agent.ExecutionInput{
		Prompt:           "test prompt",
		Timeout:          utils.Duration(5 * time.Minute),
		WorkingDirectory: &workingDir,
	}

	id, err := repo.Create(ctx, input, sampleAgentConfig)
	require.NoError(t, err)
	exec, err := repo.Get(ctx, id)
	require.NoError(t, err)

	t.Run("initial state has no result", func(t *testing.T) {
		hasResult, err := exec.HasResult(ctx)
		require.NoError(t, err)
		assert.False(t, hasResult)

		_, err = exec.GetResult(ctx)
		require.ErrorIs(t, err, agent.ErrExecutionNoResult)
	})

	result := agent.ExecutionResult{} // Empty result for now

	t.Run("set result", func(t *testing.T) {
		err := exec.SetResult(ctx, result)
		require.NoError(t, err)

		hasResult, err := exec.HasResult(ctx)
		require.NoError(t, err)
		assert.True(t, hasResult)

		retrievedResult, err := exec.GetResult(ctx)
		require.NoError(t, err)
		assert.Equal(t, result, retrievedResult)

		// Verify status update
		status, err := exec.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, agent.ExecutionSucceeded, status.State)
		assert.NotNil(t, status.FinishedAt)
		assert.WithinDuration(t, time.Now(), *status.FinishedAt, time.Second)
		assert.WithinDuration(t, time.Now(), status.UpdatedAt, time.Second)
	})

	t.Run("can set result twice", func(t *testing.T) {
		updatedResult := agent.ExecutionResult{Response: "updated response"}

		err := exec.SetResult(ctx, updatedResult)
		require.NoError(t, err)

		retrievedResult, err := exec.GetResult(ctx)
		require.NoError(t, err)
		assert.Equal(t, updatedResult, retrievedResult)
	})
}

func TestExecution_UpdateStatus(t *testing.T) {
	memFs := afero.NewMemMapFs()
	basePath := "/tmp/test-executions"
	repo, err := NewExecutionRepository(basePath, memFs)
	require.NoError(t, err)
	ctx := context.Background()
	workingDir := "/app"

	input := agent.ExecutionInput{
		Prompt:           "test prompt",
		Timeout:          utils.Duration(5 * time.Minute),
		WorkingDirectory: &workingDir,
	}

	id, err := repo.Create(ctx, input, sampleAgentConfig)
	require.NoError(t, err)
	exec, err := repo.Get(ctx, id)
	require.NoError(t, err)

	t.Run("update status from created to running", func(t *testing.T) {
		status, err := exec.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, agent.ExecutionCreated, status.State)

		status.State = agent.ExecutionRunning
		err = exec.UpdateStatus(ctx, status)
		require.NoError(t, err)

		updatedStatus, err := exec.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, agent.ExecutionRunning, updatedStatus.State)
		assert.WithinDuration(t, time.Now(), updatedStatus.UpdatedAt, time.Second)
		assert.True(t, updatedStatus.UpdatedAt.After(status.UpdatedAt)) // UpdatedAt should be newer
	})

	t.Run("update status to started", func(t *testing.T) {
		status, err := exec.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, agent.ExecutionRunning, status.State)

		status.State = agent.ExecutionStarted
		err = exec.UpdateStatus(ctx, status)
		require.NoError(t, err)

		updatedStatus, err := exec.GetStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, agent.ExecutionStarted, updatedStatus.State)
	})
}
