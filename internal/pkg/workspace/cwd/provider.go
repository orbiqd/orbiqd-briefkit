package cwd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	aferocopy "go.nhat.io/aferocopy/v2"

	"github.com/google/uuid"
	"github.com/spf13/afero"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace"
)

// Provider provisions isolated workspace copies from the current working directory.
// The scheme for this provider is "cwd" (e.g. cwd://).
type Provider struct {
	fs       afero.Fs
	runsPath string
	getwd    func() (string, error)
}

// NewProvider creates a Provider that stores per-execution copies under runsPath.
// getwd is called on each Provision to resolve the current working directory.
func NewProvider(fs afero.Fs, runsPath string, getwd func() (string, error)) *Provider {
	return &Provider{fs: fs, runsPath: runsPath, getwd: getwd}
}

// Provision copies the current working directory into a fresh per-execution directory.
// Returns ErrResolveCwd when the current working directory cannot be determined.
func (p *Provider) Provision(ctx context.Context, _ url.URL) (workspace.ProvisionResult, error) {
	srcPath, err := p.getwd()
	if err != nil {
		return workspace.ProvisionResult{}, fmt.Errorf("%w: %w", ErrResolveCwd, err)
	}

	if !filepath.IsAbs(srcPath) {
		return workspace.ProvisionResult{}, fmt.Errorf("%w: path must be absolute", ErrResolveCwd)
	}

	info, err := p.fs.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return workspace.ProvisionResult{}, ErrSourceNotFound
		}
		return workspace.ProvisionResult{}, fmt.Errorf("stat cwd: %w", err)
	}

	if !info.IsDir() {
		return workspace.ProvisionResult{}, ErrSourceNotDirectory
	}

	runDir := filepath.Join(p.runsPath, uuid.NewString())
	if err := p.fs.MkdirAll(runDir, 0o755); err != nil {
		return workspace.ProvisionResult{}, fmt.Errorf("create run directory: %w", err)
	}

	if err := aferocopy.Copy(srcPath, runDir, aferocopy.Options{
		SrcFs:  p.fs,
		DestFs: p.fs,
		OnSymlink: func(_ afero.Fs, _ string) aferocopy.SymlinkAction {
			return aferocopy.Shallow
		},
	}); err != nil {
		_ = p.fs.RemoveAll(runDir)
		return workspace.ProvisionResult{}, fmt.Errorf("copy workspace: %w", err)
	}

	cleanup := func() error {
		return p.fs.RemoveAll(runDir)
	}

	return workspace.ProvisionResult{WorkDir: runDir, Cleanup: cleanup}, nil
}

var (
	// ErrResolveCwd indicates the current working directory could not be determined.
	ErrResolveCwd = errors.New("workspace cwd resolution failed")

	// ErrSourceNotFound indicates the current working directory does not exist on the filesystem.
	ErrSourceNotFound = errors.New("workspace source not found")

	// ErrSourceNotDirectory indicates the current working directory path is not a directory.
	ErrSourceNotDirectory = errors.New("workspace source is not a directory")
)
