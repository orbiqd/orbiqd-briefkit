package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a new type based on time.Duration that allows for marshalling/unmarshalling
// to/from a human-readable string format (e.g., "1h30m").
type Duration time.Duration

// MarshalJSON implements the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This implementation can unmarshal both the numeric (nanoseconds) and string formats.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case nil:
		// Handle JSON null by setting to zero value
		*d = 0
		return nil
	case float64:
		// Handle the numeric format (default Go duration serialization)
		*d = Duration(value)
		return nil
	case string:
		// Handle the string format
		parsed, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(parsed)
		return nil
	default:
		return fmt.Errorf("invalid duration: unsupported type %T", value)
	}
}
