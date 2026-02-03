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

	// Timeout limits the total execution duration.
	Timeout *time.Duration
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

// AskWithTimeout sets the execution timeout.
func AskWithTimeout(timeout time.Duration) AskOption {
	return func(options *AskOptions) {
		options.Timeout = &timeout
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
