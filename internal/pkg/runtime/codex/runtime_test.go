package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/neongreen/mono/lib/toml"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type writeFailFs struct {
	afero.Fs
}

type readFailFs struct {
	afero.Fs
	path string
}

type statFailFs struct {
	afero.Fs
	path string
	err  error
}

func (fs statFailFs) Stat(name string) (os.FileInfo, error) {
	if name == fs.path {
		return nil, fs.err
	}
	return fs.Fs.Stat(name)
}

func TestRuntime_Discovery_WhenExecutableAvailable_ThenReturnsTrue(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)

	found, err := NewRuntime().Discovery(context.Background())

	require.NoError(t, err)
	assert.True(t, found)
}

func TestRuntime_Discovery_WhenExecutableMissing_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	t.Setenv(envExecutablePath, filepath.Join(t.TempDir(), "missing-codex"))

	found, err := NewRuntime().Discovery(context.Background())

	require.Error(t, err)
	assert.False(t, found)
	assert.Contains(t, err.Error(), "executable from "+envExecutablePath+" not found")
}

func TestRuntime_Discovery_WhenExecutableNotFound_ThenReturnsFalse(t *testing.T) {
	resetCodexMockEnv(t)
	t.Setenv(envExecutablePath, "")
	t.Setenv("PATH", "")

	found, err := NewRuntime().Discovery(context.Background())

	require.NoError(t, err)
	assert.False(t, found)
}

func TestRuntime_Discovery_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRuntime().Discovery(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_GetInfo_WhenVersionAvailable_ThenReturnsVersion(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)

	info, err := NewRuntime().GetInfo(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "1.0.0", info.Version)
}

func TestRuntime_GetInfo_WhenVersionCommandFails_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("MOCK_CODEX_VERSION_FAIL", "1")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "read codex version")
}

func TestRuntime_GetInfo_WhenOutputHasNoSemver_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("MOCK_CODEX_VERSION_NO_SEMVER", "1")

	_, err := NewRuntime().GetInfo(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse codex version from output")
}

func TestRuntime_GetInfo_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRuntime().GetInfo(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_RegisterMCPServer_WhenNameMissing_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	err := NewRuntime().RegisterMCPServer(context.Background(), "", validMCPServer())

	require.Error(t, err)
	assert.Equal(t, "missing mcp server name", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenSTDIOConfigMissing_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	err := NewRuntime().RegisterMCPServer(
		context.Background(),
		briefkit.RuntimeMCPServerName("briefkit"),
		briefkit.RuntimeMCPServer{},
	)

	require.Error(t, err)
	assert.Equal(t, "missing mcp server stdio configuration", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenCommandMissing_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	err := NewRuntime().RegisterMCPServer(
		context.Background(),
		briefkit.RuntimeMCPServerName("briefkit"),
		briefkit.RuntimeMCPServer{
			STDIO: &briefkit.RuntimeSTDIOMCPServer{Command: "   "},
		},
	)

	require.Error(t, err)
	assert.Equal(t, "missing mcp server stdio command", err.Error())
}

func TestRuntime_RegisterMCPServer_WhenRemoveNotFound_ThenReturnsNil(t *testing.T) {
	resetCodexMockEnv(t)
	home := setTempHome(t)
	setCodexMockExecutable(t)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".codex"), 0o750))
	t.Setenv("MOCK_CODEX_MCP_NOT_FOUND", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), briefkit.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.NoError(t, err)
}

func TestRuntime_RegisterMCPServer_WhenRemoveFails_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("MOCK_CODEX_MCP_REMOVE_FAIL", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), briefkit.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex mcp server removal")
}

func TestRuntime_RegisterMCPServer_WhenAddFailsWithOutput_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("MOCK_CODEX_MCP_ADD_FAIL", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), briefkit.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex mcp server registration")
	assert.Contains(t, err.Error(), "already exists")
}

func TestRuntime_RegisterMCPServer_WhenAddFailsWithoutOutput_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("MOCK_CODEX_MCP_ADD_FAIL", "1")
	t.Setenv("MOCK_CODEX_MCP_ADD_NO_OUTPUT", "1")

	err := NewRuntime().RegisterMCPServer(context.Background(), briefkit.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex mcp server registration")
	assert.Contains(t, err.Error(), "exit status")
}

