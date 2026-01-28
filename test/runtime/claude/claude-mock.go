package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
)

// CLI structure for Kong parser.
var CLI struct {
	MCP  MCPCmd  `cmd:"" help:"MCP server management."`
	Exec ExecCmd `cmd:"" default:"withargs" help:"Execute prompt (default)."`

	Version VersionFlag `name:"version" help:"Print version information."`
}

// VersionFlag handles --version flag using BeforeApply hook.
type VersionFlag bool

// BeforeApply is called by Kong before applying the flag value.
//
//nolint:unparam // Kong requires BeforeApply() error signature
func (v VersionFlag) BeforeApply(app *kong.Kong) error {
	config := loadMockConfig()

	if config.VersionFail {
		_, _ = fmt.Fprintln(os.Stderr, "version check failed")
		app.Exit(1)
	}

	if config.VersionNoSemver {
		fmt.Println("claude-code development-build")
	} else {
		fmt.Println("claude-code 1.0.0-mock")
	}

	app.Exit(0)
	return nil
}

// ExecCmd is the default command that handles prompt execution.
type ExecCmd struct {
	Print           bool   `help:"Print output." default:"true"`
	Verbose         bool   `help:"Verbose output." default:"true"`
	OutputFormat    string `name:"output-format" help:"Output format." default:"stream-json"`
	Model           string `help:"Model to use."`
	Resume          string `help:"Conversation ID to resume."`
	Settings        string `help:"Settings JSON."`
	PermissionMode  string `name:"permission-mode" help:"Permission mode."`
	DisallowedTools string `name:"disallowed-tools" help:"Disallowed tools."`
}

// MCPCmd handles mcp subcommand.
type MCPCmd struct {
	Add    MCPAddCmd    `cmd:"" help:"Add MCP server."`
	Remove MCPRemoveCmd `cmd:"" help:"Remove MCP server."`
}

// MCPAddCmd handles mcp add subcommand.
type MCPAddCmd struct {
	Scope string   `name:"scope" help:"Scope for the server." default:"user"`
	Name  string   `arg:"" help:"Server name."`
	Cmd   string   `arg:"" help:"Command to run."`
	Args  []string `arg:"" optional:"" help:"Command arguments."`
}

// MCPRemoveCmd handles mcp remove subcommand.
type MCPRemoveCmd struct {
	Scope string `name:"scope" help:"Scope for the server." default:"user"`
	Name  string `arg:"" help:"Server name."`
}

// MockConfig holds all configuration from environment variables.
type MockConfig struct {
	Fail           bool
	ExitCode       int
	Stderr         string
	MalformedJSON  bool
	ResultError    bool
	NoResult       bool
	Signal         string
	PartialFail    bool
	MultiAssistant bool
	EmptyLines     bool

	VersionFail     bool
	VersionNoSemver bool

	MCPNotFound    bool
	MCPRemoveFail  bool
	MCPAddFail     bool
	MCPAddNoOutput bool
}

// loadMockConfig reads all MOCK_CLAUDE_* environment variables.
func loadMockConfig() MockConfig {
	exitCode := 1
	if code := os.Getenv("MOCK_CLAUDE_EXIT_CODE"); code != "" {
		if parsed, err := strconv.Atoi(code); err == nil {
			exitCode = parsed
		}
	}

	return MockConfig{
		Fail:           envBool("MOCK_CLAUDE_FAIL"),
		ExitCode:       exitCode,
		Stderr:         os.Getenv("MOCK_CLAUDE_STDERR"),
		MalformedJSON:  envBool("MOCK_CLAUDE_MALFORMED_JSON"),
		ResultError:    envBool("MOCK_CLAUDE_RESULT_ERROR"),
		NoResult:       envBool("MOCK_CLAUDE_NO_RESULT"),
		Signal:         os.Getenv("MOCK_CLAUDE_SIGNAL"),
		PartialFail:    envBool("MOCK_CLAUDE_PARTIAL_FAIL"),
		MultiAssistant: envBool("MOCK_CLAUDE_MULTI_ASSISTANT"),
		EmptyLines:     envBool("MOCK_CLAUDE_EMPTY_LINES"),

		VersionFail:     envBool("MOCK_CLAUDE_VERSION_FAIL"),
		VersionNoSemver: envBool("MOCK_CLAUDE_VERSION_NO_SEMVER"),

		MCPNotFound:    envBool("MOCK_CLAUDE_MCP_NOT_FOUND"),
		MCPRemoveFail:  envBool("MOCK_CLAUDE_MCP_REMOVE_FAIL"),
		MCPAddFail:     envBool("MOCK_CLAUDE_MCP_ADD_FAIL"),
		MCPAddNoOutput: envBool("MOCK_CLAUDE_MCP_ADD_NO_OUTPUT"),
	}
}

func envBool(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes"
}

// JSON event structures matching instance.go expectations.
type content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type message struct {
	Content []content `json:"content"`
}

