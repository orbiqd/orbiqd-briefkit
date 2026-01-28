package claude

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_Discovery_WhenExecutableAvailable_ThenReturnsTrue(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)

	found, err := NewRuntime().Discovery(context.Background())

	require.NoError(t, err)
	assert.True(t, found)
}

func TestRuntime_Discovery_WhenExecutableMissing_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("CLAUDE_EXECUTABLE", filepath.Join(t.TempDir(), "missing-claude"))

	found, err := NewRuntime().Discovery(context.Background())

	require.Error(t, err)
	assert.False(t, found)
	assert.Contains(t, err.Error(), "executable from CLAUDE_EXECUTABLE not found")
}

func TestRuntime_Discovery_WhenExecutableNotFound_ThenReturnsFalse(t *testing.T) {
	resetClaudeMockEnv(t)
	t.Setenv("CLAUDE_EXECUTABLE", "")
	t.Setenv("PATH", "")

	found, err := NewRuntime().Discovery(context.Background())

	require.NoError(t, err)
	assert.False(t, found)
}

func TestRuntime_Discovery_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRuntime().Discovery(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_GetInfo_WhenVersionAvailable_ThenReturnsVersion(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)

	info, err := NewRuntime().GetInfo(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", info.Version)
}

func TestRuntime_GetInfo_WhenVersionCommandFails_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)
	t.Setenv("MOCK_CLAUDE_VERSION_FAIL", "1")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read claude version")
}

func TestRuntime_GetInfo_WhenOutputHasNoSemver_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)
	t.Setenv("MOCK_CLAUDE_VERSION_NO_SEMVER", "1")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse claude version from output")
}

func TestRuntime_GetInfo_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRuntime().GetInfo(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_RegisterMCPServer_WhenNameMissing_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	err := NewRuntime().RegisterMCPServer(context.Background(), "", validMCPServer())

	require.Error(t, err)
	assert.Equal(t, "missing mcp server name", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenSTDIOConfigMissing_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	err := NewRuntime().RegisterMCPServer(
		context.Background(),
		agent.RuntimeMCPServerName("briefkit"),
		agent.RuntimeMCPServer{},
	)

	require.Error(t, err)
	assert.Equal(t, "missing mcp server stdio configuration", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenCommandMissing_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	err := NewRuntime().RegisterMCPServer(
		context.Background(),
		agent.RuntimeMCPServerName("briefkit"),
		agent.RuntimeMCPServer{
			STDIO: &agent.RuntimeSTDIOMCPServer{Command: "   "},
		},
	)

	require.Error(t, err)
	assert.Equal(t, "missing mcp server stdio command", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenRemoveNotFound_ThenReturnsNil(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)
	t.Setenv("MOCK_CLAUDE_MCP_NOT_FOUND", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.NoError(t, err)
}

func TestRuntime_RegisterMCPServer_WhenRemoveFails_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)
	t.Setenv("MOCK_CLAUDE_MCP_REMOVE_FAIL", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude mcp server removal")
}

func TestRuntime_RegisterMCPServer_WhenAddFailsWithOutput_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)
	t.Setenv("MOCK_CLAUDE_MCP_ADD_FAIL", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude mcp server registration")
	assert.Contains(t, err.Error(), "MCP server registration failed")
}

func TestRuntime_RegisterMCPServer_WhenAddFailsWithoutOutput_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)
	t.Setenv("MOCK_CLAUDE_MCP_ADD_FAIL", "1")
	t.Setenv("MOCK_CLAUDE_MCP_ADD_NO_OUTPUT", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude mcp server registration")
	assert.Contains(t, err.Error(), "exit status")
}

func TestRuntime_RegisterMCPServer_WhenAddSucceeds_ThenReturnsNil(t *testing.T) {
	resetClaudeMockEnv(t)
	setClaudeMockExecutable(t)

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.NoError(t, err)
}

func TestRuntime_RegisterMCPServer_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetClaudeMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewRuntime().RegisterMCPServer(ctx, agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func validMCPServer() agent.RuntimeMCPServer {
	return agent.RuntimeMCPServer{
		STDIO: &agent.RuntimeSTDIOMCPServer{
			Command:   "echo",
			Arguments: []string{"hello"},
		},
	}
}

func setClaudeMockExecutable(t *testing.T) {
	t.Helper()

	path, err := os.Executable()
	require.NoError(t, err)
	require.FileExists(t, path)

	t.Setenv("CLAUDE_EXECUTABLE", path)
	t.Setenv("BRIEFKIT_CLAUDE_MOCK", "1")
}

func resetClaudeMockEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"MOCK_CLAUDE_FAIL",
		"MOCK_CLAUDE_EXIT_CODE",
		"MOCK_CLAUDE_STDERR",
		"MOCK_CLAUDE_MALFORMED_JSON",
		"MOCK_CLAUDE_RESULT_ERROR",
		"MOCK_CLAUDE_NO_RESULT",
		"MOCK_CLAUDE_SIGNAL",
		"MOCK_CLAUDE_PARTIAL_FAIL",
		"MOCK_CLAUDE_MULTI_ASSISTANT",
		"MOCK_CLAUDE_EMPTY_LINES",
		"MOCK_CLAUDE_VERSION_FAIL",
		"MOCK_CLAUDE_VERSION_NO_SEMVER",
		"MOCK_CLAUDE_MCP_NOT_FOUND",
		"MOCK_CLAUDE_MCP_REMOVE_FAIL",
		"MOCK_CLAUDE_MCP_ADD_FAIL",
		"MOCK_CLAUDE_MCP_ADD_NO_OUTPUT",
		"BRIEFKIT_CLAUDE_MOCK",
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}
}
