package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnyToStruct_WhenInputStruct_ThenReturnsEquivalentStruct(t *testing.T) {
	type inputModel struct {
		Name string

		Count int
	}

	input := inputModel{Name: "alpha", Count: 2}

	result, err := AnyToStruct[inputModel](input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, input, *result)
}

func TestAnyToStruct_WhenInputMap_ThenMapsFieldsCorrectly(t *testing.T) {
	type inputModel struct {
		Name string

		Count int
	}

	input := map[string]any{
		"name":  "bravo",
		"count": 7,
	}

	result, err := AnyToStruct[inputModel](input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, inputModel{Name: "bravo", Count: 7}, *result)
}

func TestAnyToStruct_WhenInputHasNestedFields_ThenPreservesStructure(t *testing.T) {
	type nestedItem struct {
		ID string

		Value int
	}

	type nestedMeta struct {
		Enabled bool
	}

	type inputModel struct {
		Items []nestedItem

		Meta nestedMeta
	}

	input := map[string]any{
		"items": []map[string]any{
			{"id": "item-1", "value": 3},
			{"id": "item-2", "value": 5},
		},
		"meta": map[string]any{
			"enabled": true,
		},
	}

	result, err := AnyToStruct[inputModel](input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, inputModel{
		Items: []nestedItem{
			{ID: "item-1", Value: 3},
			{ID: "item-2", Value: 5},
		},
		Meta: nestedMeta{Enabled: true},
	}, *result)
}

func TestAnyToStruct_WhenInputIsNil_ThenReturnsZeroValueStruct(t *testing.T) {
	type inputModel struct {
		Name string

		Count int
	}

	var input any

	result, err := AnyToStruct[inputModel](input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, inputModel{}, *result)
}

func TestAnyToStruct_WhenMarshalFails_ThenReturnsWrappedError(t *testing.T) {
	input := func() {}

	result, err := AnyToStruct[struct{}](input)

	require.Error(t, err)
	assert.Nil(t, result)
	require.ErrorContains(t, err, "json marshal:")

	var unsupportedTypeError *json.UnsupportedTypeError
	require.ErrorAs(t, err, &unsupportedTypeError)
}

func TestAnyToStruct_WhenUnmarshalFails_ThenReturnsWrappedError(t *testing.T) {
	type inputModel struct {
		Count int
	}

	input := map[string]any{
		"count": "not-a-number",
	}

	result, err := AnyToStruct[inputModel](input)

	require.Error(t, err)
	assert.Nil(t, result)
	require.ErrorContains(t, err, "json unmarshal:")

	var unmarshalTypeError *json.UnmarshalTypeError
	require.ErrorAs(t, err, &unmarshalTypeError)
}
