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

func TestStateExecutionListCmd_Run_EmptyRepository_ReturnsEmptyList(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	ctx := context.Background()
	cmd := &StateExecutionListCmd{}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, repo)
	})

	require.NoError(t, err)
	var result ExecutionListOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Empty(t, result.Items)
	assert.Equal(t, 0, result.Count)
}

func TestStateExecutionListCmd_Run_WithExecutions_ReturnsExecutionList(t *testing.T) {
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

	id1, err := repo.Create(ctx, input, agentConfig)
	require.NoError(t, err)

	id2, err := repo.Create(ctx, input, agentConfig)
	require.NoError(t, err)

	cmd := &StateExecutionListCmd{}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, repo)
	})

	require.NoError(t, err)
	var result ExecutionListOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.Count)

	ids := []briefkit.ExecutionID{result.Items[0].Id, result.Items[1].Id}
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
	assert.Equal(t, briefkit.ExecutionCreated, result.Items[0].Status.State)
	assert.Equal(t, briefkit.ExecutionCreated, result.Items[1].Status.State)
}
