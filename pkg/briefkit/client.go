package briefkit

import "context"

type AskOption func(options *AskOptions)
type AskOptions struct {
	model *string
}

func AskWithModel(model string) AskOption {
	return func(options *AskOptions) {
		options.model = &model
	}
}

type AskResult struct {
}

type Client interface {
	Ask(ctx context.Context, prompt string, opts ...AskOption) (*AskResult, error)
}
