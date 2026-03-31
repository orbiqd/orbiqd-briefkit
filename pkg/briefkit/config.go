package briefkit

import (
	"context"
	"errors"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

type Config struct {
	Runtime struct {
		Kind    RuntimeKind     `json:"kind"`
		Config  RuntimeConfig   `json:"config"`
		Feature RuntimeFeatures `json:"feature,omitempty"`
	} `json:"runtime"`

	// Timeout defines the default execution timeout for the agent.
	Timeout *utils.Duration `json:"timeout,omitempty"`
}

type ConfigRepository interface {
	Exists(ctx context.Context, id AgentID) (bool, error)
	Get(ctx context.Context, id AgentID) (Config, error)
	Update(ctx context.Context, id AgentID, config Config) error
	List(ctx context.Context) ([]AgentID, error)
}

var (
	// ErrAgentConfigNotFound indicates the agent configuration does not exist.
	ErrAgentConfigNotFound = errors.New("agent config not found")
)
