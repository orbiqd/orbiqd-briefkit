package briefkitctl

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	fsstore "github.com/orbiqd/orbiqd-briefkit/internal/pkg/store/fs"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentListCmd_Run_EmptyRepository_ReturnsEmptyList(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()
	cmd := &AgentListCmd{}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, repo)
	})

	require.NoError(t, err)
	var result AgentListOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Empty(t, result.Items)
	assert.Equal(t, 0, result.Count)
}

func TestAgentListCmd_Run_WithAgents_ReturnsAgentList(t *testing.T) {
	fs := afero.NewMemMapFs()
	repo, err := fsstore.NewConfigRepository("/tmp/agents", fs)
	require.NoError(t, err)
	ctx := context.Background()

	config1 := briefkit.Config{}
	config1.Runtime.Kind = "claude"
	require.NoError(t, repo.Update(ctx, "claude", config1))

	config2 := briefkit.Config{}
	config2.Runtime.Kind = "codex"
	require.NoError(t, repo.Update(ctx, "codex", config2))

	cmd := &AgentListCmd{}

	output := captureStdout(t, func() {
		err = cmd.Run(ctx, repo)
	})

	require.NoError(t, err)
	var result AgentListOutput
	require.NoError(t, json.Unmarshal(output, &result))
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 2, result.Count)
	assert.Equal(t, briefkit.AgentID("claude"), result.Items[0].ID)
	assert.Equal(t, briefkit.RuntimeKind("claude"), result.Items[0].RuntimeKind)
	assert.Equal(t, briefkit.AgentID("codex"), result.Items[1].ID)
	assert.Equal(t, briefkit.RuntimeKind("codex"), result.Items[1].RuntimeKind)
}

func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.Bytes()
}
