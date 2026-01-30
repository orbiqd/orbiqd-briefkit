package gemini

import (
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGeminiArguments_WhenCreated_ThenSetsDefaults(t *testing.T) {
	args := NewGeminiArguments()

	require.NotNil(t, args)
	assert.Nil(t, args.Sandbox)
	assert.Nil(t, args.Model)
	assert.Nil(t, args.Resume)
	require.NotNil(t, args.OutputFormat)
	assert.Equal(t, "stream-json", *args.OutputFormat)
}

func TestGeminiArguments_ToSlice_WhenDefaults_ThenReturnsOutputFormatOnly(t *testing.T) {
	args := NewGeminiArguments()

	assert.Equal(t, []string{"--output-format=stream-json"}, args.ToSlice())
}

func TestGeminiArguments_ToSlice_WhenAllFieldsSet_ThenReturnsOrderedArgs(t *testing.T) {
	args := &GeminiArguments{
		Sandbox:      utils.ToPointer(true),
		Model:        utils.ToPointer("gemini-2.0-flash"),
		Resume:       utils.ToPointer("session-123"),
		OutputFormat: utils.ToPointer("stream-json"),
	}

	assert.Equal(t, []string{
		"--sandbox=true",
		"--model=gemini-2.0-flash",
		"--resume=session-123",
		"--output-format=stream-json",
	}, args.ToSlice())
}

func TestGeminiArguments_ToSlice_WhenSandboxFalse_ThenIncludesSandboxFalseFlag(t *testing.T) {
	args := &GeminiArguments{
		Sandbox: utils.ToPointer(false),
	}

	assert.Equal(t, []string{"--sandbox=false"}, args.ToSlice())
}

func TestGeminiArguments_ToSlice_WhenEmpty_ThenReturnsEmptySlice(t *testing.T) {
	args := &GeminiArguments{}

	assert.Empty(t, args.ToSlice())
}

func TestGeminiArguments_ApplyRuntimeConfig_WhenNil_ThenKeepsDefaults(t *testing.T) {
	args := NewGeminiArguments()

	err := args.ApplyRuntimeConfig(nil)

	require.NoError(t, err)
	require.NotNil(t, args.OutputFormat)
	assert.Equal(t, "stream-json", *args.OutputFormat)
}

func TestGeminiArguments_ApplyRuntimeConfig_WhenStruct_ThenNoChange(t *testing.T) {
	args := NewGeminiArguments()
	config := RuntimeConfig{}

	err := args.ApplyRuntimeConfig(config)

	require.NoError(t, err)
	require.NotNil(t, args.OutputFormat)
	assert.Equal(t, "stream-json", *args.OutputFormat)
}

func TestGeminiArguments_ApplyRuntimeConfig_WhenPointer_ThenNoChange(t *testing.T) {
	args := NewGeminiArguments()
	config := &RuntimeConfig{}

	err := args.ApplyRuntimeConfig(config)

	require.NoError(t, err)
	require.NotNil(t, args.OutputFormat)
	assert.Equal(t, "stream-json", *args.OutputFormat)
}

func TestGeminiArguments_ApplyRuntimeConfig_WhenNilPointer_ThenKeepsDefaults(t *testing.T) {
	args := NewGeminiArguments()
	var config *RuntimeConfig

	err := args.ApplyRuntimeConfig(config)

	require.NoError(t, err)
	require.NotNil(t, args.OutputFormat)
	assert.Equal(t, "stream-json", *args.OutputFormat)
}

func TestGeminiArguments_ApplyRuntimeConfig_WhenMarshalFails_ThenReturnsError(t *testing.T) {
	args := NewGeminiArguments()
	config := make(chan int)

	err := args.ApplyRuntimeConfig(config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal gemini runtime config")
}

func TestGeminiArguments_ApplyRuntimeFeatures_WhenNil_ThenNoChange(t *testing.T) {
	args := NewGeminiArguments()
	features := agent.RuntimeFeatures{}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	assert.Nil(t, args.Sandbox)
}

func TestGeminiArguments_ApplyRuntimeFeatures_WhenEnableSandboxTrue_ThenSetsSandbox(t *testing.T) {
	args := NewGeminiArguments()
	features := agent.RuntimeFeatures{
		EnableSandbox: utils.ToPointer(true),
	}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	require.NotNil(t, args.Sandbox)
	assert.True(t, *args.Sandbox)
}

func TestGeminiArguments_ApplyRuntimeFeatures_WhenEnableSandboxFalse_ThenSetsSandbox(t *testing.T) {
	args := NewGeminiArguments()
	features := agent.RuntimeFeatures{
		EnableSandbox: utils.ToPointer(false),
	}

	err := args.ApplyRuntimeFeatures(features)

	require.NoError(t, err)
	require.NotNil(t, args.Sandbox)
	assert.False(t, *args.Sandbox)
}

func TestGeminiArguments_ApplyExecutionInput_WhenModelNil_ThenNoChange(t *testing.T) {
	args := NewGeminiArguments()
	input := agent.ExecutionInput{
		Model: nil,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	assert.Nil(t, args.Model)
}

func TestGeminiArguments_ApplyExecutionInput_WhenModelHasWhitespace_ThenTrimsAndSets(t *testing.T) {
	args := NewGeminiArguments()
	model := "  gemini-2.0-flash  "
	input := agent.ExecutionInput{
		Model: &model,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	require.NotNil(t, args.Model)
	assert.Equal(t, "gemini-2.0-flash", *args.Model)
}

func TestGeminiArguments_ApplyExecutionInput_WhenModelEmptyAfterTrim_ThenReturnsError(t *testing.T) {
	args := NewGeminiArguments()
	model := "   "
	input := agent.ExecutionInput{
		Model: &model,
	}

	err := args.ApplyExecutionInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "model cannot be empty")
	assert.Nil(t, args.Model)
}

func TestGeminiArguments_ApplyExecutionInput_WhenConversationIDNil_ThenNoChange(t *testing.T) {
	args := NewGeminiArguments()
	input := agent.ExecutionInput{
		ConversationID: nil,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	assert.Nil(t, args.Resume)
}

func TestGeminiArguments_ApplyExecutionInput_WhenConversationIDProvided_ThenSetsResume(t *testing.T) {
	args := NewGeminiArguments()
	conversationID := agent.ConversationID("session-abc-123")
	input := agent.ExecutionInput{
		ConversationID: &conversationID,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	require.NotNil(t, args.Resume)
	assert.Equal(t, "session-abc-123", *args.Resume)
}

func TestGeminiArguments_ApplyExecutionInput_WhenConversationIDHasWhitespace_ThenTrimsAndSets(t *testing.T) {
	args := NewGeminiArguments()
	conversationID := agent.ConversationID("  session-xyz  ")
	input := agent.ExecutionInput{
		ConversationID: &conversationID,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	require.NotNil(t, args.Resume)
	assert.Equal(t, "session-xyz", *args.Resume)
}

func TestGeminiArguments_ApplyExecutionInput_WhenConversationIDEmptyAfterTrim_ThenReturnsError(t *testing.T) {
	args := NewGeminiArguments()
	conversationID := agent.ConversationID("   ")
	input := agent.ExecutionInput{
		ConversationID: &conversationID,
	}

	err := args.ApplyExecutionInput(input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "conversationID cannot be empty")
	assert.Nil(t, args.Resume)
}

func TestGeminiArguments_ApplyExecutionInput_WhenBothModelAndConversationID_ThenSetsBoth(t *testing.T) {
	args := NewGeminiArguments()
	model := "gemini-2.0-flash"
	conversationID := agent.ConversationID("session-123")
	input := agent.ExecutionInput{
		Model:          &model,
		ConversationID: &conversationID,
	}

	err := args.ApplyExecutionInput(input)

	require.NoError(t, err)
	require.NotNil(t, args.Model)
	assert.Equal(t, "gemini-2.0-flash", *args.Model)
	require.NotNil(t, args.Resume)
	assert.Equal(t, "session-123", *args.Resume)
}
