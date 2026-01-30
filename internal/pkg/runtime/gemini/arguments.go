package gemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mcuadros/go-defaults"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

// RuntimeConfig defines runtime options for Gemini execution.
type RuntimeConfig struct {
	// No specific configuration needed for Gemini at the moment
}

// GeminiArguments constructs command-line arguments for the Gemini CLI runtime.
type GeminiArguments struct {
	Sandbox      *bool
	Model        *string
	Resume       *string
	OutputFormat *string
}

// NewGeminiArguments creates a new GeminiArguments instance with sensible defaults.
// OutputFormat is set to "stream-json" by default for parseable responses.
func NewGeminiArguments() *GeminiArguments {
	return &GeminiArguments{
		OutputFormat: utils.ToPointer("stream-json"),
	}
}

// ToSlice converts GeminiArguments into a command-line argument slice.
// Returns arguments in deterministic order.
func (arguments *GeminiArguments) ToSlice() []string {
	var list []string

	if arguments.Sandbox != nil {
		if *arguments.Sandbox {
			list = append(list, "--sandbox=true")
		} else {
			list = append(list, "--sandbox=false")
		}
	}

	if arguments.Model != nil {
		list = append(list, "--model="+*arguments.Model)
	}

	if arguments.Resume != nil {
		list = append(list, "--resume="+*arguments.Resume)
	}

	if arguments.OutputFormat != nil {
		list = append(list, "--output-format="+*arguments.OutputFormat)
	}

	return list
}

// ApplyRuntimeConfig applies gemini-specific runtime configuration to arguments.
// Handles type conversion from the opaque RuntimeConfig interface.
func (arguments *GeminiArguments) ApplyRuntimeConfig(config agent.RuntimeConfig) error {
	var geminiConfig RuntimeConfig

	switch typed := config.(type) {
	case nil:
		break
	case RuntimeConfig:
		geminiConfig = typed
	case *RuntimeConfig:
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

// ApplyRuntimeFeatures applies runtime feature flags to gemini arguments.
func (arguments *GeminiArguments) ApplyRuntimeFeatures(features agent.RuntimeFeatures) error {
	if features.EnableSandbox == nil {
		return nil
	}

	arguments.Sandbox = features.EnableSandbox
	return nil
}

// ApplyExecutionInput applies execution-specific inputs like model selection.
// Returns error if model or conversationID value is provided but empty.
func (arguments *GeminiArguments) ApplyExecutionInput(executionInput agent.ExecutionInput) error {
	if executionInput.Model != nil {
		modelValue := strings.TrimSpace(*executionInput.Model)
		if modelValue == "" {
			return errors.New("model cannot be empty")
		}
		arguments.Model = &modelValue
	}

	if executionInput.ConversationID != nil {
		conversationID := strings.TrimSpace(string(*executionInput.ConversationID))
		if conversationID == "" {
			return errors.New("conversationID cannot be empty")
		}
		arguments.Resume = &conversationID
	}

	return nil
}
