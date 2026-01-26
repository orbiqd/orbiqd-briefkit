package codex

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mcuadros/go-defaults"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

// CodexArguments constructs command-line arguments for the Codex CLI runtime.
type CodexArguments struct {
	JSON             *bool
	SkipGitRepoCheck *bool
	Model            *string
	ConfigOverrides  map[string]any
}

// NewCodexArguments creates a new CodexArguments instance with sensible defaults.
// JSON output is enabled by default for parseable responses.
func NewCodexArguments() *CodexArguments {
	return &CodexArguments{
		JSON:            utils.ToPointer(true),
		ConfigOverrides: make(map[string]any),
	}
}

// ToSlice converts CodexArguments into a command-line argument slice.
// Returns arguments in deterministic order with sorted config overrides.
func (arguments *CodexArguments) ToSlice() []string {
	var list []string

	if arguments.JSON != nil && *arguments.JSON {
		list = append(list, "--json")
	}

	if arguments.SkipGitRepoCheck != nil && *arguments.SkipGitRepoCheck {
		list = append(list, "--skip-git-repo-check")
	}

	if arguments.Model != nil {
		list = append(list, "--model="+*arguments.Model)
	}

	// RuntimeConfig overrides - sorted for determinism
	if len(arguments.ConfigOverrides) > 0 {
		keys := make([]string, 0, len(arguments.ConfigOverrides))
		for key := range arguments.ConfigOverrides {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := arguments.ConfigOverrides[key]
			var valueStr string
			switch v := value.(type) {
			case bool:
				valueStr = strconv.FormatBool(v)
			case string:
				valueStr = v
			default:
				continue
			}
			list = append(list, fmt.Sprintf("--config=%s=%s", key, valueStr))
		}
	}

	return list
}

// ApplyRuntimeConfig applies codex-specific runtime configuration to arguments.
// Handles type conversion from the opaque RuntimeConfig interface.
func (arguments *CodexArguments) ApplyRuntimeConfig(config agent.RuntimeConfig) error {
	var codexConfig RuntimeConfig
	applyDefaults := false

	switch typed := config.(type) {
	case nil:
		applyDefaults = true
	case RuntimeConfig:
		codexConfig = typed
	case *RuntimeConfig:
		if typed != nil {
			codexConfig = *typed
		} else {
			applyDefaults = true
		}
	case map[string]any:
		if _, ok := typed["requireWorkspaceRepository"]; !ok {
			applyDefaults = true
		}
		payload, err := json.Marshal(typed)
		if err != nil {
			return fmt.Errorf("marshal codex runtime config: %w", err)
		}

		if err := json.Unmarshal(payload, &codexConfig); err != nil {
			return fmt.Errorf("unmarshal codex runtime config: %w", err)
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

	if applyDefaults {
		defaults.SetDefaults(&codexConfig)
	}

	if !codexConfig.RequireWorkspaceRepository {
		arguments.SkipGitRepoCheck = utils.ToPointer(true)
	}

	return nil
}

// ApplyRuntimeFeatures currently ignores runtime feature flags.
func (arguments *CodexArguments) ApplyRuntimeFeatures(_ agent.RuntimeFeatures) error {
	return nil
}

// ApplyExecutionInput applies execution-specific inputs like model selection.
// Returns error if model value is provided but empty.
func (arguments *CodexArguments) ApplyExecutionInput(executionInput agent.ExecutionInput) error {
	if executionInput.Model != nil {
		modelValue := strings.TrimSpace(*executionInput.Model)
		if modelValue == "" {
			return errors.New("model cannot be empty")
		}
		arguments.Model = &modelValue
	}

	return nil
}
