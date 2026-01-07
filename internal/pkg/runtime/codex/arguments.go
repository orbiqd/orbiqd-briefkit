package codex

import (
	"fmt"
	"strings"
)

type arguments struct {
	flags           map[string]bool
	values          map[string]string
	configOverrides map[string]string
}

func defaultArguments() *arguments {
	argument := &arguments{
		flags:           map[string]bool{},
		values:          map[string]string{},
		configOverrides: map[string]string{},
	}

	return argument
}

func (arguments *arguments) SetFlag(name string) {
	arguments.flags[name] = true
}

func (arguments *arguments) SetValue(name string, value any) error {
	valueStr, err := arguments.valueToString(value)
	if err != nil {
		return err
	}

	arguments.values[name] = valueStr

	return nil
}

func (arguments *arguments) SetConfigOverride(key string, value any) error {
	valueStr, err := arguments.valueToString(value)
	if err != nil {
		return err
	}

	arguments.values[key] = valueStr

	return nil
}

func (arguments *arguments) valueToString(value any) (string, error) {
	switch value := (value).(type) {
	case string:
		if strings.TrimSpace(value) == "" {
			return "", fmt.Errorf("empty string")
		}

		return value, nil

	case bool:
		if value {
			return "true", nil
		} else {
			return "false", nil
		}
	}

	return "", fmt.Errorf("unsupported type %t", value)
}

func (arguments *arguments) ToList() []string {
	var list []string

	for flag := range arguments.flags {
		list = append(list, fmt.Sprintf("--%s", flag))
	}

	for key, value := range arguments.values {
		list = append(list, fmt.Sprintf("--%s=\"%s\"", key, value))
	}

	for key, value := range arguments.configOverrides {
		list = append(list, fmt.Sprintf("--config=\"%s=%s\"", key, value))
	}

	return list
}
