package codex

import (
	"encoding/json"
	"fmt"

	"github.com/mcuadros/go-defaults"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

// Config defines runtime options for Codex execution.
type Config struct {
	// RequireWorkspaceRepository enforces that codex workdir must be a GIT repository.
	RequireWorkspaceRepository bool `json:"requireWorkspaceRepository" default:"true"`
}

func applyRuntimeConfigArguments(runtimeArguments *arguments, config agent.RuntimeConfig) error {
	var codexConfig Config

	switch typed := config.(type) {
	case nil:
		break
	case Config:
		codexConfig = typed
	case *Config:
		if typed != nil {
			codexConfig = *typed
		}
	default:
		payload, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("marshal codex runtime config: %w", err)
		}

		if err := json.Unmarshal(payload, &codexConfig); err != nil {
			return fmt.Errorf("unmarshal codex runtime config: %w", err)
		}
	}

	defaults.SetDefaults(&codexConfig)

	if !codexConfig.RequireWorkspaceRepository {
		runtimeArguments.SetFlag("skip-git-repo-check")
	}

	return nil
}

func applyRuntimeFeaturesArguments(runtimeArguments *arguments, features agent.RuntimeFeatures) error {
	var err error

	if features.EnableNetworkAccess != nil {
		err = runtimeArguments.SetConfigOverride("sandbox_workspace_write.network_access", *features.EnableNetworkAccess)
		if err != nil {
			return fmt.Errorf("enable network access: %w", err)
		}
	}

	if features.EnableWebSearch != nil {
		err = runtimeArguments.SetConfigOverride("features.web_search_request", *features.EnableWebSearch)
		if err != nil {
			return fmt.Errorf("enable web search: %w", err)
		}
	}

	return nil
}

func applyExecutionInputArguments(runtimeArguments *arguments, executionInput agent.ExecutionInput) error {
	var err error

	if executionInput.Model != nil {
		err = runtimeArguments.SetValue("model", *executionInput.Model)
		if err != nil {
			return fmt.Errorf("set model: %w", err)
		}
	}

	return nil
}