type event struct {
	Type      string   `json:"type"`
	Subtype   string   `json:"subtype,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
	Message   *message `json:"message,omitempty"`
	Result    string   `json:"result,omitempty"`
}

// Output step functions for composable JSON stream generation.
func writeInit(w io.Writer, sessionID string) error {
	return json.NewEncoder(w).Encode(event{
		Type:      "system",
		Subtype:   "init",
		SessionID: sessionID,
	})
}

func writeAssistant(w io.Writer, text string) error {
	return json.NewEncoder(w).Encode(event{
		Type: "assistant",
		Message: &message{
			Content: []content{
				{Type: "text", Text: text},
			},
		},
	})
}

func writeResultSuccess(w io.Writer, result string) error {
	return json.NewEncoder(w).Encode(event{
		Type:    "result",
		Subtype: "success",
		Result:  result,
	})
}

func writeResultError(w io.Writer, errMsg string) error {
	return json.NewEncoder(w).Encode(event{
		Type:    "result",
		Subtype: "error",
		Result:  errMsg,
	})
}

func writeMalformedJSON(w io.Writer) error {
	_, err := fmt.Fprintln(w, `{"type":"system","subtype":"init","session_id":`)
	return err
}

func writeEmptyLine(w io.Writer) error {
	_, err := fmt.Fprintln(w)
	return err
}

func writeNonJSONLine(w io.Writer, text string) error {
	_, err := fmt.Fprintln(w, text)
	return err
}

// Run handles prompt execution (default command).
//
//nolint:unparam // Kong requires Run() error signature
func (cmd *ExecCmd) Run() error {
	config := loadMockConfig()

	prompt := readStdin()
	if prompt == "" {
		prompt = "MOCK_RESPONSE"
	}

	sessionID := "mock-session-id-12345"
	if cmd.Resume != "" {
		sessionID = cmd.Resume
	}

	responseText := cmd.buildResponseText(prompt)

	if config.Signal != "" {
		handleSignal(config)
		return nil
	}

	_ = writeNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")

	if config.EmptyLines {
		_ = writeEmptyLine(os.Stdout)
	}

	if config.MalformedJSON {
		_ = writeMalformedJSON(os.Stdout)
		return nil
	}

	_ = writeInit(os.Stdout, sessionID)

	if config.EmptyLines {
		_ = writeEmptyLine(os.Stdout)
	}

	time.Sleep(10 * time.Millisecond)

	if config.PartialFail {
		_ = writeAssistant(os.Stdout, responseText)
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	if config.MultiAssistant {
		_ = writeAssistant(os.Stdout, "First part of response")
		if config.EmptyLines {
			_ = writeEmptyLine(os.Stdout)
		}
		_ = writeAssistant(os.Stdout, "Second part of response")
		if config.EmptyLines {
			_ = writeEmptyLine(os.Stdout)
		}
		_ = writeAssistant(os.Stdout, "Third part of response")
		responseText = "First part of responseSecond part of responseThird part of response"
	} else {
		_ = writeAssistant(os.Stdout, responseText)
	}

	if config.EmptyLines {
		_ = writeEmptyLine(os.Stdout)
	}

	if config.NoResult {
		return nil
	}

	if config.ResultError {
		_ = writeResultError(os.Stdout, "execution error occurred")
		return nil
	}

	_ = writeResultSuccess(os.Stdout, responseText)

	if config.Fail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	return nil
}

func (cmd *ExecCmd) buildResponseText(prompt string) string {
	text := "Mock response to: " + prompt

	if cmd.Model != "" {
		text += fmt.Sprintf(" (Model: %s)", cmd.Model)
	}

	if cmd.Resume != "" {
		text += fmt.Sprintf(" (Resumed: %s)", cmd.Resume)
	}

	if cmd.PermissionMode != "" {
		text += fmt.Sprintf(" (PermissionMode: %s)", cmd.PermissionMode)
	}

	if cmd.Settings != "" {
		text += fmt.Sprintf(" (Settings: %s)", cmd.Settings)
	}

	return text
}

// Run handles mcp add command.
//
//nolint:unparam // Kong requires Run() error signature
func (cmd *MCPAddCmd) Run() error {
	config := loadMockConfig()

	if config.MCPAddFail {
		if config.MCPAddNoOutput {
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stderr, "MCP server registration failed")
		os.Exit(1)
	}

	return nil
}

// Run handles mcp remove command.
//
//nolint:unparam // Kong requires Run() error signature
func (cmd *MCPRemoveCmd) Run() error {
	config := loadMockConfig()

	if config.MCPNotFound {
		_, _ = fmt.Fprintln(os.Stdout, "No MCP server found with name: "+cmd.Name)
		os.Exit(1)
	}

	if config.MCPRemoveFail {
		_, _ = fmt.Fprintln(os.Stderr, "MCP server removal failed")
		os.Exit(1)
	}

	return nil
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("claude"),
		kong.Description("Mock Claude CLI for testing."),
		kong.NoDefaultHelp(),
	)

	err := ctx.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readStdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	var builder strings.Builder
	for scanner.Scan() {
		builder.WriteString(scanner.Text())
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func handleSignal(config MockConfig) {
	_ = writeNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")
	_ = writeInit(os.Stdout, "mock-session-id-12345")

	switch strings.ToUpper(config.Signal) {
	case "SIGINT":
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	case "SIGTERM":
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	default:
		_, _ = fmt.Fprintln(os.Stderr, "unknown signal: "+config.Signal)
		os.Exit(1)
	}
}
