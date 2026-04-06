package cwd

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errStatFs wraps afero.Fs and overrides Stat to always return the given error.
type errStatFs struct {
	afero.Fs
	err error
}

func (f *errStatFs) Stat(_ string) (os.FileInfo, error) {
	return nil, f.err
}

// errMkdirAllFs wraps afero.Fs and overrides MkdirAll to always return the given error.
type errMkdirAllFs struct {
	afero.Fs
	err error
}

func (f *errMkdirAllFs) MkdirAll(_ string, _ os.FileMode) error {
	return f.err
}

// errWriteFs wraps afero.Fs and makes Create/OpenFile fail for paths under the given prefix.
type errWriteFs struct {
	afero.Fs
	prefix string
	err    error
}

func (f *errWriteFs) Create(name string) (afero.File, error) {
	if strings.HasPrefix(name, f.prefix) {
		return nil, f.err
	}
	return f.Fs.Create(name)
}

func (f *errWriteFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if strings.HasPrefix(name, f.prefix) && flag != os.O_RDONLY {
		return nil, f.err
	}
	return f.Fs.OpenFile(name, flag, perm)
}

func TestProvider_Provision_WhenGetwdFails_ThenReturnsErrResolveCwd(t *testing.T) {
	getwdErr := errors.New("getwd failed")
	provider := NewProvider(afero.NewMemMapFs(), "/runs", func() (string, error) {
		return "", getwdErr
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.ErrorIs(t, err, ErrResolveCwd)
	require.ErrorIs(t, err, getwdErr)
}

func TestProvider_Provision_WhenCwdNotAbsolute_ThenReturnsErrResolveCwd(t *testing.T) {
	provider := NewProvider(afero.NewMemMapFs(), "/runs", func() (string, error) {
		return "relative/path", nil
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.ErrorIs(t, err, ErrResolveCwd)
}

func TestProvider_Provision_WhenCwdNotFound_ThenReturnsErrSourceNotFound(t *testing.T) {
	provider := NewProvider(afero.NewMemMapFs(), "/runs", func() (string, error) {
		return "/nonexistent", nil
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.ErrorIs(t, err, ErrSourceNotFound)
}

func TestProvider_Provision_WhenCwdNotDirectory_ThenReturnsErrSourceNotDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/src/file.txt", []byte("data"), 0644))
	provider := NewProvider(fs, "/runs", func() (string, error) {
		return "/src/file.txt", nil
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.ErrorIs(t, err, ErrSourceNotDirectory)
}

func TestProvider_Provision_WhenStatFails_ThenReturnsError(t *testing.T) {
	statErr := errors.New("permission denied")
	fs := &errStatFs{Fs: afero.NewMemMapFs(), err: statErr}
	provider := NewProvider(fs, "/runs", func() (string, error) {
		return "/src", nil
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.Error(t, err)
	require.ErrorIs(t, err, statErr)
}

func TestProvider_Provision_WhenMkdirFails_ThenReturnsError(t *testing.T) {
	memFs := afero.NewMemMapFs()
	require.NoError(t, memFs.MkdirAll("/src", 0755))
	mkdirErr := errors.New("mkdir failed")
	fs := &errMkdirAllFs{Fs: memFs, err: mkdirErr}
	provider := NewProvider(fs, "/runs", func() (string, error) {
		return "/src", nil
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.Error(t, err)
	require.ErrorIs(t, err, mkdirErr)
}

func TestProvider_Provision_WhenCopyFails_ThenCleansUpRunDir(t *testing.T) {
	memFs := afero.NewMemMapFs()
	require.NoError(t, memFs.MkdirAll("/src", 0755))
	require.NoError(t, afero.WriteFile(memFs, "/src/file.txt", []byte("data"), 0644))
	writeErr := errors.New("write failed")
	fs := &errWriteFs{Fs: memFs, prefix: "/runs/", err: writeErr}
	provider := NewProvider(fs, "/runs", func() (string, error) {
		return "/src", nil
	})

	_, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.Error(t, err)
	entries, readErr := afero.ReadDir(memFs, "/runs")
	require.NoError(t, readErr)
	assert.Empty(t, entries, "run directory should be cleaned up after copy failure")
}

func TestProvider_Provision_WhenSuccess_ThenCopiesFilesAndReturnsWorkDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("/src/subdir", 0755))
	require.NoError(t, afero.WriteFile(fs, "/src/file.txt", []byte("hello"), 0644))
	require.NoError(t, afero.WriteFile(fs, "/src/subdir/nested.txt", []byte("nested"), 0644))
	provider := NewProvider(fs, "/runs", func() (string, error) {
		return "/src", nil
	})

	result, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})

	require.NoError(t, err)
	require.NotEmpty(t, result.WorkDir)
	require.NotNil(t, result.Cleanup)
	content, readErr := afero.ReadFile(fs, result.WorkDir+"/file.txt")
	require.NoError(t, readErr)
	assert.Equal(t, []byte("hello"), content)
	nestedContent, readNestedErr := afero.ReadFile(fs, result.WorkDir+"/subdir/nested.txt")
	require.NoError(t, readNestedErr)
	assert.Equal(t, []byte("nested"), nestedContent)
}

func TestProvider_Provision_WhenSuccess_ThenCleanupRemovesWorkDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll("/src", 0755))
	require.NoError(t, afero.WriteFile(fs, "/src/file.txt", []byte("hello"), 0644))
	provider := NewProvider(fs, "/runs", func() (string, error) {
		return "/src", nil
	})

	result, err := provider.Provision(context.Background(), url.URL{Scheme: "cwd"})
	require.NoError(t, err)
	workDir := result.WorkDir

	cleanupErr := result.Cleanup()

	require.NoError(t, cleanupErr)
	_, statErr := fs.Stat(workDir)
	assert.True(t, os.IsNotExist(statErr), "work directory should not exist after cleanup")
}
