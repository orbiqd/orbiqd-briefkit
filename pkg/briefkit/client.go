package briefkit

import (
	"context"
	"time"
)

// AskOption customizes Ask behavior.
type AskOption func(options *AskOptions)

// AskOptions holds optional Ask parameters.
type AskOptions struct {
	// ConversationID resumes an existing session when provided.
	ConversationID *ConversationID

	// Model overrides the default model for this execution.
	Model *string

	// ReasoningEffort overrides the reasoning intensity for runtimes that support it.
	// Accepted values are runtime-specific (e.g. low/medium/high/xhigh for Codex, low/medium/high/max for Claude).
	// Returns an error at execution time if the runtime does not support reasoning effort.
	ReasoningEffort *string

	// Timeout limits the total execution duration.
	Timeout *time.Duration

	// Workspace is a URI identifying the workspace source for the execution.
	// When set, the runner provisions an isolated copy and runs the agent there.
	// Supported schemes: dir:// (e.g. dir:///path/to/project).
	Workspace *string
}

// AskWithConversationID sets the conversation ID for continuation.
func AskWithConversationID(conversationID ConversationID) AskOption {
	return func(options *AskOptions) {
		options.ConversationID = &conversationID
	}
}

// AskWithModel sets the model override.
func AskWithModel(model string) AskOption {
	return func(options *AskOptions) {
		options.Model = &model
	}
}

// AskWithReasoningEffort sets the reasoning effort override.
func AskWithReasoningEffort(effort string) AskOption {
	return func(options *AskOptions) {
		options.ReasoningEffort = &effort
	}
}

// AskWithTimeout sets the execution timeout.
func AskWithTimeout(timeout time.Duration) AskOption {
	return func(options *AskOptions) {
		options.Timeout = &timeout
	}
}

// AskWithWorkspace sets the workspace URI for isolated execution.
func AskWithWorkspace(uri string) AskOption {
	return func(options *AskOptions) {
		options.Workspace = &uri
	}
}

// AskResult represents the Ask output.
type AskResult struct {
	// ConversationID identifies the session for follow-up requests.
	ConversationID ConversationID

	// Response is the agent's final output.
	Response string
}

// Client exposes public BriefKit actions.
type Client interface {
	// Ask runs a single prompt with the selected agent.
	Ask(ctx context.Context, agentID AgentID, prompt string, opts ...AskOption) (*AskResult, error)
}
