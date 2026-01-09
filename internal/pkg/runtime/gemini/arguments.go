package gemini

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

type arguments struct {
	flags  map[string]bool
	values map[string]string
}

func defaultArguments() *arguments {
	return &arguments{
		flags:  map[string]bool{},
		values: map[string]string{},
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

func (a *arguments) valueToString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", errors.New("empty string")
		}
		return v, nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case agent.ConversationID:
		if strings.TrimSpace(string(v)) == "" {
			return "", errors.New("empty string")
		}
		return string(v), nil
	case int:
		return strconv.Itoa(v), nil
	default:
		return "", fmt.Errorf("unsupported type %T", value)
	}
}

func (a *arguments) ToList() []string {
	var list []string

	for flag := range a.flags {
		list = append(list, "--"+flag)
	}

	for key, value := range a.values {
		list = append(list, fmt.Sprintf("--%s=%s", key, value))
	}

	return list
}
