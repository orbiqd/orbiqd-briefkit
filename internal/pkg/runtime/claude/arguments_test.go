package claude

import (
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
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
		convID := agent.ConversationID("test-conv-id")
		input := agent.ExecutionInput{
			Model:          utils.ToPointer("test-model"),
			ConversationID: &convID,
		}

		err := args.ApplyExecutionInput(input)
		assert.NoError(t, err)
		assert.Equal(t, "test-model", *args.Model)
		assert.Equal(t, "test-conv-id", *args.ResumeSessionID)
	})

	t.Run("skips nil values", func(t *testing.T) {
		args := NewClaudeArguments()
		input := agent.ExecutionInput{
			Model:          nil,
			ConversationID: nil,
		}

		err := args.ApplyExecutionInput(input)
		assert.NoError(t, err)
		assert.Nil(t, args.Model)
		assert.Nil(t, args.ResumeSessionID)
	})
}

func TestClaudeArguments_ApplyRuntimeFeatures(t *testing.T) {
	t.Run("disallows WebSearch when EnableWebSearch is false", func(t *testing.T) {
		args := NewClaudeArguments()
		features := agent.RuntimeFeatures{
			EnableWebSearch: utils.ToPointer(false),
		}

		err := args.ApplyRuntimeFeatures(features)
		assert.NoError(t, err)
		assert.Contains(t, args.DisallowedTools, "WebSearch")
	})

	t.Run("does nothing when EnableWebSearch is true", func(t *testing.T) {
		args := NewClaudeArguments()
		features := agent.RuntimeFeatures{
			EnableWebSearch: utils.ToPointer(true),
		}

		err := args.ApplyRuntimeFeatures(features)
		assert.NoError(t, err)
		assert.Empty(t, args.DisallowedTools)
	})

	t.Run("does nothing when EnableWebSearch is nil", func(t *testing.T) {
		args := NewClaudeArguments()
		features := agent.RuntimeFeatures{
			EnableWebSearch: nil,
		}

		err := args.ApplyRuntimeFeatures(features)
		assert.NoError(t, err)
		assert.Empty(t, args.DisallowedTools)
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
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "marshal claude runtime config")
	})
}
