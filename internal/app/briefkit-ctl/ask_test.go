package briefkitctl

import (
	"context"
	"testing"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	fsstore "github.com/orbiqd/orbiqd-briefkit/internal/pkg/store/fs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskCmd_Run_WhenAgentConfigMissing_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	execRepo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	configRepo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()

	cmd := &AskCmd{
		AgentID: "missing-agent",
		Timeout: 5 * time.Minute,
		Prompt:  "test prompt",
	}

	err = cmd.Run(ctx, execRepo, configRepo)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent config does not exist")
}

func TestAskCmd_Run_WhenExecutionSucceeds_PrintsResponse(t *testing.T) {
	fs := afero.NewMemMapFs()
	execRepo, err := fsstore.NewExecutionRepository("/tmp/executions", fs)
	require.NoError(t, err)
	configRepo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()

	agentConfig := agent.Config{}
	agentConfig.Runtime.Kind = "codex"
	require.NoError(t, configRepo.Update(ctx, "codex", agentConfig))

	originalSpawnRunner := spawnRunner
	t.Cleanup(func() {
		spawnRunner = originalSpawnRunner
	})

	spawnRunner = func(ctx context.Context, executionID agent.ExecutionID) error {
		execution, err := execRepo.Get(ctx, executionID)
		if err != nil {
			return err
		}

		return execution.SetResult(ctx, agent.ExecutionResult{
			ConversationID: "conversation-1",
			Response:       "done",
		})
	}

	cmd := &AskCmd{
		AgentID: "codex",
		Timeout: time.Minute,
		Prompt:  "test prompt",
	}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, execRepo, configRepo)
	})

	require.NoError(t, err)
	assert.Contains(t, string(output), "done")
}
