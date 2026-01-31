package briefkitctl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	fsstore "github.com/orbiqd/orbiqd-briefkit/internal/pkg/store/fs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateExecutionCreateCmd_Run_MissingAgent_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	execRepo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	configRepo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()

	cmd := &StateExecutionCreateCmd{
		AgentID:    "nonexistent",
		Prompt:     "test prompt",
		WorkingDir: t.TempDir(),
		Timeout:    "5m",
	}

	err = cmd.Run(ctx, execRepo, configRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load agent config")
}

func TestStateExecutionCreateCmd_Run_InvalidTimeout_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	execRepo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	configRepo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()

	agentConfig := agent.Config{}
	agentConfig.Runtime.Kind = "codex"
	require.NoError(t, configRepo.Update(ctx, "codex", agentConfig))

	cmd := &StateExecutionCreateCmd{
		AgentID:    "codex",
		Prompt:     "test prompt",
		WorkingDir: t.TempDir(),
		Timeout:    "invalid",
	}

	err = cmd.Run(ctx, execRepo, configRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse timeout")
}

func TestStateExecutionCreateCmd_Run_ValidInput_CreatesExecution(t *testing.T) {
	fs := afero.NewMemMapFs()
	execRepo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	configRepo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()

	agentConfig := agent.Config{}
	agentConfig.Runtime.Kind = "codex"
	require.NoError(t, configRepo.Update(ctx, "codex", agentConfig))

	workingDir := t.TempDir()
	cmd := &StateExecutionCreateCmd{
		AgentID:    "codex",
		Prompt:     "test prompt",
		WorkingDir: workingDir,
		Timeout:    "5m",
	}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, execRepo, configRepo)
	})

	require.NoError(t, err)

	var result executionCreateOutput
	require.NoError(t, json.Unmarshal(output, &result))
	require.NoError(t, result.ID.Validate())

	execution, err := execRepo.Get(ctx, result.ID)
	require.NoError(t, err)

	input, err := execution.GetInput(ctx)
	require.NoError(t, err)
	assert.Equal(t, "test prompt", input.Prompt)
	assert.Equal(t, workingDir, *input.WorkingDirectory)
}