func TestRuntime_RegisterMCPServer_WhenAddSucceeds_ThenWritesTimeoutConfig(t *testing.T) {
	resetCodexMockEnv(t)
	home := setTempHome(t)
	setCodexMockExecutable(t)
	require.NoError(t, os.MkdirAll(filepath.Join(home, ".codex"), 0o750))

	err := NewRuntime().RegisterMCPServer(context.Background(), briefkit.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.NoError(t, err)
	assertConfigTimeout(t, afero.NewOsFs(), filepath.Join(home, ".codex", "config.toml"), `mcp_servers."briefkit".tool_timeout_sec`)
}

func TestRuntime_RegisterMCPServer_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := NewRuntime().RegisterMCPServer(ctx, briefkit.RuntimeMCPServerName("briefkit"), validMCPServer())

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestRuntime_GetDefaultConfig_WhenCalled_ThenReturnsExpectedDefaults(t *testing.T) {
	config, err := NewRuntime().GetDefaultConfig(context.Background())

	require.NoError(t, err)
	codexConfig, ok := config.(RuntimeConfig)
	require.True(t, ok)
	assert.False(t, codexConfig.RequireWorkspaceRepository)
}

func TestRuntime_GetDefaultFeatures_WhenCalled_ThenReturnsEmptyFeatures(t *testing.T) {
	features, err := NewRuntime().GetDefaultFeatures(context.Background())

	require.NoError(t, err)
	assert.Equal(t, briefkit.RuntimeFeatures{}, features)
}

func TestRuntime_Execute_WhenValidConfig_ThenReturnsInstance(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", t.TempDir())

	config := briefkit.Config{}
	config.Runtime.Config = RuntimeConfig{RequireWorkspaceRepository: false}

	instance, err := NewRuntime().Execute(
		context.Background(),
		briefkit.ExecutionID("test-exec"),
		briefkit.ExecutionInput{Prompt: "hello"},
		config,
	)

	require.NoError(t, err)
	require.NotNil(t, instance)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := instance.Wait(ctx)

	require.NoError(t, err)
	assert.Contains(t, result.Response, "hello")
}

func TestRuntime_Execute_WhenConfigAsMap_ThenReturnsInstance(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", t.TempDir())

	config := briefkit.Config{}
	config.Runtime.Config = map[string]any{"requireWorkspaceRepository": false}

	instance, err := NewRuntime().Execute(
		context.Background(),
		briefkit.ExecutionID("test-exec"),
		briefkit.ExecutionInput{Prompt: "hello"},
		config,
	)

	require.NoError(t, err)
	require.NotNil(t, instance)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = instance.Wait(ctx)

	require.NoError(t, err)
}

func TestRuntime_Execute_WhenConfigInvalid_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", t.TempDir())

	config := briefkit.Config{}
	config.Runtime.Config = "invalid-not-a-struct"

	_, err := NewRuntime().Execute(
		context.Background(),
		briefkit.ExecutionID("test-exec"),
		briefkit.ExecutionInput{Prompt: "hello"},
		config,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "convert runtime config")
}

func TestRuntime_Execute_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	resetCodexMockEnv(t)
	setCodexMockExecutable(t)
	t.Setenv("BRIEFKIT_RUNTIME_LOG_DIR", t.TempDir())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	config := briefkit.Config{}
	config.Runtime.Config = RuntimeConfig{RequireWorkspaceRepository: false}

	_, err := NewRuntime().Execute(
		ctx,
		briefkit.ExecutionID("test-exec"),
		briefkit.ExecutionInput{Prompt: "hello"},
		config,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func (fs readFailFs) Open(name string) (afero.File, error) {
	if name == fs.path {
		return nil, errors.New("read failure")
	}

	return fs.Fs.Open(name)
}

func (fs writeFailFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, errors.New("write failure")
	}

	return fs.Fs.OpenFile(name, flag, perm)
}

func TestSetMcpServerTimeout_WhenConfigMissing_ThenCreatesFileWithTimeout(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")

	require.NoError(t, fs.MkdirAll(filepath.Dir(configPath), 0o755))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.NoError(t, err)
	assertConfigTimeout(t, fs, configPath, `mcp_servers."briefkit".tool_timeout_sec`)
}

func TestSetMcpServerTimeout_WhenConfigExists_ThenPreservesOtherKeys(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")
	existing := "foo = \"bar\"\n\n[mcp_servers.briefkit]\ncommand = \"bin/briefkit-mcp\"\n"

	require.NoError(t, fs.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, afero.WriteFile(fs, configPath, []byte(existing), 0o644))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.NoError(t, err)
	doc := readTomlDoc(t, fs, configPath)
	value, err := doc.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", value)
	commandValue, err := doc.Get("mcp_servers.briefkit.command")
	require.NoError(t, err)
	assert.Equal(t, "bin/briefkit-mcp", commandValue)
}

func TestSetMcpServerTimeout_WhenTimeoutExists_ThenOverwritesValue(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")
	existing := "[mcp_servers.briefkit]\ntool_timeout_sec = 5\n"

	require.NoError(t, fs.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, afero.WriteFile(fs, configPath, []byte(existing), 0o644))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.NoError(t, err)
	assertConfigTimeout(t, fs, configPath, `mcp_servers."briefkit".tool_timeout_sec`)
}

func TestSetMcpServerTimeout_WhenConfigExists_ThenPreservesFileMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")
	existing := "[mcp_servers]\n"

	require.NoError(t, fs.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, afero.WriteFile(fs, configPath, []byte(existing), 0o640))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.NoError(t, err)
	info, err := fs.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
}

