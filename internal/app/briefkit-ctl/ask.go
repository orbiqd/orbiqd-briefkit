package briefkitctl

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

// AskCmd runs a prompt with specified model and options.
type AskCmd struct {
	Timeout        *time.Duration           `help:""`
	Model          *string                  `help:"Select model for execution."`
	ConversationID *briefkit.ConversationID `help:"Conversation ID for execution."`

	AgentID briefkit.AgentID `arg:"" help:"ID of the agent." required:"true"`
	Prompt  string           `arg:"" help:"Prompt to execute" required:"true"`
}

func (command *AskCmd) Run(ctx context.Context, client briefkit.Client) error {
	var askOptions []briefkit.AskOption

	if command.Timeout != nil {
		askOptions = append(askOptions, briefkit.AskWithTimeout(*command.Timeout))
	}

	if command.ConversationID != nil {
		askOptions = append(askOptions, briefkit.AskWithConversationID(*command.ConversationID))
	}

	if command.Model != nil {
		askOptions = append(askOptions, briefkit.AskWithModel(*command.Model))
	}

	result, err := client.Ask(ctx, command.AgentID, command.Prompt, askOptions...)
	if err != nil {
		return fmt.Errorf("ask: %w", err)
	}

	slog.Info("Execution finished successfully.", slog.String("conversationId", string(result.ConversationID)))

	fmt.Println(result.Response)

	return nil
}
