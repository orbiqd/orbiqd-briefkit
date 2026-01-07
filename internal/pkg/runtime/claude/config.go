package claude

import (
	"encoding/json"
	"fmt"

	"github.com/mcuadros/go-defaults"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

type Config struct {
}

func applyRuntimeConfigArguments(args *arguments, config agent.RuntimeConfig) error {
	var claudeConfig Config

	switch typed := config.(type) {
	case nil:
		break
	case Config:
		claudeConfig = typed
	case *Config:
		if typed != nil {
			claudeConfig = *typed
		}
	default:
		payload, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("marshal claude runtime config: %w", err)
		}

		if err := json.Unmarshal(payload, &claudeConfig); err != nil {
			return fmt.Errorf("unmarshal claude runtime config: %w", err)
		}
	}

	defaults.SetDefaults(&claudeConfig)

	return nil
}

func applyRuntimeFeaturesArguments(args *arguments, features agent.RuntimeFeatures) error {
	if features.EnableWebSearch != nil && !*features.EnableWebSearch {
		err := args.SetValue("disallowed-tools", "WebSearch")
		if err != nil {
			return fmt.Errorf("disable web search: %w", err)
		}
	}

	return nil
}

func applyExecutionInputArguments(args *arguments, executionInput agent.ExecutionInput) error {
	var err error

	if executionInput.Model != nil {
		err = args.SetValue("model", *executionInput.Model)
		if err != nil {
			return fmt.Errorf("set model: %w", err)
		}
	}

	if executionInput.ConversationID != nil {
		err = args.SetValue("resume", *executionInput.ConversationID)
		if err != nil {
			return fmt.Errorf("set resume: %w", err)
		}
	}

	return nil
}
