package cli

import (
	"fmt"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	fsstore "github.com/orbiqd/orbiqd-briefkit/internal/pkg/store/fs"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
)

type StoreConfig struct {
	StatePath       string `short:"s" help:"Base directory for runtime state." default:"~/.orbiqd/briefkit/state" env:"BRIEFKIT_STATE_PATH"`
	AgentConfigPath string `help:"Directory with agent definition files." default:"~/.orbiqd/briefkit/agents" env:"BRIEFKIT_AGENT_CONFIG_PATH"`
}

const executionRepositoryDirName = "executions"

func CreateExecutionRepositoryFromConfig(config StoreConfig) (briefkit.ExecutionRepository, error) {
	expanded, err := homedir.Expand(config.StatePath)
	if err != nil {
		return nil, fmt.Errorf("expand state path: %w", err)
	}

	cleaned := filepath.Clean(expanded)
	if !filepath.IsAbs(cleaned) {
		return nil, fmt.Errorf("state path must be absolute: %s", config.StatePath)
	}

	repositoryPath := filepath.Join(cleaned, executionRepositoryDirName)
	fs := afero.NewOsFs()
	repository, err := fsstore.NewExecutionRepository(repositoryPath, fs)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func CreateConfigRepositoryFromConfig(config StoreConfig) (briefkit.ConfigRepository, error) {
	expanded, err := homedir.Expand(config.AgentConfigPath)
	if err != nil {
		return nil, fmt.Errorf("expand agent config path: %w", err)
	}

	cleaned := filepath.Clean(expanded)
	if !filepath.IsAbs(cleaned) {
		return nil, fmt.Errorf("agent config path must be absolute: %s", config.AgentConfigPath)
	}

	fs := afero.NewOsFs()
	repository, err := fsstore.NewConfigRepository(cleaned, fs)
	if err != nil {
		return nil, err
	}

	return repository, nil
}
