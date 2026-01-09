package claude

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func TestClaudeConfig_RoundTrip_PreservesUnknownFields(t *testing.T) {
	fs := afero.NewMemMapFs()
	tmpDir := "/tmp/test"
	err := fs.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "claude-test.json")
	t.Setenv(envConfigPath, configPath)

	originalJSON := `{
  "mcpServers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "env": {"NODE_ENV": "production"}
    }
  },
  "theme": "dark",
  "font_size": 14,
  "telemetry": {"enabled": false}
}`

	// Write original
	err = afero.WriteFile(fs, configPath, []byte(originalJSON), 0644)
	require.NoError(t, err)

	// Read
	data, err := readClaudeConfig(fs)
	require.NoError(t, err)

	// Verify filesystem server exists using gjson
	assert.True(t, gjson.GetBytes(data, "mcpServers.filesystem").Exists())
	assert.Equal(t, "stdio", gjson.GetBytes(data, "mcpServers.filesystem.type").String())
	assert.Equal(t, "npx", gjson.GetBytes(data, "mcpServers.filesystem.command").String())

	// Modify: add new server using sjson
	data, err = sjson.SetBytes(data, "mcpServers.new-server.type", "stdio")
	require.NoError(t, err)
	data, err = sjson.SetBytes(data, "mcpServers.new-server.command", "python")
	require.NoError(t, err)

	// Write back
	err = writeClaudeConfig(fs, data)
	require.NoError(t, err)

	// Verify preservation
	outputBytes, err := afero.ReadFile(fs, configPath)
	require.NoError(t, err)

	// Check unknown fields survived
	assert.Equal(t, "dark", gjson.GetBytes(outputBytes, "theme").String())
	assert.Equal(t, int64(14), gjson.GetBytes(outputBytes, "font_size").Int())
	assert.False(t, gjson.GetBytes(outputBytes, "telemetry.enabled").Bool())

	// Check new server present
	assert.True(t, gjson.GetBytes(outputBytes, "mcpServers.new-server").Exists())
}

func TestReadWriteConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	tmpDir := "/tmp/test"
	err := fs.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "claude-test.json")
	t.Setenv(envConfigPath, configPath)

	t.Run("read non-existent file returns empty JSON", func(t *testing.T) {
		data, err := readClaudeConfig(fs)
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "{}", string(data))
	})

	t.Run("write and read back", func(t *testing.T) {
		// Create test data with sjson
		data := []byte("{}")
		data, err = sjson.SetBytes(data, "mcpServers.test-server.type", "stdio")
		require.NoError(t, err)
		data, err = sjson.SetBytes(data, "mcpServers.test-server.command", "echo")
		require.NoError(t, err)
		data, err = sjson.SetBytes(data, "custom", "value")
		require.NoError(t, err)

		err = writeClaudeConfig(fs, data)
		require.NoError(t, err)

		// Verify file exists
		exists, err := afero.Exists(fs, configPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Read back
		readData, err := readClaudeConfig(fs)
		require.NoError(t, err)
		assert.True(t, gjson.GetBytes(readData, "mcpServers.test-server").Exists())
		assert.Equal(t, "value", gjson.GetBytes(readData, "custom").String())
	})

	t.Run("read valid file directly", func(t *testing.T) {
		content, err := afero.ReadFile(fs, configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "test-server")
	})

	t.Run("write error handling with readonly filesystem", func(t *testing.T) {
		readonlyFs := afero.NewReadOnlyFs(fs)

		testData := []byte(`{"mcpServers":{"test":{"type":"stdio","command":"test"}}}`)
		err := writeClaudeConfig(readonlyFs, testData)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "operation not permitted")
	})
}

