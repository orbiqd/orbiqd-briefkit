package claude

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/alecthomas/kong"
)

const claudeMockEnvKey = "BRIEFKIT_CLAUDE_MOCK"

func init() {
	if os.Getenv(claudeMockEnvKey) != "1" {
		return
	}

	flag.Bool("version", false, "")
	flag.Bool("print", false, "")
	flag.Bool("verbose", false, "")
	flag.String("output-format", "", "")
	flag.String("model", "", "")
	flag.String("resume", "", "")
	flag.String("settings", "", "")
	flag.String("permission-mode", "", "")
	flag.String("disallowed-tools", "", "")
	flag.String("scope", "", "")
}

func TestMain(m *testing.M) {
	if os.Getenv(claudeMockEnvKey) == "1" {
		os.Exit(runClaudeMock())
	}

	os.Exit(m.Run())
}

func runClaudeMock() int {
	ctx := kong.Parse(&claudeMockCLI,
		kong.Name("claude"),
		kong.Description("Mock Claude CLI for testing."),
		kong.NoDefaultHelp(),
	)

	if err := ctx.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

var claudeMockCLI struct {
	MCP  claudeMockMCPCmd  `cmd:"" help:"MCP server management."`
	Exec claudeMockExecCmd `cmd:"" default:"withargs" help:"Execute prompt (default)."`

	Version claudeMockVersionFlag `name:"version" help:"Print version information."`
}

type claudeMockVersionFlag bool

func (v claudeMockVersionFlag) BeforeApply(app *kong.Kong) error {
	config := loadClaudeMockConfig()

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

type claudeMockExecCmd struct {
	Print           bool   `help:"Print output." default:"true"`
	Verbose         bool   `help:"Verbose output." default:"true"`
	OutputFormat    string `name:"output-format" help:"Output format." default:"stream-json"`
	Model           string `help:"Model to use."`
	Resume          string `help:"Conversation ID to resume."`
	Settings        string `help:"Settings JSON."`
	PermissionMode  string `name:"permission-mode" help:"Permission mode."`
	DisallowedTools string `name:"disallowed-tools" help:"Disallowed tools."`
}

type claudeMockMCPCmd struct {
	Add    claudeMockMCPAddCmd    `cmd:"" help:"Add MCP server."`
	Remove claudeMockMCPRemoveCmd `cmd:"" help:"Remove MCP server."`
}

type claudeMockMCPAddCmd struct {
	Scope string   `name:"scope" help:"Scope for the server." default:"user"`
	Name  string   `arg:"" help:"Server name."`
	Cmd   string   `arg:"" help:"Command to run."`
	Args  []string `arg:"" optional:"" help:"Command arguments."`
}

type claudeMockMCPRemoveCmd struct {
	Scope string `name:"scope" help:"Scope for the server." default:"user"`
	Name  string `arg:"" help:"Server name."`
}

type claudeMockConfig struct {
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

	LargeOutput bool

	VersionFail     bool
	VersionNoSemver bool

	MCPNotFound       bool
	MCPNotFoundScoped bool
	MCPRemoveFail     bool
	MCPAddFail        bool
	MCPAddNoOutput    bool
}

func loadClaudeMockConfig() claudeMockConfig {
	exitCode := 1
	if code := os.Getenv("MOCK_CLAUDE_EXIT_CODE"); code != "" {
		if parsed, err := strconv.Atoi(code); err == nil {
			exitCode = parsed
		}
	}

	return claudeMockConfig{
		LargeOutput:    envBool("MOCK_CLAUDE_LARGE_OUTPUT"),
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

		MCPNotFound:       envBool("MOCK_CLAUDE_MCP_NOT_FOUND"),
		MCPNotFoundScoped: envBool("MOCK_CLAUDE_MCP_NOT_FOUND_SCOPED"),
		MCPRemoveFail:     envBool("MOCK_CLAUDE_MCP_REMOVE_FAIL"),
		MCPAddFail:        envBool("MOCK_CLAUDE_MCP_ADD_FAIL"),
		MCPAddNoOutput:    envBool("MOCK_CLAUDE_MCP_ADD_NO_OUTPUT"),
	}
}

func envBool(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes"
}

type claudeMockContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type claudeMockMessage struct {
	Content []claudeMockContent `json:"content"`
}

type claudeMockEvent struct {
	Type      string             `json:"type"`
	Subtype   string             `json:"subtype,omitempty"`
	SessionID string             `json:"session_id,omitempty"`
	Message   *claudeMockMessage `json:"message,omitempty"`
	Result    string             `json:"result,omitempty"`
}

func writeMockInit(w io.Writer, sessionID string) error {
	return json.NewEncoder(w).Encode(claudeMockEvent{
		Type:      "system",
		Subtype:   "init",
		SessionID: sessionID,
	})
}

func writeMockAssistant(w io.Writer, text string) error {
	return json.NewEncoder(w).Encode(claudeMockEvent{
		Type: "assistant",
		Message: &claudeMockMessage{
			Content: []claudeMockContent{
				{Type: "text", Text: text},
			},
		},
	})
}

func writeMockResultSuccess(w io.Writer, result string) error {
	return json.NewEncoder(w).Encode(claudeMockEvent{
		Type:    "result",
		Subtype: "success",
		Result:  result,
	})
}

func writeMockResultError(w io.Writer, errMsg string) error {
	return json.NewEncoder(w).Encode(claudeMockEvent{
		Type:    "result",
		Subtype: "error",
		Result:  errMsg,
	})
}

func writeMockMalformedJSON(w io.Writer) error {
	_, err := fmt.Fprintln(w, `{"type":"system","subtype":"init","session_id":`)
	return err
}

func writeMockEmptyLine(w io.Writer) error {
	_, err := fmt.Fprintln(w)
	return err
}

func writeMockNonJSONLine(w io.Writer, text string) error {
	_, err := fmt.Fprintln(w, text)
	return err
}

func (cmd *claudeMockExecCmd) Run() error {
	config := loadClaudeMockConfig()

	prompt := readMockStdin()
	if prompt == "" {
		prompt = "MOCK_RESPONSE"
	}

	sessionID := "mock-session-id-12345"
	if cmd.Resume != "" {
		sessionID = cmd.Resume
	}

	responseText := cmd.buildResponseText(prompt)
	if config.LargeOutput {
		responseText = strings.Repeat("x", 100_000)
	}

	if config.Signal != "" {
		handleMockSignal(config)
		return nil
	}

	_ = writeMockNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.MalformedJSON {
		_ = writeMockMalformedJSON(os.Stdout)
		return nil
	}

	_ = writeMockInit(os.Stdout, sessionID)

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	time.Sleep(10 * time.Millisecond)

	if config.PartialFail {
		_ = writeMockAssistant(os.Stdout, responseText)
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	if config.MultiAssistant {
		_ = writeMockAssistant(os.Stdout, "First part of response")
		if config.EmptyLines {
			_ = writeMockEmptyLine(os.Stdout)
		}
		_ = writeMockAssistant(os.Stdout, "Second part of response")
		if config.EmptyLines {
			_ = writeMockEmptyLine(os.Stdout)
		}
		_ = writeMockAssistant(os.Stdout, "Third part of response")
		responseText = "First part of responseSecond part of responseThird part of response"
	} else {
		_ = writeMockAssistant(os.Stdout, responseText)
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.NoResult {
		return nil
	}

	if config.ResultError {
		_ = writeMockResultError(os.Stdout, "execution error occurred")
		return nil
	}

	_ = writeMockResultSuccess(os.Stdout, responseText)

	if config.Fail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	return nil
}

func (cmd *claudeMockExecCmd) buildResponseText(prompt string) string {
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

func (cmd *claudeMockMCPAddCmd) Run() error {
	config := loadClaudeMockConfig()

	if config.MCPAddFail {
		if config.MCPAddNoOutput {
			os.Exit(1)
		}
		_, _ = fmt.Fprintln(os.Stderr, "MCP server registration failed")
		os.Exit(1)
	}

	return nil
}

func (cmd *claudeMockMCPRemoveCmd) Run() error {
	config := loadClaudeMockConfig()

	if config.MCPNotFound {
		_, _ = fmt.Fprintln(os.Stdout, "No MCP server found with name: "+cmd.Name)
		os.Exit(1)
	}

	if config.MCPNotFoundScoped {
		_, _ = fmt.Fprintln(os.Stdout, "No user-scoped MCP server found with name: "+cmd.Name)
		os.Exit(1)
	}

	if config.MCPRemoveFail {
		_, _ = fmt.Fprintln(os.Stderr, "MCP server removal failed")
		os.Exit(1)
	}

	return nil
}

func readMockStdin() string {
	scanner := bufio.NewScanner(os.Stdin)
	var builder strings.Builder
	for scanner.Scan() {
		builder.WriteString(scanner.Text())
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func handleMockSignal(config claudeMockConfig) {
	_ = writeMockNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")
	_ = writeMockInit(os.Stdout, "mock-session-id-12345")

	switch strings.ToUpper(config.Signal) {
	case "SIGINT":
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	case "SIGTERM":
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	default:
		_, _ = fmt.Fprintln(os.Stderr, "unknown signal: "+config.Signal)
		os.Exit(1)
	}

	select {}
}
