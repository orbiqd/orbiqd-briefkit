package claude

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
)

var ErrMCPServerNotFound = fmt.Errorf("mcp server not found")

// ClaudeConfig represents the full configuration file ~/.claude.json
// It preserves unknown fields to avoid overwriting user's custom Claude CLI settings
type ClaudeConfig struct {
	MCPServers  map[string]ClaudeMCPServerConfig `json:"mcpServers"`
	OtherFields map[string]json.RawMessage       `json:"-"`
}

// UnmarshalJSON custom deserializer that preserves unknown fields
func (config *ClaudeConfig) UnmarshalJSON(data []byte) error {
	type Alias ClaudeConfig
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(config),
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	config.OtherFields = make(map[string]json.RawMessage)
	for key, value := range raw {
		if key != "mcpServers" {
			config.OtherFields[key] = value
		}
	}

	return nil
}

// MarshalJSON custom serializer that includes other preserved fields
func (config *ClaudeConfig) MarshalJSON() ([]byte, error) {
	result := make(map[string]any)

	for key, value := range config.OtherFields {
		result[key] = value
	}

	if config.MCPServers != nil {
		result["mcpServers"] = config.MCPServers
	}

	return json.MarshalIndent(result, "", "  ")
}

// ClaudeMCPServerConfig represents a single MCP server in Claude CLI format
type ClaudeMCPServerConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// readClaudeConfig reads the full Claude CLI configuration from ~/.claude.json
// Returns an empty configuration (without error) if the file does not exist
func readClaudeConfig() (*ClaudeConfig, error) {
	configPath, err := locateConfigPath()
	if err != nil {
		return nil, fmt.Errorf("locate config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ClaudeConfig{
				MCPServers:  make(map[string]ClaudeMCPServerConfig),
				OtherFields: make(map[string]json.RawMessage),
			}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty")
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]ClaudeMCPServerConfig)
	}

	return &config, nil
}

// writeClaudeConfig writes the full Claude CLI configuration to ~/.claude.json atomically
// Uses pattern: write to temp file → rename (atomic operation)
func writeClaudeConfig(config *ClaudeConfig) error {
	configPath, err := locateConfigPath()
	if err != nil {
		return fmt.Errorf("locate config path: %w", err)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	tmpPath := configPath + "~"

	exists, err := fileExists(tmpPath)
	if err != nil {
		return fmt.Errorf("check temp file existence: %w", err)
	}
	if exists {
		return fmt.Errorf("temp file %s already exists: %w", tmpPath, os.ErrExist)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, configPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// toRuntimeMCPServer converts ClaudeMCPServerConfig → agent.RuntimeMCPServer
func toRuntimeMCPServer(claudeServer ClaudeMCPServerConfig) (agent.RuntimeMCPServer, error) {
	if claudeServer.Type != "stdio" {
		return agent.RuntimeMCPServer{}, fmt.Errorf("unsupported server type: %s (only stdio is supported)", claudeServer.Type)
	}

	return agent.RuntimeMCPServer{
		STDIO: &agent.RuntimeSTDIOMCPServer{
			Command:   claudeServer.Command,
			Arguments: claudeServer.Args,
		},
	}, nil
}

// toClaudeMCPServerConfig converts agent.RuntimeMCPServer → ClaudeMCPServerConfig
func toClaudeMCPServerConfig(runtimeServer agent.RuntimeMCPServer) (ClaudeMCPServerConfig, error) {
	if runtimeServer.STDIO == nil {
		return ClaudeMCPServerConfig{}, fmt.Errorf("only STDIO servers are supported")
	}

	return ClaudeMCPServerConfig{
		Type:    "stdio",
		Command: runtimeServer.STDIO.Command,
		Args:    runtimeServer.STDIO.Arguments,
	}, nil
}
