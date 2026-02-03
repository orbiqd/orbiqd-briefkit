package briefkitctl

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	fsstore "github.com/orbiqd/orbiqd-briefkit/internal/pkg/store/fs"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateExecutionShowCmd_Run_InvalidID_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	ctx := context.Background()

	cmd := &StateExecutionShowCmd{ID: "invalid-id"}

	err = cmd.Run(ctx, repo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "validate execution id")
}

func TestStateExecutionShowCmd_Run_NonExistentID_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	ctx := context.Background()

	cmd := &StateExecutionShowCmd{ID: "00000000-0000-0000-0000-000000000001"}

	err = cmd.Run(ctx, repo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load execution")
}

func TestStateExecutionShowCmd_Run_ExistingExecution_ReturnsDetails(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	ctx := context.Background()

	agentConfig := briefkit.Config{}
	agentConfig.Runtime.Kind = "codex"

	input := briefkit.ExecutionInput{
		Prompt:  "test prompt",
		Timeout: utils.Duration(5 * time.Minute),
	}

	id, err := repo.Create(ctx, input, agentConfig)
	require.NoError(t, err)

	cmd := &StateExecutionShowCmd{ID: string(id)}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, repo)
	})

	require.NoError(t, err)
	var result ExecutionShowOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Equal(t, briefkit.ExecutionCreated, result.Status.State)
	assert.Equal(t, "test prompt", result.Input.Prompt)
	assert.Nil(t, result.Result)
}

func TestStateExecutionShowCmd_Run_ExecutionWithResult_ReturnsResult(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	ctx := context.Background()

	agentConfig := briefkit.Config{}
	agentConfig.Runtime.Kind = "codex"

	input := briefkit.ExecutionInput{
		Prompt:  "test prompt",
		Timeout: utils.Duration(5 * time.Minute),
	}

	id, err := repo.Create(ctx, input, agentConfig)
	require.NoError(t, err)

	execution, err := repo.Get(ctx, id)
	require.NoError(t, err)

	result := briefkit.ExecutionResult{
		Response:       "test response",
		ConversationID: "conv-123",
	}
	require.NoError(t, execution.SetResult(ctx, result))

	cmd := &StateExecutionShowCmd{ID: string(id)}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, repo)
	})

	require.NoError(t, err)
	var showOutput ExecutionShowOutput
	require.NoError(t, json.Unmarshal(output, &showOutput))
	assert.Equal(t, briefkit.ExecutionSucceeded, showOutput.Status.State)
	require.NotNil(t, showOutput.Result)
	assert.Equal(t, "test response", showOutput.Result.Response)
	assert.Equal(t, briefkit.ConversationID("conv-123"), showOutput.Result.ConversationID)
}
