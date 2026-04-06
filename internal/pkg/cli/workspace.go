package cli

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/workspace"
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

	providers := map[string]workspace.Provider{
		"dir": dirProvider,
	}

	if gitProvider, err := git.NewProviderWithDefaults(runsPath); err == nil {
		providers["git+https"] = gitProvider
		providers["git+ssh"] = gitProvider
	}

	return workspace.NewManager(providers), nil
}
