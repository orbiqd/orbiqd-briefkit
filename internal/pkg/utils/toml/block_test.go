package toml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasKey_WhenBlockIsNil_ThenReturnsFalse(t *testing.T) {
	var block *Block

	assert.False(t, block.HasKey("key"))
}

func TestHasKey_WhenFileIsNil_ThenReturnsFalse(t *testing.T) {
	block := &Block{}

	assert.False(t, block.HasKey("key"))
}

func TestHasKey_WhenKeyIsBlank_ThenReturnsFalse(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{name: "Empty", key: ""},
		{name: "Whitespace", key: "  "},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := newParsedTestFile("[section]\nkey = 1\n")
			require.NoError(t, err)
			block := file.OpenBlock("section")

			assert.False(t, block.HasKey(testCase.key))
		})
	}
}

func TestHasKey_WhenBlockNotFound_ThenReturnsFalse(t *testing.T) {
	file, err := newParsedTestFile("key = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.False(t, block.HasKey("key"))
}

func TestHasKey_WhenKeyExistsInBlock_ThenReturnsTrue(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.True(t, block.HasKey("key"))
}

func TestHasKey_WhenKeyExistsInOtherBlock_ThenReturnsFalse(t *testing.T) {
	file, err := newParsedTestFile("[section]\n\n[other]\nkey = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.False(t, block.HasKey("key"))
}

func TestHasKey_WhenKeyLineIsCommentedOut_ThenReturnsFalse(t *testing.T) {
	file, err := newParsedTestFile("[section]\n# key = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.False(t, block.HasKey("key"))
}

func TestHasKey_WhenKeyHasInlineComment_ThenReturnsTrue(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1 # note\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.True(t, block.HasKey("key"))
}

func TestHasKey_WhenKeyHasExtraSpaces_ThenReturnsTrue(t *testing.T) {
	file, err := newParsedTestFile("[section]\n  key  =  1  \n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.True(t, block.HasKey("key"))
}

func TestHasKey_WhenFileUsesCRLF_ThenReturnsTrue(t *testing.T) {
	file, err := newParsedTestFile("[section]\r\nkey = 1\r\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	assert.True(t, block.HasKey("key"))
}

func TestSetKey_WhenBlockIsNil_ThenDoesNothing(t *testing.T) {
	var block *Block

	assert.NotPanics(t, func() {
		block.SetKey("key", "value")
	})
}

func TestSetKey_WhenFileIsNil_ThenDoesNothing(t *testing.T) {
	block := &Block{blockName: "section"}

	assert.NotPanics(t, func() {
		block.SetKey("key", "value")
	})
}

func TestSetKey_WhenKeyIsBlank_ThenDoesNothing(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{name: "Empty", key: ""},
		{name: "Whitespace", key: "  "},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := newParsedTestFile("[section]\nkey = 1\n")
			require.NoError(t, err)
			block := file.OpenBlock("section")
			original := string(file.buffer)

			block.SetKey(testCase.key, "value")

			assert.Equal(t, original, string(file.buffer))
		})
	}
}

func TestSetKey_WhenBlockExistsAndKeyExists_ThenUpdatesValue(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("key", 2)

	assert.Equal(t, "[section]\nkey = 2\n", string(file.buffer))
}

func TestSetKey_WhenBlockExistsAndKeyMissing_ThenAppendsKeyToBlock(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n[other]\nvalue = 2\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("new", 2)

	assert.Equal(t, "[section]\nkey = 1\nnew = 2\n[other]\nvalue = 2\n", string(file.buffer))
}

func TestSetKey_WhenBlockMissing_ThenAppendsBlockAtFileEnd(t *testing.T) {
	file, err := newParsedTestFile("key = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("new", 2)

	assert.Equal(t, "key = 1\n[section]\nnew = 2\n", string(file.buffer))
}

func TestSetKey_WhenFileEndsWithoutNewline_ThenInsertsSeparatorNewline(t *testing.T) {
	file, err := newParsedTestFile("key = 1")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("new", 2)

	assert.Equal(t, "key = 1\n[section]\nnew = 2\n", string(file.buffer))
}

func TestSetKey_WhenBlockEndsWithoutNewline_ThenInsertsSeparatorNewline(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("new", 2)

	assert.Equal(t, "[section]\nkey = 1\nnew = 2\n", string(file.buffer))
}

func TestSetKey_WhenFileUsesCRLF_ThenPreservesCRLF(t *testing.T) {
	file, err := newParsedTestFile("[section]\r\nkey = 1\r\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("new", 2)

	assert.Equal(t, "[section]\r\nkey = 1\r\nnew = 2\r\n", string(file.buffer))
}

func TestSetKey_WhenMainBlockAndKeyMissing_ThenAppendsToMain(t *testing.T) {
	file, err := newParsedTestFile("key = 1\n[section]\nvalue = 2\n")
	require.NoError(t, err)
	block := file.OpenBlock(MainBlock)

	block.SetKey("new", 2)

	assert.Equal(t, "key = 1\nnew = 2\n[section]\nvalue = 2\n", string(file.buffer))
}

func TestSetKey_WhenValueIsString_ThenQuotesValue(t *testing.T) {
	file, err := newParsedTestFile("[section]\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("name", "demo")

	assert.Equal(t, "[section]\nname = \"demo\"\n", string(file.buffer))
}

func TestSetKey_WhenValueIsBoolOrNumber_ThenFormatsPlain(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		value    any
		expected string
	}{
		{name: "BoolTrue", key: "enabled", value: true, expected: "[section]\nenabled = true\n"},
		{name: "BoolFalse", key: "enabled", value: false, expected: "[section]\nenabled = false\n"},
		{name: "Int", key: "count", value: 3, expected: "[section]\ncount = 3\n"},
		{name: "Int8", key: "count", value: int8(-5), expected: "[section]\ncount = -5\n"},
		{name: "Int16", key: "count", value: int16(12), expected: "[section]\ncount = 12\n"},
		{name: "Int32", key: "count", value: int32(34), expected: "[section]\ncount = 34\n"},
		{name: "Int64", key: "count", value: int64(56), expected: "[section]\ncount = 56\n"},
		{name: "Uint", key: "count", value: uint(7), expected: "[section]\ncount = 7\n"},
		{name: "Uint8", key: "count", value: uint8(8), expected: "[section]\ncount = 8\n"},
		{name: "Uint16", key: "count", value: uint16(16), expected: "[section]\ncount = 16\n"},
		{name: "Uint32", key: "count", value: uint32(32), expected: "[section]\ncount = 32\n"},
		{name: "Uint64", key: "count", value: uint64(64), expected: "[section]\ncount = 64\n"},
		{name: "Float32", key: "ratio", value: float32(1.5), expected: "[section]\nratio = 1.5\n"},
		{name: "Float64", key: "ratio", value: 1.5, expected: "[section]\nratio = 1.5\n"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := newParsedTestFile("[section]\n")
			require.NoError(t, err)
			block := file.OpenBlock("section")

			block.SetKey(testCase.key, testCase.value)

			assert.Equal(t, testCase.expected, string(file.buffer))
		})
	}
}

func TestSetKey_WhenOtherBlocksExist_ThenKeepsTheirContentIntact(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n[other]\nvalue = 2\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("key", "longer")

	assert.Equal(t, "[section]\nkey = \"longer\"\n[other]\nvalue = 2\n", string(file.buffer))
}

func TestSetKey_WhenValueIsOtherType_ThenFormatsWithSprint(t *testing.T) {
	type customValue struct {
		name string
	}

	file, err := newParsedTestFile("[section]\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	block.SetKey("custom", customValue{name: "demo"})

	assert.Equal(t, "[section]\ncustom = {demo}\n", string(file.buffer))
}

func TestUnsetKey_WhenBlockIsNil_ThenReturnsBlockNotFound(t *testing.T) {
	var block *Block

	err := block.UnsetKey("key")

	require.ErrorIs(t, err, ErrBlockNotFound)
}

func TestUnsetKey_WhenFileIsNil_ThenReturnsBlockNotFound(t *testing.T) {
	block := &Block{blockName: "section"}

	err := block.UnsetKey("key")

	require.ErrorIs(t, err, ErrBlockNotFound)
}

func TestUnsetKey_WhenKeyIsBlank_ThenReturnsKeyNotFound(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{name: "Empty", key: ""},
		{name: "Whitespace", key: "  "},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := newParsedTestFile("[section]\nkey = 1\n")
			require.NoError(t, err)
			block := file.OpenBlock("section")

			err = block.UnsetKey(testCase.key)

			require.ErrorIs(t, err, ErrKeyNotFound)
		})
	}
}

func TestUnsetKey_WhenBlockIsMissing_ThenReturnsBlockNotFound(t *testing.T) {
	file, err := newParsedTestFile("key = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	err = block.UnsetKey("key")

	require.ErrorIs(t, err, ErrBlockNotFound)
}

func TestUnsetKey_WhenKeyIsMissing_ThenReturnsKeyNotFound(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	err = block.UnsetKey("other")

	require.ErrorIs(t, err, ErrKeyNotFound)
}

func TestUnsetKey_WhenKeyExists_ThenRemovesLine(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n[other]\nvalue = 2\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	err = block.UnsetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "[section]\n[other]\nvalue = 2\n", string(file.buffer))
}

func TestUnsetKey_WhenKeyHasInlineComment_ThenRemovesLine(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1 # note\nother = 2\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	err = block.UnsetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "[section]\nother = 2\n", string(file.buffer))
}

func TestUnsetKey_WhenFileUsesCRLF_ThenPreservesCRLF(t *testing.T) {
	file, err := newParsedTestFile("[section]\r\nkey = 1\r\nother = 2\r\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	err = block.UnsetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "[section]\r\nother = 2\r\n", string(file.buffer))
}

func TestGetKey_WhenBlockIsNil_ThenReturnsBlockNotFound(t *testing.T) {
	var block *Block

	value, err := block.GetKey("key")

	require.ErrorIs(t, err, ErrBlockNotFound)
	assert.Nil(t, value)
}

func TestGetKey_WhenFileIsNil_ThenReturnsBlockNotFound(t *testing.T) {
	block := &Block{blockName: "section"}

	value, err := block.GetKey("key")

	require.ErrorIs(t, err, ErrBlockNotFound)
	assert.Nil(t, value)
}

func TestGetKey_WhenKeyIsBlank_ThenReturnsKeyNotFound(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{name: "Empty", key: ""},
		{name: "Whitespace", key: "  "},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := newParsedTestFile("[section]\nkey = 1\n")
			require.NoError(t, err)
			block := file.OpenBlock("section")

			value, err := block.GetKey(testCase.key)

			require.ErrorIs(t, err, ErrKeyNotFound)
			assert.Nil(t, value)
		})
	}
}

