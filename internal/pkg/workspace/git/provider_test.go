package git

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParseURI(raw string) url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return *u
}

type runCall struct {
	name string
	args []string
}

type runResult struct {
	output []byte
	err    error
}

type stubRunner struct {
	calls   []runCall
	results []runResult
}

func (s *stubRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	idx := len(s.calls)
	s.calls = append(s.calls, runCall{name: name, args: args})
	if idx < len(s.results) {
		return s.results[idx].output, s.results[idx].err
	}
	return nil, nil
}

func TestProvider_Provision_WhenSchemeIsGitPlusHTTPS_ThenCloneURLUsesHTTPS(t *testing.T) {
	stub := &stubRunner{}
	p := NewProvider(stub, t.TempDir())

	_, _ = p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git"))

	require.Len(t, stub.calls, 1)
	cloneURL := stub.calls[0].args[1]
	assert.True(t, strings.HasPrefix(cloneURL, "https://"), "expected https:// clone URL, got: %s", cloneURL)
}

func TestProvider_Provision_WhenSchemeIsGitPlusSSH_ThenCloneURLUsesSSH(t *testing.T) {
	stub := &stubRunner{}
	p := NewProvider(stub, t.TempDir())

	_, _ = p.Provision(context.Background(), mustParseURI("git+ssh://git@github.com/org/repo.git"))

	require.Len(t, stub.calls, 1)
	cloneURL := stub.calls[0].args[1]
	assert.True(t, strings.HasPrefix(cloneURL, "ssh://git@"), "expected ssh://git@ clone URL, got: %s", cloneURL)
}

func TestProvider_Provision_WhenCloneURLHasQueryParams_ThenStrippedFromCloneURL(t *testing.T) {
	stub := &stubRunner{}
	p := NewProvider(stub, t.TempDir())

	_, _ = p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git?ref=main"))

	require.NotEmpty(t, stub.calls)
	cloneURL := stub.calls[0].args[1]
	assert.NotContains(t, cloneURL, "?", "query params should be stripped from clone URL")
	assert.NotContains(t, cloneURL, "ref=", "ref param should not appear in clone URL")
}

func TestProvider_Provision_WhenNoRef_ThenClonesWithoutCheckout(t *testing.T) {
	stub := &stubRunner{}
	p := NewProvider(stub, t.TempDir())

	result, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git"))

	require.NoError(t, err)
	assert.Len(t, stub.calls, 1)
	assert.Equal(t, "clone", stub.calls[0].args[0])
	assert.NotEmpty(t, result.WorkDir)
}

func TestProvider_Provision_WhenRefProvided_ThenClonesAndChecksOut(t *testing.T) {
	stub := &stubRunner{}
	p := NewProvider(stub, t.TempDir())

	result, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git?ref=v1.2.3"))

	require.NoError(t, err)
	require.Len(t, stub.calls, 2)
	assert.Equal(t, "clone", stub.calls[0].args[0])
	assert.Equal(t, "-C", stub.calls[1].args[0])
	assert.Equal(t, "checkout", stub.calls[1].args[2])
	assert.Equal(t, "v1.2.3", stub.calls[1].args[3])
	assert.NotEmpty(t, result.WorkDir)
}

func TestProvider_Provision_WhenCloneFails_ThenReturnsErrCloneFailed(t *testing.T) {
	stub := &stubRunner{
		results: []runResult{{output: []byte("fatal: repo not found"), err: errors.New("exit 128")}},
	}
	p := NewProvider(stub, t.TempDir())

	_, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git"))

	require.Error(t, err)
	require.ErrorIs(t, err, ErrCloneFailed)
}

func TestProvider_Provision_WhenCheckoutFails_ThenReturnsErrCheckoutFailed(t *testing.T) {
	stub := &stubRunner{
		results: []runResult{
			{output: nil, err: nil},
			{output: []byte("error: pathspec 'bad-ref' did not match"), err: errors.New("exit 1")},
		},
	}
	p := NewProvider(stub, t.TempDir())

	_, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git?ref=bad-ref"))

	require.Error(t, err)
	require.ErrorIs(t, err, ErrCheckoutFailed)
}

func TestProvider_Provision_WhenSuccess_ThenReturnsWorkDirUnderRunsPath(t *testing.T) {
	runsPath := t.TempDir()
	stub := &stubRunner{}
	p := NewProvider(stub, runsPath)

	result, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git"))

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result.WorkDir, runsPath), "WorkDir should be under runsPath")
}

func TestProvider_Provision_WhenSuccess_ThenCleanupIsNotNil(t *testing.T) {
	stub := &stubRunner{}
	p := NewProvider(stub, t.TempDir())

	result, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git"))

	require.NoError(t, err)
	assert.NotNil(t, result.Cleanup)
}

func TestProvider_Provision_WhenRunsPathCreationFails_ThenReturnsError(t *testing.T) {
	// Create a file at the parent path so MkdirAll cannot create a directory there.
	parent := t.TempDir()
	blocker := parent + "/blocker"
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))

	stub := &stubRunner{}
	p := NewProvider(stub, blocker+"/runs")

	_, err := p.Provision(context.Background(), mustParseURI("git+https://github.com/org/repo.git"))

	require.Error(t, err)
}
