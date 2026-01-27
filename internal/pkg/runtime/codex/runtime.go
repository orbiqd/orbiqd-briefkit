package codex

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/neongreen/mono/lib/toml"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
	"github.com/spf13/afero"
)

var semverPattern = regexp.MustCompile(`\d+\.\d+\.\d+`)

const Codex = agent.RuntimeKind("codex")
const defaultCodexConfigPath = "~/.codex/config.toml"
const codexDefaultToolTimeout = 10 * time.Minute

// RuntimeConfig defines runtime options for Codex execution.
type RuntimeConfig struct {
	// RequireWorkspaceRepository enforces that codex workdir must be a GIT repository.
	RequireWorkspaceRepository bool `json:"requireWorkspaceRepository" default:"true"`
}

type Runtime struct {
}

func NewRuntime() *Runtime {
	return &Runtime{}
}

func (runtime *Runtime) Execute(ctx context.Context, executionId agent.ExecutionID, executionInput agent.ExecutionInput, agentConfig agent.Config) (agent.RuntimeInstance, error) {
	logDir, err := cli.ResolveRuntimeLogDir()
	if err != nil {
		return nil, err
	}

	runtimeConfig, err := utils.AnyToStruct[RuntimeConfig](agentConfig.Runtime.Config)
	if err != nil {
		return nil, fmt.Errorf("convert runtime config: %w", err)
	}

	instance, err := newInstance(ctx, executionId, executionInput, *runtimeConfig, agentConfig.Runtime.Feature, logDir)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (runtime *Runtime) Discovery(ctx context.Context) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	_, err := utils.LookupExecutable(ctx, []string{"codex"})
	if err == nil {
		return true, nil
	}

	if errors.Is(err, exec.ErrNotFound) {
		return false, nil
	}

	return false, err
}

func (runtime *Runtime) GetDefaultConfig(ctx context.Context) (agent.RuntimeConfig, error) {
	return RuntimeConfig{
		RequireWorkspaceRepository: false,
	}, nil
}

func (runtime *Runtime) GetDefaultFeatures(ctx context.Context) (agent.RuntimeFeatures, error) {
	return agent.RuntimeFeatures{}, nil
}

func (runtime *Runtime) GetInfo(ctx context.Context) (agent.RuntimeInfo, error) {
	if err := ctx.Err(); err != nil {
		return agent.RuntimeInfo{}, err
	}

	path, err := utils.LookupExecutable(ctx, []string{"codex"})
	if err != nil {
		return agent.RuntimeInfo{}, fmt.Errorf("lookup codex executable: %w", err)
	}

	// #nosec G204 - path comes from LookupExecutable with hardcoded name
	output, err := exec.CommandContext(ctx, path, "--version").CombinedOutput()
	if err != nil {
		return agent.RuntimeInfo{}, fmt.Errorf("read codex version: %w", err)
	}

	version := semverPattern.FindString(string(output))
	if version == "" {
		return agent.RuntimeInfo{}, fmt.Errorf("parse codex version from output: %s", strings.TrimSpace(string(output)))
	}

	return agent.RuntimeInfo{Version: version}, nil
}

func (runtime *Runtime) RegisterMCPServer(ctx context.Context, serverName agent.RuntimeMCPServerName, server agent.RuntimeMCPServer) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	name := strings.TrimSpace(string(serverName))
	if name == "" {
		return errors.New("missing mcp server name")
	}

	if server.STDIO == nil {
		return errors.New("missing mcp server stdio configuration")
	}

	command := strings.TrimSpace(server.STDIO.Command)
	if command == "" {
		return errors.New("missing mcp server stdio command")
	}

	path, err := utils.LookupExecutable(ctx, []string{"codex"})
	if err != nil {
		return err
	}

	if err := runtime.removeMcpServer(ctx, path, name); err != nil {
		return err
	}

	if err := runtime.addMcpServer(ctx, path, name, command, server.STDIO.Arguments); err != nil {
		return err
	}

	fs := afero.NewOsFs()
	doc, configPath, mode, err := runtime.openConfig(ctx, fs)
	if err != nil {
		return err
	}

	if err := runtime.setMcpServerTimeout(ctx, fs, name, codexDefaultToolTimeout, doc, configPath, mode); err != nil {
		return err
	}

	return nil
}

func (runtime *Runtime) removeMcpServer(ctx context.Context, codexPath string, name string) error {
	// Remove existing server first (for consistency with Claude runtime)
	// #nosec G204 - path comes from LookupExecutable with hardcoded name
	removeOutput, removeErr := exec.CommandContext(ctx, codexPath, "mcp", "remove", name).CombinedOutput()
	if removeErr != nil {
		outputStr := strings.ToLower(strings.TrimSpace(string(removeOutput)))
		if !strings.Contains(outputStr, "no mcp server named") {
			return fmt.Errorf("codex mcp server removal: %s", strings.TrimSpace(string(removeOutput)))
		}
	}

	return nil
}

func (runtime *Runtime) addMcpServer(ctx context.Context, codexPath string, name string, command string, arguments []string) error {
	// Add the server (note: -- separator required before command)
	addArgs := []string{"mcp", "add", name, "--"}
	addArgs = append(addArgs, command)
	addArgs = append(addArgs, arguments...)

	// #nosec G204 - path comes from LookupExecutable with hardcoded name
	addOutput, addErr := exec.CommandContext(ctx, codexPath, addArgs...).CombinedOutput()
	if addErr != nil {
		trimmedOutput := strings.TrimSpace(string(addOutput))
		if trimmedOutput == "" {
			return fmt.Errorf("codex mcp server registration: %w", addErr)
		}
		return fmt.Errorf("codex mcp server registration: %s", trimmedOutput)
	}

	return nil
}

func (runtime *Runtime) openConfig(ctx context.Context, fs afero.Fs) (*toml.Document, string, os.FileMode, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", 0, err
	}

	configPath, err := homedir.Expand(defaultCodexConfigPath)
	if err != nil {
		return nil, "", 0, fmt.Errorf("codex config path expansion: %w", err)
	}

	configDir := filepath.Dir(configPath)
	if err := fs.MkdirAll(configDir, 0o755); err != nil {
		return nil, "", 0, fmt.Errorf("codex config directory creation: %w", err)
	}

	info, err := fs.Stat(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, "", 0, fmt.Errorf("codex config stat: %w", err)
	}

	var (
		doc  *toml.Document
		mode os.FileMode = 0o600
	)

	if err == nil {
		mode = info.Mode()
		contents, readErr := afero.ReadFile(fs, configPath)
		if readErr != nil {
			return nil, "", 0, fmt.Errorf("codex config read: %w", readErr)
		}

		doc, err = toml.Parse(contents)
		if err != nil {
			return nil, "", 0, fmt.Errorf("codex config parse: %w", err)
		}
	} else {
		doc, err = toml.ParseString("")
		if err != nil {
			return nil, "", 0, fmt.Errorf("codex config parse: %w", err)
		}
	}

	return doc, configPath, mode, nil
}

func (runtime *Runtime) setMcpServerTimeout(ctx context.Context, fs afero.Fs, name string, timeout time.Duration, doc *toml.Document, configPath string, mode os.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
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
