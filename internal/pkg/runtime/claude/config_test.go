package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaudeConfig_RoundTrip_PreservesUnknownFields(t *testing.T) {
	// Simulate a configuration file with a mix of known fields (mcpServers)
	// and unknown user fields ("theme", "font_size", "telemetry")
	// which must be preserved during round-trip.
	originalJSON := `{
  "mcpServers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-filesystem",
        "/Users/user/code"
      ],
      "env": {
        "NODE_ENV": "production"
      }
    }
  },
  "theme": "dark",
  "font_size": 14,
  "telemetry": {
    "enabled": false
  }
}`

	// 1. Unmarshal (Deserialization)
	var config ClaudeConfig
	err := json.Unmarshal([]byte(originalJSON), &config)
	require.NoError(t, err, "Deserialization should succeed")

	// Verify: Known fields are correctly mapped
	require.Contains(t, config.MCPServers, "filesystem")
	fsServer := config.MCPServers["filesystem"]
	assert.Equal(t, "stdio", fsServer.Type)
	assert.Equal(t, "npx", fsServer.Command)
	assert.Len(t, fsServer.Args, 3)
	assert.Equal(t, "production", fsServer.Env["NODE_ENV"])

	// Verify: Unknown fields are captured in OtherFields
	assert.Contains(t, config.OtherFields, "theme")
	assert.Contains(t, config.OtherFields, "font_size")
	assert.Contains(t, config.OtherFields, "telemetry")

	// 2. Modification (Simulate adding a server by our application)
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]ClaudeMCPServerConfig)
	}
	config.MCPServers["new-server"] = ClaudeMCPServerConfig{
		Type:    "stdio",
		Command: "python",
	}

	// 3. Marshal (Serialization)
	outputBytes, err := json.MarshalIndent(&config, "", "  ")
	require.NoError(t, err, "Serialization should succeed")

	// 4. Final Verification
	// Decode back to a map to verify everything is present
	var outputMap map[string]interface{}
	err = json.Unmarshal(outputBytes, &outputMap)
	require.NoError(t, err)

	// Verify: Unknown fields survived the round-trip
	assert.Equal(t, "dark", outputMap["theme"])
	assert.Equal(t, float64(14), outputMap["font_size"]) // JSON numbers are float64 in interface{}

	telemetry, ok := outputMap["telemetry"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, telemetry["enabled"])

	// Verify: Our modifications are present
	servers, ok := outputMap["mcpServers"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, servers, "filesystem")
	assert.Contains(t, servers, "new-server")
}

func TestClaudeConfig_PartialUnmarshal(t *testing.T) {
	t.Run("only unknown fields", func(t *testing.T) {
		jsonStr := `{"custom_setting": "value"}`
		var config ClaudeConfig
		err := json.Unmarshal([]byte(jsonStr), &config)
		require.NoError(t, err)

		assert.Empty(t, config.MCPServers)
		assert.Contains(t, config.OtherFields, "custom_setting")
	})

	t.Run("only known fields", func(t *testing.T) {
		jsonStr := `{"mcpServers": {"s1": {"type": "stdio"}}}`
		var config ClaudeConfig
		err := json.Unmarshal([]byte(jsonStr), &config)
		require.NoError(t, err)

		assert.Contains(t, config.MCPServers, "s1")
		// mcpServers should not be duplicated in OtherFields
		_, hasMcpInOther := config.OtherFields["mcpServers"]
		assert.False(t, hasMcpInOther, "mcpServers should not be duplicated in OtherFields")
	})
}

func TestConverters(t *testing.T) {
	t.Run("toRuntimeMCPServer", func(t *testing.T) {
		t.Run("success stdio", func(t *testing.T) {
			input := ClaudeMCPServerConfig{
				Type:    "stdio",
				Command: "test-cmd",
				Args:    []string{"arg1", "arg2"},
			}
			result, err := toRuntimeMCPServer(input)
			require.NoError(t, err)
			require.NotNil(t, result.STDIO)
			assert.Equal(t, "test-cmd", result.STDIO.Command)
			assert.Equal(t, []string{"arg1", "arg2"}, result.STDIO.Arguments)
		})

		t.Run("unsupported type", func(t *testing.T) {
			input := ClaudeMCPServerConfig{
				Type: "websocket",
			}
			_, err := toRuntimeMCPServer(input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported server type")
		})
	})

	t.Run("toClaudeMCPServerConfig", func(t *testing.T) {
		t.Run("success stdio", func(t *testing.T) {
			input := agent.RuntimeMCPServer{
				STDIO: &agent.RuntimeSTDIOMCPServer{
					Command:   "test-cmd",
					Arguments: []string{"arg1"},
				},
			}
			result, err := toClaudeMCPServerConfig(input)
			require.NoError(t, err)
			assert.Equal(t, "stdio", result.Type)
			assert.Equal(t, "test-cmd", result.Command)
			assert.Equal(t, []string{"arg1"}, result.Args)
		})

		t.Run("missing stdio", func(t *testing.T) {
			input := agent.RuntimeMCPServer{
				STDIO: nil,
			}
			_, err := toClaudeMCPServerConfig(input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "only STDIO servers are supported")
		})
	})
}

func TestReadWriteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "claude-test.json")

	// Set env var to point to our temp file
	t.Setenv(envConfigPath, configPath)

	t.Run("read non-existent file returns empty config", func(t *testing.T) {
		config, err := readClaudeConfig()
		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Empty(t, config.MCPServers)
		assert.Empty(t, config.OtherFields)
	})

	t.Run("write and read back", func(t *testing.T) {
		config := &ClaudeConfig{
			MCPServers: map[string]ClaudeMCPServerConfig{
				"test-server": {
					Type:    "stdio",
					Command: "echo",
				},
			},
			OtherFields: map[string]json.RawMessage{
				"custom": json.RawMessage(`"value"`),
			},
		}

		err := writeClaudeConfig(config)
		require.NoError(t, err)

		// Verify file exists
		exists, err := fileExists(configPath)
		require.NoError(t, err)
		assert.True(t, exists)

		// Read back
		readConfig, err := readClaudeConfig()
		require.NoError(t, err)
		assert.Contains(t, readConfig.MCPServers, "test-server")
		assert.Contains(t, readConfig.OtherFields, "custom")
	})

	t.Run("read valid file directly", func(t *testing.T) {
		// Verify content on disk
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "test-server")
	})

	t.Run("write error handling", func(t *testing.T) {
		// Force error by setting path to a directory
		badPath := filepath.Join(tmpDir, "bad-dir")
		err := os.Mkdir(badPath, 0755)
		require.NoError(t, err)
		t.Setenv(envConfigPath, badPath)

		config := &ClaudeConfig{}
		err = writeClaudeConfig(config)
		assert.Error(t, err)
	})
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	
	t.Run("file exists", func(t *testing.T) {
		path := filepath.Join(tmpDir, "exists.txt")
		err := os.WriteFile(path, []byte("content"), 0644)
		require.NoError(t, err)

		exists, err := fileExists(path)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("file does not exist", func(t *testing.T) {
		path := filepath.Join(tmpDir, "missing.txt")
		exists, err := fileExists(path)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}