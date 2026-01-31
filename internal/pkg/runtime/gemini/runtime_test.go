package gemini

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
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)

	found, err := NewRuntime().Discovery(context.Background())

	require.NoError(t, err)
	assert.True(t, found)
}

func TestRuntime_Discovery_WhenExecutableNotFound_ThenReturnsFalse(t *testing.T) {
	resetGeminiMockEnv(t)
	t.Setenv("PATH", "")

	found, err := NewRuntime().Discovery(context.Background())

	require.NoError(t, err)
	assert.False(t, found)
}

func TestRuntime_Discovery_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRuntime().Discovery(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_GetInfo_WhenVersionAvailable_ThenReturnsVersion(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)

	info, err := NewRuntime().GetInfo(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", info.Version)
}

func TestRuntime_GetInfo_WhenVersionCommandFails_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_VERSION_FAIL", "1")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read gemini version")
}

func TestRuntime_GetInfo_WhenOutputHasNoSemver_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_VERSION_NO_SEMVER", "1")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse gemini version from output")
}

func TestRuntime_GetInfo_WhenExecutableLookupFails_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	t.Setenv("PATH", "")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "lookup gemini executable")
}

func TestRuntime_GetInfo_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRuntime().GetInfo(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_GetDefaultConfig_ReturnsEmptyConfig(t *testing.T) {
	config, err := NewRuntime().GetDefaultConfig(context.Background())

	require.NoError(t, err)
	assert.Equal(t, RuntimeConfig{}, config)
}

func TestRuntime_GetDefaultFeatures_ReturnsEmptyFeatures(t *testing.T) {
	features, err := NewRuntime().GetDefaultFeatures(context.Background())

	require.NoError(t, err)
	assert.Equal(t, agent.RuntimeFeatures{}, features)
}

func TestRuntime_RegisterMCPServer_WhenNameMissing_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	err := NewRuntime().RegisterMCPServer(context.Background(), "", validMCPServer())

	require.Error(t, err)
	assert.Equal(t, "missing mcp server name", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenSTDIOConfigMissing_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	err := NewRuntime().RegisterMCPServer(
		context.Background(),
		agent.RuntimeMCPServerName("briefkit"),
		agent.RuntimeMCPServer{},
	)

	require.Error(t, err)
	assert.Equal(t, "missing mcp server stdio configuration", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenCommandMissing_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	err := NewRuntime().RegisterMCPServer(
		context.Background(),
		agent.RuntimeMCPServerName("briefkit"),
		agent.RuntimeMCPServer{STDIO: &agent.RuntimeSTDIOMCPServer{Command: "   "}},
	)

	require.Error(t, err)
	assert.Equal(t, "missing mcp server stdio command", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenRemoveNotFound_ThenReturnsNil(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_MCP_NOT_FOUND", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.NoError(t, err)
}

func TestRuntime_RegisterMCPServer_WhenRemoveFails_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_MCP_REMOVE_FAIL", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gemini mcp server removal")
}

func TestRuntime_RegisterMCPServer_WhenAddFailsWithOutput_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_MCP_ADD_FAIL", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gemini mcp server registration")
	assert.Contains(t, err.Error(), "registration failed")
}

func TestRuntime_RegisterMCPServer_WhenAddFailsWithoutOutput_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)
	t.Setenv("MOCK_GEMINI_MCP_ADD_FAIL", "1")
	t.Setenv("MOCK_GEMINI_MCP_ADD_NO_OUTPUT", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "gemini mcp server registration")
	assert.Contains(t, err.Error(), "exit status")
}

func TestRuntime_RegisterMCPServer_WhenAddSucceeds_ThenReturnsNil(t *testing.T) {
	resetGeminiMockEnv(t)
	setGeminiMockExecutable(t)

	err := NewRuntime().RegisterMCPServer(context.Background(), agent.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.NoError(t, err)
}

func TestRuntime_RegisterMCPServer_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetGeminiMockEnv(t)
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

func setGeminiMockExecutable(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	target, err := os.Executable()
	require.NoError(t, err)
	require.FileExists(t, target)

	executablePath := filepath.Join(tempDir, "gemini")
	require.NoError(t, os.Link(target, executablePath))

	t.Setenv("PATH", tempDir)
	t.Setenv(geminiMockEnvKey, "1")
}

func resetGeminiMockEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"MOCK_GEMINI_FAIL",
		"MOCK_GEMINI_EXIT_CODE",
		"MOCK_GEMINI_STDERR",
		"MOCK_GEMINI_MALFORMED_JSON",
		"MOCK_GEMINI_NO_INIT",
		"MOCK_GEMINI_NO_MESSAGE",
		"MOCK_GEMINI_SIGNAL",
		"MOCK_GEMINI_PARTIAL_FAIL",
		"MOCK_GEMINI_MULTI_MESSAGE",
		"MOCK_GEMINI_EMPTY_LINES",
		"MOCK_GEMINI_UNKNOWN_EVENTS",
		"MOCK_GEMINI_ERROR_RESULT",
		"MOCK_GEMINI_VERSION_FAIL",
		"MOCK_GEMINI_VERSION_NO_SEMVER",
		"MOCK_GEMINI_MCP_NOT_FOUND",
		"MOCK_GEMINI_MCP_REMOVE_FAIL",
		"MOCK_GEMINI_MCP_ADD_FAIL",
		"MOCK_GEMINI_MCP_ADD_NO_OUTPUT",
		"MOCK_GEMINI_INVALID_SESSION",
		geminiMockEnvKey,
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}
}
