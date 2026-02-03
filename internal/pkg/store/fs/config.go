package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
)

const agentConfigFileExt = ".yaml"

// ConfigRepository is an implementation of agent.ConfigRepository that stores
// agent configurations on the file system.
type ConfigRepository struct {
	basePath string
	fs       afero.Fs
}

// NewConfigRepository creates a new file system-based config repository and
// ensures the base path exists.
func NewConfigRepository(basePath string, fs afero.Fs) (*ConfigRepository, error) {
	if err := fs.MkdirAll(basePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create agent config path: %w", err)
	}

	return &ConfigRepository{
		basePath: basePath,
		fs:       fs,
	}, nil
}

// Exists reports whether an agent config with the given identifier exists.
func (r *ConfigRepository) Exists(ctx context.Context, id briefkit.AgentID) (bool, error) {
	if err := id.Validate(); err != nil {
		return false, err
	}

	return afero.Exists(r.fs, r.configFilePath(id))
}

// Get loads the agent config for the given identifier.
func (r *ConfigRepository) Get(ctx context.Context, id briefkit.AgentID) (briefkit.Config, error) {
	if err := id.Validate(); err != nil {
		return briefkit.Config{}, err
	}

	config, err := readYAML[briefkit.Config](r.fs, r.configFilePath(id))
	if err != nil {
		return briefkit.Config{}, err
	}

	return config, nil
}

// Update persists the agent config for the given identifier.
func (r *ConfigRepository) Update(ctx context.Context, id briefkit.AgentID, config briefkit.Config) error {
	if err := id.Validate(); err != nil {
		return err
	}

	if err := writeYAML(r.fs, r.configFilePath(id), config); err != nil {
		return err
	}

	return nil
}

// List returns the identifiers of all available agent configs.
func (r *ConfigRepository) List(ctx context.Context) ([]briefkit.AgentID, error) {
	entries, err := afero.ReadDir(r.fs, r.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []briefkit.AgentID{}, nil
		}
		return nil, err
	}

	ids := make([]briefkit.AgentID, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != agentConfigFileExt {
			continue
		}

		id := briefkit.AgentID(strings.TrimSuffix(name, agentConfigFileExt))
		if err := id.Validate(); err != nil {
			continue
		}

		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	return ids, nil
}

func (r *ConfigRepository) configFilePath(id briefkit.AgentID) string {
	return filepath.Join(r.basePath, string(id)+agentConfigFileExt)
}
