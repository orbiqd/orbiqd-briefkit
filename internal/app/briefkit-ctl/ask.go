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
	Auto           bool                     `help:"Enable automatic mode"`
	Timeout        time.Duration            `default:"5m"`
	Model          *string                  `help:"Select model for execution."`
	ConversationID *briefkit.ConversationID `help:"Conversation ID for execution."`

	AgentID briefkit.AgentID `arg:"" help:"ID of the agent." required:"true"`
	Prompt  string           `arg:"" help:"Prompt to execute" required:"true"`
}

func (command *AskCmd) Run(ctx context.Context, client briefkit.Client) error {
	result, err := client.Ask(ctx, command.AgentID, command.Prompt)
	if err != nil {
		return fmt.Errorf("ask: %w", err)
	}

	slog.Info("Execution finished successfully.", slog.String("conversationId", string(result.ConversationID)))

	fmt.Println(result.Response)

	return nil
}
