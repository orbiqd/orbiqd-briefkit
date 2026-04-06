package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cli/safeexec"
	"github.com/google/uuid"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace"
)

// commandRunner executes external commands and returns their combined output.
type commandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// Provider provisions isolated workspace copies by cloning a remote git repository.
// Supported schemes: git+https (e.g. git+https://github.com/org/repo.git?ref=main)
// and git+ssh (e.g. git+ssh://git@github.com/org/repo.git?ref=v1.2.3).
// The optional ref query param selects a branch, tag, or commit; omitting it
// uses the repository's default branch.
type Provider struct {
	runner   commandRunner
	runsPath string
}

// NewProvider creates a Provider with an injected commandRunner. Prefer this constructor
// in tests; use NewProviderWithDefaults for production code.
func NewProvider(runner commandRunner, runsPath string) *Provider {
	return &Provider{runner: runner, runsPath: runsPath}
}

// NewProviderWithDefaults creates a production Provider that shells out to the system
// git binary. Returns ErrGitNotFound when git is not available on the host.
func NewProviderWithDefaults(runsPath string) (*Provider, error) {
	gitPath, err := safeexec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGitNotFound, err)
	}
	return &Provider{
		runner:   &execRunner{gitPath: gitPath},
		runsPath: runsPath,
	}, nil
}

// Provision clones the repository identified by uri into a fresh per-execution directory.
// The URI scheme must be git+https or git+ssh; the git+ prefix is stripped when
// constructing the clone URL so that git receives a valid https:// or ssh:// URL.
func (p *Provider) Provision(ctx context.Context, uri url.URL) (workspace.ProvisionResult, error) {
	cloneURL := reconstructCloneURL(uri)
	ref := uri.Query().Get("ref")

	if err := os.MkdirAll(p.runsPath, 0o750); err != nil {
		return workspace.ProvisionResult{}, fmt.Errorf("create runs directory: %w", err)
	}

	runDir := filepath.Join(p.runsPath, uuid.NewString())

	if out, err := p.runner.Run(ctx, "git", "clone", cloneURL, runDir); err != nil {
		_ = os.RemoveAll(runDir)
		return workspace.ProvisionResult{}, fmt.Errorf("%w: %s", ErrCloneFailed, string(out))
	}

	if ref != "" {
		if out, err := p.runner.Run(ctx, "git", "-C", runDir, "checkout", ref); err != nil {
			_ = os.RemoveAll(runDir)
			return workspace.ProvisionResult{}, fmt.Errorf("%w: %s", ErrCheckoutFailed, string(out))
		}
	}

	cleanup := func() error {
		return os.RemoveAll(runDir)
	}

	return workspace.ProvisionResult{WorkDir: runDir, Cleanup: cleanup}, nil
}

// reconstructCloneURL builds the actual clone URL by stripping the "git+" prefix from
// the scheme and omitting query params (git does not understand them).
func reconstructCloneURL(uri url.URL) string {
	cloneScheme := strings.TrimPrefix(uri.Scheme, "git+")
	clone := url.URL{
		Scheme: cloneScheme,
		User:   uri.User,
		Host:   uri.Host,
		Path:   uri.Path,
	}
	return clone.String()
}

// execRunner is the production commandRunner that shells out to the system git binary.
type execRunner struct {
	gitPath string
}

func (r *execRunner) Run(ctx context.Context, _ string, args ...string) ([]byte, error) {
	// #nosec G204
	out, err := exec.CommandContext(ctx, r.gitPath, args...).CombinedOutput()
	return out, err
}

var (
	// ErrGitNotFound indicates the git executable is not available on the host.
	ErrGitNotFound = errors.New("git executable not found")

	// ErrCloneFailed indicates the git clone command failed.
	ErrCloneFailed = errors.New("git clone failed")

	// ErrCheckoutFailed indicates the git checkout command failed.
	ErrCheckoutFailed = errors.New("git checkout failed")
)
