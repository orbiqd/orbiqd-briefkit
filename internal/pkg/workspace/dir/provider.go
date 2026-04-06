package dir

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

// Provider provisions isolated workspace copies from a local directory source.
// The scheme for this provider is "dir" (e.g. dir:///path/to/project).
type Provider struct {
	fs       afero.Fs
	runsPath string
}

// NewProvider creates a Provider that stores per-execution copies under runsPath.
func NewProvider(fs afero.Fs, runsPath string) *Provider {
	return &Provider{fs: fs, runsPath: runsPath}
}

// Provision copies the directory identified by uri into a fresh per-execution directory.
// Returns ErrSourceNotFound when the source path does not exist.
// Returns ErrSourceNotDirectory when the source path is not a directory.
func (p *Provider) Provision(ctx context.Context, uri url.URL) (workspace.ProvisionResult, error) {
	srcPath := uri.Path
	if !filepath.IsAbs(srcPath) {
		return workspace.ProvisionResult{}, ErrSourceNotAbsolute
	}

	info, err := p.fs.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return workspace.ProvisionResult{}, ErrSourceNotFound
		}
		return workspace.ProvisionResult{}, fmt.Errorf("stat source: %w", err)
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
	// ErrSourceNotFound indicates the source directory does not exist.
	ErrSourceNotFound = errors.New("workspace source not found")

	// ErrSourceNotDirectory indicates the source path is not a directory.
	ErrSourceNotDirectory = errors.New("workspace source is not a directory")

	// ErrSourceNotAbsolute indicates the source path extracted from the URI is not absolute.
	ErrSourceNotAbsolute = errors.New("workspace source path must be absolute")
)
