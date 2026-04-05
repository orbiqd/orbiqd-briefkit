package codex

import (
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCodexArguments_WhenCreated_ThenSetsDefaults(t *testing.T) {
	args := NewCodexArguments()

	require.NotNil(t, args)
	require.NotNil(t, args.JSON)
	assert.True(t, *args.JSON)
	assert.Nil(t, args.SkipGitRepoCheck)
	assert.Nil(t, args.Model)
	assert.Nil(t, args.SandboxMode)
	require.NotNil(t, args.ConfigOverrides)
	assert.Empty(t, args.ConfigOverrides)
}

func TestCodexArguments_ToSlice_WhenDefaults_ThenReturnsJsonFlag(t *testing.T) {
	args := NewCodexArguments()

	assert.Equal(t, []string{"--json"}, args.ToSlice())
}

func TestCodexArguments_ToSlice_WhenJSONDisabled_ThenReturnsEmptySlice(t *testing.T) {
	args := &CodexArguments{
		JSON: utils.ToPointer(false),
	}

	assert.Empty(t, args.ToSlice())
}

func TestCodexArguments_ToSlice_WhenFlagsAndModelSet_ThenReturnsOrderedArgs(t *testing.T) {
	args := &CodexArguments{
		JSON:             utils.ToPointer(true),
		SkipGitRepoCheck: utils.ToPointer(true),
		Model:            utils.ToPointer("gpt-4"),
		ConfigOverrides: map[string]any{
			"b": true,
			"a": "alpha",
		},
	}

	assert.Equal(t, []string{
		"--json",
		"--skip-git-repo-check",
		"--model=gpt-4",
		"--config=a=alpha",
		"--config=b=true",
	}, args.ToSlice())
}

func TestCodexArguments_ToSlice_WhenConfigOverridesProvided_ThenSortsAndFormats(t *testing.T) {
	args := &CodexArguments{
		ConfigOverrides: map[string]any{
			"z": "last",
			"a": true,
			"m": "middle",
		},
	}

	assert.Equal(t, []string{
		"--config=a=true",
		"--config=m=middle",
		"--config=z=last",
	}, args.ToSlice())
}

func TestCodexArguments_ToSlice_WhenConfigOverridesContainUnsupportedTypes_ThenSkipsThem(t *testing.T) {
	args := &CodexArguments{
		ConfigOverrides: map[string]any{
			"a": 123,
			"b": "two",
			"c": true,
		},
	}

	assert.Equal(t, []string{
		"--config=b=two",
		"--config=c=true",
	}, args.ToSlice())
}

func TestCodexArguments_ToSlice_WhenSandboxModeSet_ThenAddsSandboxFlag(t *testing.T) {
	args := &CodexArguments{
		SandboxMode: utils.ToPointer("workspace-write"),
	}

	assert.Equal(t, []string{"--sandbox=workspace-write"}, args.ToSlice())
}

func TestCodexArguments_ApplyRuntimeConfig_WhenNilOrNilPointer_ThenKeepsDefaults(t *testing.T) {
	var nilConfig *RuntimeConfig

	tests := []struct {
		name   string
		config briefkit.RuntimeConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name:   "nil pointer",
			config: nilConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := NewCodexArguments()

			err := args.ApplyRuntimeConfig(tt.config)

			require.NoError(t, err)
			assert.Nil(t, args.SkipGitRepoCheck)
		})
	}
}

func TestCodexArguments_ApplyRuntimeConfig_WhenRequireWorkspaceRepositoryFalse_ThenSetsSkipGitRepoCheck(t *testing.T) {
	args := NewCodexArguments()
	config := RuntimeConfig{
		RequireWorkspaceRepository: false,
	}

	err := args.ApplyRuntimeConfig(config)

	require.NoError(t, err)
	require.NotNil(t, args.SkipGitRepoCheck)
	assert.True(t, *args.SkipGitRepoCheck)
}