func TestReadClaudeConfig_RealWorldScenarios(t *testing.T) {
	testCases := []struct {
		name           string
		filename       string
		expectError    bool
		errorContains  string
		validateConfig func(t *testing.T, data []byte)
	}{
		{
			name:        "valid stdio only",
			filename:    ".claude.valid-stdio-only.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				servers := gjson.GetBytes(data, "mcpServers")
				assert.True(t, servers.Exists())

				// Count servers
				count := 0
				servers.ForEach(func(_, _ gjson.Result) bool {
					count++
					return true
				})
				require.Equal(t, 3, count)

				assert.True(t, gjson.GetBytes(data, "mcpServers.filesystem").Exists())
				assert.True(t, gjson.GetBytes(data, "mcpServers.github").Exists())
				assert.True(t, gjson.GetBytes(data, "mcpServers.postgres").Exists())

				assert.Equal(t, "stdio", gjson.GetBytes(data, "mcpServers.filesystem.type").String())
				assert.Equal(t, "npx", gjson.GetBytes(data, "mcpServers.filesystem.command").String())
				assert.Equal(t, "production", gjson.GetBytes(data, "mcpServers.filesystem.env.NODE_ENV").String())
			},
		},
		{
			name:        "mixed stdio and sse",
			filename:    ".claude.mixed-stdio-and-sse.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				assert.True(t, gjson.GetBytes(data, "mcpServers.filesystem").Exists())
				assert.True(t, gjson.GetBytes(data, "mcpServers.remote-api").Exists())
				assert.True(t, gjson.GetBytes(data, "mcpServers.local-python").Exists())

				assert.Equal(t, "stdio", gjson.GetBytes(data, "mcpServers.filesystem.type").String())
				assert.Equal(t, "sse", gjson.GetBytes(data, "mcpServers.remote-api.type").String())
				assert.Equal(t, "stdio", gjson.GetBytes(data, "mcpServers.local-python.type").String())
			},
		},
		{
			name:        "mixed stdio and websocket",
			filename:    ".claude.mixed-stdio-and-websocket.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				assert.Equal(t, "stdio", gjson.GetBytes(data, "mcpServers.github.type").String())
				assert.Equal(t, "websocket", gjson.GetBytes(data, "mcpServers.websocket-service.type").String())
			},
		},
		{
			name:        "all non-stdio",
			filename:    ".claude.all-non-stdio.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				assert.Equal(t, "sse", gjson.GetBytes(data, "mcpServers.sse-server-1.type").String())
				assert.Equal(t, "sse", gjson.GetBytes(data, "mcpServers.sse-server-2.type").String())
				assert.Equal(t, "websocket", gjson.GetBytes(data, "mcpServers.websocket-server.type").String())
			},
		},
		{
			name:        "empty servers",
			filename:    ".claude.empty-servers.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				servers := gjson.GetBytes(data, "mcpServers")
				if servers.Exists() {
					// Count should be 0
					count := 0
					servers.ForEach(func(_, _ gjson.Result) bool {
						count++
						return true
					})
					assert.Equal(t, 0, count)
				}
			},
		},
		{
			name:        "with other fields",
			filename:    ".claude.with-other-fields.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				assert.True(t, gjson.GetBytes(data, "mcpServers.filesystem").Exists())

				// Check other fields exist
				assert.True(t, gjson.GetBytes(data, "theme").Exists())
				assert.True(t, gjson.GetBytes(data, "font_size").Exists())
				assert.True(t, gjson.GetBytes(data, "editor").Exists())
				assert.True(t, gjson.GetBytes(data, "telemetry").Exists())
				assert.True(t, gjson.GetBytes(data, "customSettings").Exists())
			},
		},
		{
			name:        "complex stdio",
			filename:    ".claude.complex-stdio.json",
			expectError: false,
			validateConfig: func(t *testing.T, data []byte) {
				t.Helper()
				assert.Equal(t, "stdio", gjson.GetBytes(data, "mcpServers.complex-server.type").String())
				assert.Equal(t, "/usr/local/bin/custom-mcp-server", gjson.GetBytes(data, "mcpServers.complex-server.command").String())

				args := gjson.GetBytes(data, "mcpServers.complex-server.args")
				assert.True(t, args.IsArray())
				assert.Len(t, args.Array(), 10)

				assert.Equal(t, "production", gjson.GetBytes(data, "mcpServers.complex-server.env.NODE_ENV").String())
				assert.Equal(t, "sk-1234567890abcdef", gjson.GetBytes(data, "mcpServers.complex-server.env.API_KEY").String())
			},
		},
		{
			name:          "invalid syntax",
			filename:      ".claude.invalid-syntax.json",
			expectError:   true,
			errorContains: "unmarshal config",
		},
	}

	fs := afero.NewOsFs()

	testDataDir, err := filepath.Abs("../../../../test/runtime/claude")
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := filepath.Join(testDataDir, tc.filename)
			t.Setenv(envConfigPath, configPath)

			config, err := readClaudeConfig(fs)

			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tc.validateConfig != nil {
					tc.validateConfig(t, config)
				}
			}
		})
	}
}

func TestClaudeConfig_PreservesFormatting(t *testing.T) {
	fs := afero.NewMemMapFs()
	tmpDir := "/tmp/test"
	err := fs.MkdirAll(tmpDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(tmpDir, "claude-test.json")
	t.Setenv(envConfigPath, configPath)

	// Original JSON with specific formatting and key order
	originalJSON := `{
    "customField": "value",
    "mcpServers": {
        "existing": {
            "type": "stdio",
            "command": "node"
        }
    },
    "anotherField": 123
}`

	err = afero.WriteFile(fs, configPath, []byte(originalJSON), 0644)
	require.NoError(t, err)

	// Read config
	data, err := readClaudeConfig(fs)
	require.NoError(t, err)

	// Add a new server using sjson
	data, err = sjson.SetBytes(data, "mcpServers.new-server.type", "stdio")
	require.NoError(t, err)
	data, err = sjson.SetBytes(data, "mcpServers.new-server.command", "python")
	require.NoError(t, err)

	// Write back
	err = writeClaudeConfig(fs, data)
	require.NoError(t, err)

	// Verify custom fields preserved
	outputBytes, err := afero.ReadFile(fs, configPath)
	require.NoError(t, err)

	output := string(outputBytes)

	// Verify custom fields unchanged
	assert.Contains(t, output, `"customField"`)
	assert.Contains(t, output, `"value"`)
	assert.Contains(t, output, `"anotherField"`)
	assert.Contains(t, output, `123`)

	// Verify new server added
	assert.Contains(t, output, `"new-server"`)
	assert.True(t, gjson.Get(output, "mcpServers.new-server").Exists())
	assert.Equal(t, "stdio", gjson.Get(output, "mcpServers.new-server.type").String())
	assert.Equal(t, "python", gjson.Get(output, "mcpServers.new-server.command").String())

	// Verify existing server unchanged
	assert.True(t, gjson.Get(output, "mcpServers.existing").Exists())
	assert.Equal(t, "node", gjson.Get(output, "mcpServers.existing.command").String())
}
