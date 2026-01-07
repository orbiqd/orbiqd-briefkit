package utils

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDuration_MarshalJSON(t *testing.T) {
	t.Run("positive duration", func(t *testing.T) {
		d := Duration(10 * time.Second)
		b, err := json.Marshal(d)
		require.NoError(t, err)
		assert.Equal(t, `"10s"`, string(b))
	})

	t.Run("zero duration", func(t *testing.T) {
		d := Duration(0)
		b, err := json.Marshal(d)
		require.NoError(t, err)
		assert.Equal(t, `"0s"`, string(b))
	})
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    []byte
		expectedDur time.Duration
		expectErr   bool
	}{
		{
			name:        "from valid string",
			jsonData:    []byte(`"1h30m"`),
			expectedDur: 90 * time.Minute,
			expectErr:   false,
		},
		{
			name:        "from valid number",
			jsonData:    []byte(`300000000000`), // 5 minutes in nanoseconds
			expectedDur: 5 * time.Minute,
			expectErr:   false,
		},
		{
			name:      "from invalid string",
			jsonData:  []byte(`"invalid duration"`),
			expectErr: true,
		},
		{
			name:      "from unsupported type",
			jsonData:  []byte(`true`),
			expectErr: true,
		},
		{
			name:      "from malformed json",
			jsonData:  []byte(`{"key":`),
			expectErr: true,
		},
		{
			name:        "from zero string",
			jsonData:    []byte(`"0s"`),
			expectedDur: 0,
			expectErr:   false,
		},
		{
			name:        "from zero number",
			jsonData:    []byte(`0`),
			expectedDur: 0,
			expectErr:   false,
		},
		{
			name:        "from null",
			jsonData:    []byte(`null`),
			expectedDur: 0,
			// Unmarshaling null into a pointer receiver is a no-op in encoding/json.
			// The value remains the zero value for the type.
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal(tt.jsonData, &d)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDur, time.Duration(d))
			}
		})
	}
}

// TestDuration_StructIntegration tests the marshalling and unmarshalling
// of the Duration type when embedded in another struct.
func TestDuration_StructIntegration(t *testing.T) {
	type Config struct {
		Timeout Duration `json:"timeout"`
	}

	t.Run("marshal struct with duration", func(t *testing.T) {
		cfg := Config{Timeout: Duration(15 * time.Minute)}
		b, err := json.Marshal(cfg)
		require.NoError(t, err)
		assert.JSONEq(t, `{"timeout": "15m0s"}`, string(b))
	})

	t.Run("unmarshal struct with duration from string", func(t *testing.T) {
		jsonData := `{"timeout": "30s"}`
		var cfg Config
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, time.Duration(cfg.Timeout))
	})

	t.Run("unmarshal struct with duration from number", func(t *testing.T) {
		jsonData := `{"timeout": 90000000000}` // 90 seconds
		var cfg Config
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)
		assert.Equal(t, 90*time.Second, time.Duration(cfg.Timeout))
	})
}
