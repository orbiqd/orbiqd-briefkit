package codex

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/lib/toml"
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
	assertConfigTimeout(t, fs, configPath, `mcp_servers."briefkit".tool_timeout_sec`, int64(600))
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
	assertConfigTimeout(t, fs, configPath, `mcp_servers."briefkit".tool_timeout_sec`, int64(600))
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

func TestSetMcpServerTimeout_WhenServerNameHasDot_ThenUsesQuotedKey(t *testing.T) {
	fs := afero.NewMemMapFs()
	runtime := &Runtime{}
	configPath := filepath.Join("config", "config.toml")

	require.NoError(t, fs.MkdirAll(filepath.Dir(configPath), 0o755))

	err := runtime.setMcpServerTimeout(context.Background(), fs, "brief.kit", configPath)

	require.NoError(t, err)
	assertConfigTimeout(t, fs, configPath, `mcp_servers."brief.kit".tool_timeout_sec`, int64(600))
}

func assertConfigTimeout(t *testing.T, fs afero.Fs, configPath string, key string, expected int64) {
	t.Helper()

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
