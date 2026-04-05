package claude

import (
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeArguments_ToSlice(t *testing.T) {
	tests := []struct {
		name     string
		args     *ClaudeArguments
		expected []string
	}{
		{
			name: "default arguments",
			args: NewClaudeArguments(),
			expected: []string{
				"--print",
				"--verbose",
				"--output-format=stream-json",
			},
		},
		{
			name: "all arguments set",
			args: &ClaudeArguments{
				Print:           utils.ToPointer(true),
				Verbose:         utils.ToPointer(true),
				OutputFormat:    utils.ToPointer("json"),
				Model:           utils.ToPointer("claude-3-5-sonnet"),
				ResumeSessionID: utils.ToPointer("session-123"),
				DisallowedTools: []string{"WebSearch", "Bash"},
				Settings:        map[string]any{"key": "value"},
			},
			expected: []string{
				"--print",
				"--verbose",
				"--output-format=json",
				"--model=claude-3-5-sonnet",
				"--resume=session-123",
				"--disallowed-tools=WebSearch,Bash",
				`--settings={"key":"value"}`,
			},
		},
		{
			name: "boolean flags false",
			args: &ClaudeArguments{
				Print:   utils.ToPointer(false),
				Verbose: utils.ToPointer(false),
			},
			expected: nil,
		},
		{
			name: "empty values",
			args: &ClaudeArguments{
				OutputFormat:    utils.ToPointer(""),
				Model:           utils.ToPointer(""),
				ResumeSessionID: utils.ToPointer(""),
			},
			expected: []string{
				"--output-format=",
				"--model=",
				"--resume=",
			},
		},
		{
			name: "permission mode set",
			args: &ClaudeArguments{
				PermissionMode: utils.ToPointer("bypassPermissions"),
			},
			expected: []string{
				"--permission-mode=bypassPermissions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.args.ToSlice())
		})
	}
}

func TestClaudeArguments_ApplyExecutionInput(t *testing.T) {
	t.Run("applies model and conversation id", func(t *testing.T) {
		args := NewClaudeArguments()
		convID := briefkit.ConversationID("test-conv-id")
		input := briefkit.ExecutionInput{
			Model:          utils.ToPointer("test-model"),
			ConversationID: &convID,
		}

		err := args.ApplyExecutionInput(input)
		require.NoError(t, err)
		assert.Equal(t, "test-model", *args.Model)
		assert.Equal(t, "test-conv-id", *args.ResumeSessionID)
	})

	t.Run("skips nil values", func(t *testing.T) {
		args := NewClaudeArguments()
		input := briefkit.ExecutionInput{
			Model:          nil,
			ConversationID: nil,
		}

		err := args.ApplyExecutionInput(input)
		require.NoError(t, err)
		assert.Nil(t, args.Model)
		assert.Nil(t, args.ResumeSessionID)
	})

	t.Run("trims whitespace from model", func(t *testing.T) {
		args := NewClaudeArguments()
		input := briefkit.ExecutionInput{
			Model: utils.ToPointer("  claude-sonnet-4-5  "),
		}

		err := args.ApplyExecutionInput(input)
		require.NoError(t, err)
		assert.Equal(t, "claude-sonnet-4-5", *args.Model)
	})

	t.Run("returns error for empty model", func(t *testing.T) {
		args := NewClaudeArguments()
		input := briefkit.ExecutionInput{
			Model: utils.ToPointer(""),
		}

		err := args.ApplyExecutionInput(input)
		require.Error(t, err)
		assert.EqualError(t, err, "model cannot be empty")
	})

	t.Run("returns error for whitespace-only model", func(t *testing.T) {
		args := NewClaudeArguments()
		input := briefkit.ExecutionInput{
			Model: utils.ToPointer("   "),
		}

		err := args.ApplyExecutionInput(input)
		require.Error(t, err)
		assert.EqualError(t, err, "model cannot be empty")
	})
}

func TestClaudeArguments_ApplyRuntimeFeatures(t *testing.T) {
	t.Run("EnableSandbox=nil uses runtime default", func(t *testing.T) {
		args := NewClaudeArguments()
		features := briefkit.RuntimeFeatures{}

		err := args.ApplyRuntimeFeatures(features)
		require.NoError(t, err)
		assert.Nil(t, args.PermissionMode)
		assert.Nil(t, args.Settings["sandbox"])
	})

	t.Run("EnableSandbox=false sets bypassPermissions", func(t *testing.T) {
		args := NewClaudeArguments()
		features := briefkit.RuntimeFeatures{
			EnableSandbox: utils.ToPointer(false),
		}

		err := args.ApplyRuntimeFeatures(features)
		require.NoError(t, err)
		require.NotNil(t, args.PermissionMode)
		assert.Equal(t, "bypassPermissions", *args.PermissionMode)
	})

	t.Run("EnableSandbox=true enables sandbox via settings", func(t *testing.T) {
		args := NewClaudeArguments()
		features := briefkit.RuntimeFeatures{
			EnableSandbox: utils.ToPointer(true),
		}

		err := args.ApplyRuntimeFeatures(features)
		require.NoError(t, err)
		assert.Nil(t, args.PermissionMode)
		require.NotNil(t, args.Settings["sandbox"])
		sandboxSettings := args.Settings["sandbox"].(map[string]any)
		assert.True(t, sandboxSettings["enabled"].(bool))
	})
}

func TestClaudeArguments_ApplyRuntimeConfig(t *testing.T) {
	t.Run("applies direct Config", func(t *testing.T) {
		args := NewClaudeArguments()
		config := RuntimeConfig{}

		err := args.ApplyRuntimeConfig(config)
		assert.NoError(t, err)
	})

	t.Run("applies pointer to Config", func(t *testing.T) {
		args := NewClaudeArguments()
		config := &RuntimeConfig{}

		err := args.ApplyRuntimeConfig(config)
		assert.NoError(t, err)
	})

	t.Run("handles nil config", func(t *testing.T) {
		args := NewClaudeArguments()
		err := args.ApplyRuntimeConfig(nil)
		assert.NoError(t, err)
	})

	t.Run("applies map config via json roundtrip", func(t *testing.T) {
		args := NewClaudeArguments()
		// Using a map triggers the default case in the switch
		config := map[string]any{
			"debug": true,
		}

		err := args.ApplyRuntimeConfig(config)
		assert.NoError(t, err)
	})

	t.Run("returns error for unmarshalable config", func(t *testing.T) {
		args := NewClaudeArguments()
		// Channels cannot be marshaled to JSON
		config := make(chan int)

		err := args.ApplyRuntimeConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "marshal claude runtime config")
	})
}
