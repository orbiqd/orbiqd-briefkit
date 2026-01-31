package fs

import (
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

func TestWriteJSON_ValidData_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.json"
	data := testData{Name: "test", Value: 42}

	err := writeJSON(fs, filePath, data)

	require.NoError(t, err)
	exists, err := afero.Exists(fs, filePath)
	require.NoError(t, err)
	assert.True(t, exists)
	content, err := afero.ReadFile(fs, filePath)
	require.NoError(t, err)
	assert.JSONEq(t, `{"name":"test","value":42}`, string(content))
}

func TestWriteJSON_TempFileExists_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.json"
	tmpPath := filePath + "~"
	data := testData{Name: "test", Value: 42}

	err := afero.WriteFile(fs, tmpPath, []byte("existing"), 0644)
	require.NoError(t, err)

	err = writeJSON(fs, filePath, data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "temp file")
	assert.Contains(t, err.Error(), "already exists")
}

func TestWriteJSON_UnmarshalableData_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.json"
	data := make(chan int) // channels cannot be marshaled to JSON

	err := writeJSON(fs, filePath, data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal")
}

func TestWriteYAML_ValidData_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.yaml"
	data := testData{Name: "test", Value: 42}

	err := writeYAML(fs, filePath, data)

	require.NoError(t, err)
	exists, err := afero.Exists(fs, filePath)
	require.NoError(t, err)
	assert.True(t, exists)
	content, err := afero.ReadFile(fs, filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: test")
	assert.Contains(t, string(content), "value: 42")
}

func TestWriteYAML_TempFileExists_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.yaml"
	tmpPath := filePath + "~"
	data := testData{Name: "test", Value: 42}

	err := afero.WriteFile(fs, tmpPath, []byte("existing"), 0644)
	require.NoError(t, err)

	err = writeYAML(fs, filePath, data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "temp file")
	assert.Contains(t, err.Error(), "already exists")
}

func TestReadJSON_ValidFile_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.json"
	content := `{"name":"test","value":42}`

	err := afero.WriteFile(fs, filePath, []byte(content), 0644)
	require.NoError(t, err)

	result, err := readJSON[testData](fs, filePath)

	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
}

func TestReadJSON_FileNotExists_ReturnsErrExecutionNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/nonexistent.json"

	_, err := readJSON[testData](fs, filePath)

	require.Error(t, err)
	assert.ErrorIs(t, err, agent.ErrExecutionNotFound)
}

func TestReadJSON_InvalidJSON_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/invalid.json"
	content := `{invalid json}`

	err := afero.WriteFile(fs, filePath, []byte(content), 0644)
	require.NoError(t, err)

	_, err = readJSON[testData](fs, filePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestReadYAML_ValidFile_Success(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.yaml"
	content := "name: test\nvalue: 42\n"

	err := afero.WriteFile(fs, filePath, []byte(content), 0644)
	require.NoError(t, err)

	result, err := readYAML[testData](fs, filePath)

	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
}

func TestReadYAML_FileNotExists_ReturnsErrAgentConfigNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/nonexistent.yaml"

	_, err := readYAML[testData](fs, filePath)

	require.Error(t, err)
	assert.ErrorIs(t, err, agent.ErrAgentConfigNotFound)
}

func TestReadYAML_InvalidYAML_ReturnsError(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/invalid.yaml"
	content := "name: test\n  invalid: indentation"

	err := afero.WriteFile(fs, filePath, []byte(content), 0644)
	require.NoError(t, err)

	_, err = readYAML[testData](fs, filePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestHasJSON_FileExists_ReturnsTrue(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/test.json"

	err := afero.WriteFile(fs, filePath, []byte(`{}`), 0644)
	require.NoError(t, err)

	exists, err := hasJSON(fs, filePath)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestHasJSON_FileNotExists_ReturnsFalse(t *testing.T) {
	fs := afero.NewMemMapFs()
	filePath := "/tmp/nonexistent.json"

	exists, err := hasJSON(fs, filePath)

	require.NoError(t, err)
	assert.False(t, exists)
}
