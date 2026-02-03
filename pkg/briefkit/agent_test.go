package briefkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentIDValidate(t *testing.T) {
	tests := []struct {
		name  string
		id    AgentID
		valid bool
	}{
		{
			name:  "valid single segment",
			id:    "codex",
			valid: true,
		},
		{
			name:  "valid multi segment",
			id:    "claude-code",
			valid: true,
		},
		{
			name:  "valid with digits",
			id:    "codex-2",
			valid: true,
		},
		{
			name:  "invalid empty",
			id:    "",
			valid: false,
		},
		{
			name:  "invalid leading dash",
			id:    "-codex",
			valid: false,
		},
		{
			name:  "invalid trailing dash",
			id:    "codex-",
			valid: false,
		},
		{
			name:  "invalid uppercase",
			id:    "Codex",
			valid: false,
		},
		{
			name:  "invalid underscore",
			id:    "claude_code",
			valid: false,
		},
		{
			name:  "invalid double dash",
			id:    "codex--v2",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if tt.valid {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, ErrAgentIDInvalid)
		})
	}
}