func TestGetKey_WhenBlockIsMissing_ThenReturnsBlockNotFound(t *testing.T) {
	file, err := newParsedTestFile("key = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.ErrorIs(t, err, ErrBlockNotFound)
	assert.Nil(t, value)
}

func TestGetKey_WhenKeyIsMissing_ThenReturnsKeyNotFound(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("other")

	require.ErrorIs(t, err, ErrKeyNotFound)
	assert.Nil(t, value)
}

func TestGetKey_WhenValueIsDoubleQuotedString_ThenReturnsString(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = \"demo\"\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "demo", value)
}

func TestGetKey_WhenValueIsSingleQuotedString_ThenReturnsString(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 'demo'\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "demo", value)
}

func TestGetKey_WhenValueIsBool_ThenReturnsBool(t *testing.T) {
	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "True", value: "true", expected: true},
		{name: "False", value: "false", expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file, err := newParsedTestFile("[section]\nkey = " + testCase.value + "\n")
			require.NoError(t, err)
			block := file.OpenBlock("section")

			value, err := block.GetKey("key")

			require.NoError(t, err)
			assert.Equal(t, testCase.expected, value)
		})
	}
}

func TestGetKey_WhenValueIsInt_ThenReturnsInt(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 42\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, int64(42), value)
}

func TestGetKey_WhenValueIsFloat_ThenReturnsFloat(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 3.14\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.InEpsilon(t, 3.14, value, 0.0001)
}

func TestGetKey_WhenValueIsUnparseable_ThenReturnsRawString(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = 1.2.3\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "1.2.3", value)
}

func TestGetKey_WhenKeyHasInlineComment_ThenParsesValue(t *testing.T) {
	file, err := newParsedTestFile("[section]\nkey = \"demo\" # note\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, "demo", value)
}

func TestGetKey_WhenKeyHasExtraSpaces_ThenParsesValue(t *testing.T) {
	file, err := newParsedTestFile("[section]\n  key  =  7  \n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, int64(7), value)
}

func TestGetKey_WhenFileUsesCRLF_ThenParsesValue(t *testing.T) {
	file, err := newParsedTestFile("[section]\r\nkey = 1\r\n")
	require.NoError(t, err)
	block := file.OpenBlock("section")

	value, err := block.GetKey("key")

	require.NoError(t, err)
	assert.Equal(t, int64(1), value)
}
