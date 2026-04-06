package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace/cwd"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace/dir"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace/git"
)

const workspaceRunsDirName = "workspaces/runs"

// CreateWorkspaceManagerFromConfig creates a WorkspaceManager rooted under the configured state path.
func CreateWorkspaceManagerFromConfig(config StoreConfig) (*workspace.Manager, error) {
	expanded, err := homedir.Expand(config.StatePath)
	if err != nil {
		return nil, fmt.Errorf("expand state path: %w", err)
	}

	cleaned := filepath.Clean(expanded)
	if !filepath.IsAbs(cleaned) {
		return nil, fmt.Errorf("state path must be absolute: %s", config.StatePath)
	}

	runsPath := filepath.Join(cleaned, workspaceRunsDirName)
	fs := afero.NewOsFs()

	dirProvider := dir.NewProvider(fs, runsPath)
	cwdProvider := cwd.NewProvider(fs, runsPath, os.Getwd)

	providers := map[string]workspace.Provider{
		"dir": dirProvider,
		"cwd": cwdProvider,
	}

	if gitProvider, err := git.NewProviderWithDefaults(runsPath); err == nil {
		providers["git+https"] = gitProvider
		providers["git+ssh"] = gitProvider
	}

	return workspace.NewManager(providers), nil
}
