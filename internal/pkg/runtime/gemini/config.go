package gemini

import (
	"encoding/json"
	"fmt"

	"github.com/mcuadros/go-defaults"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

// Config defines runtime options for Gemini execution.
type Config struct {
	// No specific configuration needed for Gemini at the moment
}

func applyRuntimeConfigArguments(config agent.RuntimeConfig) error {
	var geminiConfig Config

	switch typed := config.(type) {
	case nil:
		break
	case Config:
		geminiConfig = typed
	case *Config:
		if typed != nil {
			geminiConfig = *typed
		}
	default:
		payload, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("marshal gemini runtime config: %w", err)
		}

		if err := json.Unmarshal(payload, &geminiConfig); err != nil {
			return fmt.Errorf("unmarshal gemini runtime config: %w", err)
		}
	}

	defaults.SetDefaults(&geminiConfig)

	return nil
}

func applyRuntimeFeaturesArguments(_ *arguments, _ agent.RuntimeFeatures) {
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
