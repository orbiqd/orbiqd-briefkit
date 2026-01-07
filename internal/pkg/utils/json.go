package utils

import (
	"encoding/json"
	"fmt"
)

func AnyToStruct[O any](input any) (*O, error) {
	var output O

	jsonBuffer, err := json.Marshal(&input)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	err = json.Unmarshal(jsonBuffer, &output)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &output, nil
}
