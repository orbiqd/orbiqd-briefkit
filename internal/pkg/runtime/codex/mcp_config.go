package codex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/neongreen/mono/lib/toml"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/spf13/afero"
)

const defaultCodexConfigPath = "~/.codex/config.toml"

// EnsureMCPToolTimeout updates Codex MCP server configuration with tool timeout in seconds.
func EnsureMCPToolTimeout(ctx context.Context, fs afero.Fs, serverName agent.RuntimeMCPServerName, timeout time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	name := strings.TrimSpace(string(serverName))
	if name == "" {
		return errors.New("missing mcp server name")
	}

	if timeout <= 0 {
		return errors.New("invalid tool timeout")
	}

	configPath, err := homedir.Expand(defaultCodexConfigPath)
	if err != nil {
		return fmt.Errorf("codex config path expansion: %w", err)
	}

	configDir := filepath.Dir(configPath)
	if err := fs.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("codex config directory creation: %w", err)
	}

	info, err := fs.Stat(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("codex config stat: %w", err)
	}

	var (
		doc  *toml.Document
		mode os.FileMode = 0o600
	)

	if err == nil {
		mode = info.Mode()
		contents, readErr := afero.ReadFile(fs, configPath)
		if readErr != nil {
			return fmt.Errorf("codex config read: %w", readErr)
		}

		doc, err = toml.Parse(contents)
		if err != nil {
			return fmt.Errorf("codex config parse: %w", err)
		}
	} else {
		doc, err = toml.ParseString("")
		if err != nil {
			return fmt.Errorf("codex config parse: %w", err)
		}
	}

	timeoutSec := int64(timeout.Seconds())
	keyPath := fmt.Sprintf("mcp_servers.%s.tool_timeout_sec", name)
	if err := doc.Set(keyPath, timeoutSec); err != nil {
		return fmt.Errorf("codex config update: %w", err)
	}

	if err := afero.WriteFile(fs, configPath, []byte(doc.String()), mode); err != nil {
		return fmt.Errorf("codex config write: %w", err)
	}

	return nil
}
