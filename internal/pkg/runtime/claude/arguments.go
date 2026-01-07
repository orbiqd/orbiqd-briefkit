package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

type arguments struct {
	flags            map[string]bool
	values           map[string]string
	settingsOverride map[string]any
}

func defaultArguments() *arguments {
	return &arguments{
		flags:            map[string]bool{},
		values:           map[string]string{},
		settingsOverride: map[string]any{},
	}
}

func (a *arguments) SetFlag(name string) {
	a.flags[name] = true
}

func (a *arguments) SetValue(name string, value any) error {
	valueStr, err := a.valueToString(value)
	if err != nil {
		return err
	}

	a.values[name] = valueStr
	return nil
}

func (a *arguments) SetSettingsOverride(key string, value any) error {
	a.settingsOverride[key] = value
	return nil
}

func (a *arguments) valueToString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", fmt.Errorf("empty string")
		}
		return v, nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case agent.ConversationID:
		return string(v), nil
	default:
		return "", fmt.Errorf("unsupported type %T", value)
	}
}

func (a *arguments) ToList() []string {
	var list []string

	for flag := range a.flags {
		list = append(list, fmt.Sprintf("--%s", flag))
	}

	for key, value := range a.values {
		list = append(list, fmt.Sprintf("--%s=%s", key, value))
	}

	if len(a.settingsOverride) > 0 {
		settingsJSON, err := json.Marshal(a.settingsOverride)
		if err == nil {
			list = append(list, fmt.Sprintf("--settings=%s", string(settingsJSON)))
		}
	}

	return list
}
