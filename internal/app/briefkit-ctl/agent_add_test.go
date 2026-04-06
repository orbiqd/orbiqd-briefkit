package briefkitctl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentAddCmd_Run_WhenCalled_ThenReturnsNotImplementedError(t *testing.T) {
	cmd := &AgentAddCmd{
		ID:   "my-agent",
		Kind: "codex",
		Path: "/usr/local/bin/codex",
	}

	err := cmd.Run()

	require.Error(t, err)
	assert.EqualError(t, err, "not implemented")
}
