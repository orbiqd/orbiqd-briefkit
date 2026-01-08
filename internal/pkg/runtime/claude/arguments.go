package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mcuadros/go-defaults"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

type ClaudeArguments struct {
	Print           *bool
	Verbose         *bool
	OutputFormat    *string
	Model           *string
	ResumeSessionID *string
	DisallowedTools []string
	Settings        map[string]any
}

func NewClaudeArguments() *ClaudeArguments {
	return &ClaudeArguments{
		Print:        utils.ToPointer(true),
		Verbose:      utils.ToPointer(true),
		OutputFormat: utils.ToPointer("stream-json"),
		Settings:     make(map[string]any),
	}
}

func (arguments *ClaudeArguments) ToSlice() []string {
	var list []string

	if arguments.Print != nil && *arguments.Print {
		list = append(list, "--print")
	}

	if arguments.Verbose != nil && *arguments.Verbose {
		list = append(list, "--verbose")
	}

	if arguments.OutputFormat != nil {
		list = append(list, fmt.Sprintf("--output-format=%s", *arguments.OutputFormat))
	}

	if arguments.Model != nil {
		list = append(list, fmt.Sprintf("--model=%s", *arguments.Model))
	}

	if arguments.ResumeSessionID != nil {
		list = append(list, fmt.Sprintf("--resume=%s", *arguments.ResumeSessionID))
	}

	if len(arguments.DisallowedTools) > 0 {
		list = append(list, fmt.Sprintf("--disallowed-tools=%s", strings.Join(arguments.DisallowedTools, ",")))
	}

	if len(arguments.Settings) > 0 {
		settingsJSON, err := json.Marshal(arguments.Settings)
		if err == nil {
			list = append(list, fmt.Sprintf("--settings=%s", string(settingsJSON)))
		}
	}

	return list
}

func (arguments *ClaudeArguments) ApplyRuntimeConfig(config agent.RuntimeConfig) error {
	var claudeConfig RuntimeConfig

	switch typed := config.(type) {
	case nil:
		break
	case RuntimeConfig:
		claudeConfig = typed
	case *RuntimeConfig:
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

func (arguments *ClaudeArguments) ApplyRuntimeFeatures(features agent.RuntimeFeatures) error {
	if features.EnableWebSearch != nil && !*features.EnableWebSearch {
		arguments.DisallowedTools = append(arguments.DisallowedTools, "WebSearch")
	}

	return nil
}

func (arguments *ClaudeArguments) ApplyExecutionInput(executionInput agent.ExecutionInput) error {
	if executionInput.Model != nil {
		arguments.Model = executionInput.Model
	}

	if executionInput.ConversationID != nil {
		sessionID := string(*executionInput.ConversationID)
		arguments.ResumeSessionID = &sessionID
	}

	return nil
}
