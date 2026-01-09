package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/afero"
)

var ErrMCPServerNotFound = errors.New("mcp server not found")

// readClaudeConfig reads raw ~/.claude.json bytes
// Returns empty JSON object (without error) if file doesn't exist
func readClaudeConfig(fs afero.Fs) ([]byte, error) {
	configPath, err := locateConfigPath()
	if err != nil {
		return nil, fmt.Errorf("locate config path: %w", err)
	}

	data, err := afero.ReadFile(fs, configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte("{}"), nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if len(data) == 0 {
		return []byte("{}"), nil
	}

	// Validate JSON syntax
	if !json.Valid(data) {
		return nil, errors.New("unmarshal config: invalid JSON syntax")
	}

	return data, nil
}

// writeClaudeConfig writes raw JSON bytes to ~/.claude.json atomically
// Uses pattern: write to temp file → rename (atomic operation)
func writeClaudeConfig(fs afero.Fs, data []byte) error {
	configPath, err := locateConfigPath()
	if err != nil {
		return fmt.Errorf("locate config path: %w", err)
	}

	tmpPath := configPath + "~"

	exists, err := afero.Exists(fs, tmpPath)
	if err != nil {
		return fmt.Errorf("check temp file existence: %w", err)
	}
	if exists {
		return fmt.Errorf("temp file %s already exists: %w", tmpPath, os.ErrExist)
	}

	if err := afero.WriteFile(fs, tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := fs.Rename(tmpPath, configPath); err != nil {
		_ = fs.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
