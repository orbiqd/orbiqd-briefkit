package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pointerTestStruct struct {
	Name string

	Count int
}

func TestToPointer_WhenIntProvided_ThenReturnsPointerWithSameValue(t *testing.T) {
	value := 42

	result := ToPointer(value)

	require.NotNil(t, result)
	assert.Equal(t, value, *result)
}

func TestToPointer_WhenStringProvided_ThenReturnsPointerWithSameValue(t *testing.T) {
	value := "briefkit"

	result := ToPointer(value)

	require.NotNil(t, result)
	assert.Equal(t, value, *result)
}

func TestToPointer_WhenStructProvided_ThenReturnsPointerWithSameValue(t *testing.T) {
	value := pointerTestStruct{Name: "alpha", Count: 3}

	result := ToPointer(value)

	require.NotNil(t, result)
	assert.Equal(t, value, *result)
}

func TestToPointer_WhenPointerProvided_ThenReturnsPointerToSamePointerValue(t *testing.T) {
	value := 99
	input := &value

	result := ToPointer(input)

	require.NotNil(t, result)
	assert.Equal(t, input, *result)
}