func TestCodexArguments_ApplyRuntimeConfig_WhenMarshalFails_ThenReturnsError(t *testing.T) {
	args := NewCodexArguments()
	config := make(chan int)

	err := args.ApplyRuntimeConfig(config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal codex runtime config")
}

func TestCodexArguments_ApplyRuntimeConfig_WhenUnmarshalFails_ThenReturnsError(t *testing.T) {
	args := NewCodexArguments()
	config := map[string]any{
		"requireWorkspaceRepository": "not-a-bool",
	}

	err := args.ApplyRuntimeConfig(config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal codex runtime config")
}

func TestCodexArguments_ApplyRuntimeConfig_WhenCustomStruct_ThenUnmarshalsCorrectly(t *testing.T) {
	type customConfig struct {
		RequireWorkspaceRepository bool `json:"requireWorkspaceRepository"`
	}

	args := NewCodexArguments()

	err := args.ApplyRuntimeConfig(customConfig{RequireWorkspaceRepository: false})

	require.NoError(t, err)
	require.NotNil(t, args.SkipGitRepoCheck)
	assert.True(t, *args.SkipGitRepoCheck)
}

func TestCodexArguments_ApplyRuntimeConfig_WhenMapMissingRequiredKey_ThenAppliesDefaults(t *testing.T) {
	args := NewCodexArguments()

	err := args.ApplyRuntimeConfig(map[string]any{
		"someOtherKey": "value",
	})

	require.NoError(t, err)
	assert.Nil(t, args.SkipGitRepoCheck)
}

func TestCodexArguments_ApplyRuntimeFeatures_WhenConfigOverridesNil_ThenKeepsNil(t *testing.T) {
	args := &CodexArguments{}
	features := briefkit.RuntimeFeatures{}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	assert.Nil(t, args.ConfigOverrides)
}

func TestCodexArguments_ApplyRuntimeFeatures_WhenFeaturesNil_ThenKeepsExistingOverrides(t *testing.T) {
	args := &CodexArguments{
		ConfigOverrides: map[string]any{
			"existing": "value",
		},
	}
	features := briefkit.RuntimeFeatures{}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"existing": "value",
	}, args.ConfigOverrides)
}

func TestCodexArguments_ApplyRuntimeFeatures_WhenEnableSandboxTrue_ThenSetsSandboxMode(t *testing.T) {
	args := &CodexArguments{}
	features := briefkit.RuntimeFeatures{
		EnableSandbox: utils.ToPointer(true),
	}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	require.NotNil(t, args.SandboxMode)
	assert.Equal(t, "workspace-write", *args.SandboxMode)
}

func TestCodexArguments_ApplyRuntimeFeatures_WhenEnableSandboxFalse_ThenSetsSandboxMode(t *testing.T) {
	args := &CodexArguments{
		SandboxMode: utils.ToPointer("workspace-write"),
	}
	features := briefkit.RuntimeFeatures{
		EnableSandbox: utils.ToPointer(false),
	}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	require.NotNil(t, args.SandboxMode)
	assert.Equal(t, "danger-full-access", *args.SandboxMode)
}

func TestCodexArguments_ApplyExecutionInput_WhenModelNil_ThenNoChange(t *testing.T) {
	args := NewCodexArguments()
	input := briefkit.ExecutionInput{
		Model: nil,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	assert.Nil(t, args.Model)
}

func TestCodexArguments_ApplyExecutionInput_WhenModelHasWhitespace_ThenTrimsAndSets(t *testing.T) {
	args := NewCodexArguments()
	model := "  gpt-4  "
	input := briefkit.ExecutionInput{
		Model: &model,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	require.NotNil(t, args.Model)
	assert.Equal(t, "gpt-4", *args.Model)
}

func TestCodexArguments_ApplyExecutionInput_WhenModelEmptyAfterTrim_ThenReturnsError(t *testing.T) {
	args := NewCodexArguments()
	model := "   "
	input := briefkit.ExecutionInput{
		Model: &model,
	}

	err := args.ApplyExecutionInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "model cannot be empty")
	assert.Nil(t, args.Model)
}

func TestCodexArguments_ApplyExecutionInput_WhenReasoningEffortSet_ThenAddsConfigOverride(t *testing.T) {
	args := NewCodexArguments()
	effort := "high"
	input := briefkit.ExecutionInput{
		ReasoningEffort: &effort,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	assert.Equal(t, "high", args.ConfigOverrides["model_reasoning_effort"])
}

func TestCodexArguments_ApplyExecutionInput_WhenReasoningEffortSetAndConfigOverridesNil_ThenInitializesAndSets(t *testing.T) {
	args := &CodexArguments{}
	effort := "xhigh"
	input := briefkit.ExecutionInput{
		ReasoningEffort: &effort,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	require.NotNil(t, args.ConfigOverrides)
	assert.Equal(t, "xhigh", args.ConfigOverrides["model_reasoning_effort"])
}

func TestCodexArguments_ApplyExecutionInput_WhenReasoningEffortHasWhitespace_ThenTrimsAndSets(t *testing.T) {
	args := NewCodexArguments()
	effort := "  xhigh  "
	input := briefkit.ExecutionInput{
		ReasoningEffort: &effort,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	assert.Equal(t, "xhigh", args.ConfigOverrides["model_reasoning_effort"])
}

func TestCodexArguments_ApplyExecutionInput_WhenReasoningEffortNil_ThenNoConfigOverride(t *testing.T) {
	args := NewCodexArguments()
	input := briefkit.ExecutionInput{
		ReasoningEffort: nil,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	assert.Empty(t, args.ConfigOverrides)
}

func TestCodexArguments_ApplyExecutionInput_WhenReasoningEffortEmptyAfterTrim_ThenReturnsError(t *testing.T) {
	args := NewCodexArguments()
	effort := "   "
	input := briefkit.ExecutionInput{
		ReasoningEffort: &effort,
	}

	err := args.ApplyExecutionInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reasoningEffort cannot be empty")
	assert.Empty(t, args.ConfigOverrides)
}
