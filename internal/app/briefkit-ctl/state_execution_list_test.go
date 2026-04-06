package briefkitctl

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	fsstore "github.com/orbiqd/orbiqd-briefkit/internal/pkg/store/fs"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateExecutionListCmd_Run_WhenFindFails_ThenReturnsError(t *testing.T) {
	ctx := context.Background()
	repo := briefkit.NewMockExecutionRepository(t)
	cmd := &StateExecutionListCmd{}
	expectedErr := errors.New("storage error")

	repo.EXPECT().Find(ctx).Return(nil, expectedErr)

	err := cmd.Run(ctx, repo)

	require.Error(t, err)
	require.ErrorContains(t, err, "list executions")
	assert.ErrorIs(t, err, expectedErr)
}

func TestStateExecutionListCmd_Run_WhenGetExecutionFails_ThenSkipsAndContinues(t *testing.T) {
	ctx := context.Background()
	repo := briefkit.NewMockExecutionRepository(t)
	successExec := briefkit.NewMockExecution(t)
	cmd := &StateExecutionListCmd{}

	failID := briefkit.ExecutionID("fail-id")
	successID := briefkit.ExecutionID("success-id")
	status := briefkit.ExecutionStatus{State: briefkit.ExecutionCreated}

	repo.EXPECT().Find(ctx).Return([]briefkit.ExecutionID{failID, successID}, nil)
	repo.EXPECT().Get(ctx, failID).Return(nil, errors.New("not found"))
	repo.EXPECT().Get(ctx, successID).Return(successExec, nil)
	successExec.EXPECT().GetStatus(ctx).Return(status, nil)

	output := captureStdout(t, func() {
		err := cmd.Run(ctx, repo)
		require.NoError(t, err)
	})

	var result ExecutionListOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, successID, result.Items[0].Id)
}

func TestStateExecutionListCmd_Run_WhenGetStatusFails_ThenSkipsAndContinues(t *testing.T) {
	ctx := context.Background()
	repo := briefkit.NewMockExecutionRepository(t)
	failExec := briefkit.NewMockExecution(t)
	successExec := briefkit.NewMockExecution(t)
	cmd := &StateExecutionListCmd{}

	failID := briefkit.ExecutionID("fail-id")
	successID := briefkit.ExecutionID("success-id")
	status := briefkit.ExecutionStatus{State: briefkit.ExecutionCreated}

	repo.EXPECT().Find(ctx).Return([]briefkit.ExecutionID{failID, successID}, nil)
	repo.EXPECT().Get(ctx, failID).Return(failExec, nil)
	repo.EXPECT().Get(ctx, successID).Return(successExec, nil)
	failExec.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{}, errors.New("corrupt status"))
	successExec.EXPECT().GetStatus(ctx).Return(status, nil)

	output := captureStdout(t, func() {
		err := cmd.Run(ctx, repo)
		require.NoError(t, err)
	})

	var result ExecutionListOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 1, result.Count)
	assert.Equal(t, successID, result.Items[0].Id)
}

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
