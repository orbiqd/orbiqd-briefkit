package briefkit

import (
	"errors"
	"regexp"
)

// ConversationID identifies the session to continue an ongoing conversation with an agent.
type ConversationID string

// AgentID is the stable identifier of an agent definition.
type AgentID string

var agentIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Validate checks whether the agent identifier is present.
func (a AgentID) Validate() error {
	if !agentIDPattern.MatchString(string(a)) {
		return ErrAgentIDInvalid
	}

	return nil
}

var (
	// ErrAgentIDInvalid indicates the agent identifier is missing or invalid.
	ErrAgentIDInvalid = errors.New("agent id invalid")
)