func TestSetMcpServerTimeout_WhenContextCanceled_ThenReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runtime.setMcpServerTimeout(ctx, fs, "briefkit", configPath)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestSetMcpServerTimeout_WhenReadFails_ThenReturnsError(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")
	fs := readFailFs{Fs: baseFs, path: configPath}

	require.NoError(t, baseFs.MkdirAll(filepath.Dir(configPath), 0o755))
	require.NoError(t, afero.WriteFile(baseFs, configPath, []byte("foo = \"bar\""), 0o644))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex config read")
}

func TestSetMcpServerTimeout_WhenTomlInvalid_ThenReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")

	require.NoError(t, afero.WriteFile(fs, configPath, []byte("invalid = ["), 0o644))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex config parse")
}

func TestSetMcpServerTimeout_WhenWriteFails_ThenReturnsError(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	rs := &Runtime{}
	configPath := filepath.Join("config", "config.toml")
	fs := writeFailFs{Fs: baseFs}

	require.NoError(t, baseFs.MkdirAll(filepath.Dir(configPath), 0o755))

	err := rs.setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex config write")
}

func TestSetMcpServerTimeout_WhenStatFails_ThenReturnsError(t *testing.T) {
	baseFs := afero.NewMemMapFs()
	configPath := filepath.Join("config", "config.toml")
	fs := statFailFs{Fs: baseFs, path: configPath, err: errors.New("permission denied")}

	err := (&Runtime{}).setMcpServerTimeout(context.Background(), fs, "briefkit", configPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex config stat")
}

func TestSetMcpServerTimeout_WhenServerNameHasDot_ThenUsesQuotedKey(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")

	require.NoError(t, fs.MkdirAll(filepath.Dir(configPath), 0o755))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "brief.kit", configPath)

	require.NoError(t, err)
	assertConfigTimeout(t, fs, configPath, `mcp_servers."brief.kit".tool_timeout_sec`)
}

func assertConfigTimeout(t *testing.T, fs afero.Fs, configPath string, key string) {
	t.Helper()

	expected := int64(codexDefaultToolTimeout.Seconds())
	doc := readTomlDoc(t, fs, configPath)
	value, err := doc.Get(key)
	require.NoError(t, err)
	if value == nil {
		require.Fail(t, "expected value at "+key)
	}
	assert.Equal(t, expected, value)
}

func readTomlDoc(t *testing.T, fs afero.Fs, configPath string) *toml.Document {
	t.Helper()

	contents, err := afero.ReadFile(fs, configPath)
	require.NoError(t, err)
	doc, err := toml.Parse(contents)
	require.NoError(t, err)
	return doc
}

func validMCPServer() briefkit.RuntimeMCPServer {
	return briefkit.RuntimeMCPServer{
		STDIO: &briefkit.RuntimeSTDIOMCPServer{
			Command:   "echo",
			Arguments: []string{"hello"},
		},
	}
}

func setCodexMockExecutable(t *testing.T) {
	t.Helper()

	path, err := os.Executable()
	require.NoError(t, err)
	require.FileExists(t, path)

	t.Setenv(envExecutablePath, path)
	t.Setenv(codexMockEnvKey, "1")
}

func setTempHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	homedir.Reset()
	return home
}

func resetCodexMockEnv(t *testing.T) {
	t.Helper()

	keys := []string{
		"MOCK_CODEX_FAIL",
		"MOCK_CODEX_EXIT_CODE",
		"MOCK_CODEX_STDERR",
		"MOCK_CODEX_MALFORMED_JSON",
		"MOCK_CODEX_NO_RESULT",
		"MOCK_CODEX_NO_THREAD_STARTED",
		"MOCK_CODEX_SIGNAL",
		"MOCK_CODEX_PARTIAL_FAIL",
		"MOCK_CODEX_EMPTY_LINES",
		"MOCK_CODEX_WHITESPACE_VARIATIONS",
		"MOCK_CODEX_UNKNOWN_EVENTS",
		"MOCK_CODEX_MULTI_ITEM",
		"MOCK_CODEX_EMPTY_TEXT",
		"MOCK_CODEX_OTHER_ITEM_TYPE",
		"MOCK_CODEX_EMPTY_STDIN",
		"MOCK_CODEX_MIXED_OUTPUT",
		"MOCK_CODEX_VERSION_FAIL",
		"MOCK_CODEX_VERSION_NO_SEMVER",
		"MOCK_CODEX_MCP_NOT_FOUND",
		"MOCK_CODEX_MCP_REMOVE_FAIL",
		"MOCK_CODEX_MCP_ADD_FAIL",
		"MOCK_CODEX_MCP_ADD_NO_OUTPUT",
		"MOCK_CODEX_INVALID_SESSION",
		codexMockEnvKey,
		envExecutablePath,
	}

	for _, key := range keys {
		t.Setenv(key, "")
	}
}
