package codex

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
	"github.com/google/uuid"
)

const codexMockEnvKey = "BRIEFKIT_CODEX_MOCK"

func init() {
	if os.Getenv(codexMockEnvKey) != "1" {
		return
	}

	flag.Bool("version", false, "")
	flag.Bool("json", false, "")
	flag.Bool("skip-git-repo-check", false, "")
	flag.String("model", "", "")
	flag.String("sandbox", "", "")
}

func TestMain(m *testing.M) {
	if os.Getenv(codexMockEnvKey) == "1" {
		os.Exit(runCodexMock())
	}

	os.Exit(m.Run())
}

func runCodexMock() int {
	ctx := kong.Parse(&codexMockCLI,
		kong.Name("codex"),
		kong.Description("Mock Codex CLI for testing."),
		kong.NoDefaultHelp(),
	)

	if err := ctx.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

var codexMockCLI struct {
	MCP  codexMockMCPCmd  `cmd:"" help:"MCP server management."`
	Exec codexMockExecCmd `cmd:"" help:"Execute prompt."`

	Version codexMockVersionFlag `name:"version" help:"Print version information."`
}

type codexMockVersionFlag bool

func (v codexMockVersionFlag) BeforeApply(app *kong.Kong) error {
	config := loadCodexMockConfig()

	if config.VersionFail {
		_, _ = fmt.Fprintln(os.Stderr, "version check failed")
		app.Exit(1)
	}

	if config.VersionNoSemver {
		fmt.Println("codex-cli (development build)")
	} else {
		fmt.Println("codex-cli 1.0.0-mock")
	}

	app.Exit(0)
	return nil
}

type codexMockExecCmd struct {
	Resume  codexMockResumeCmd      `cmd:"" help:"Resume session."`
	Default codexMockExecDefaultCmd `cmd:"" default:"withargs" help:"Execute prompt directly."`
}

type codexMockExecDefaultCmd struct {
	JSON             bool     `name:"json" help:"JSON output." default:"true"`
	SkipGitRepoCheck bool     `name:"skip-git-repo-check" help:"Skip git repo check."`
	Model            string   `name:"model" short:"m" help:"Model to use."`
	Sandbox          string   `name:"sandbox" short:"s" help:"Sandbox mode."`
	Config           []string `name:"config" short:"c" help:"Config overrides."`
	Prompt           string   `arg:"" optional:"" help:"Prompt (- for stdin)."`
}

type codexMockResumeCmd struct {
	JSON             bool     `name:"json" help:"JSON output." default:"true"`
	SkipGitRepoCheck bool     `name:"skip-git-repo-check" help:"Skip git repo check."`
	Model            string   `name:"model" short:"m" help:"Model to use."`
	Sandbox          string   `name:"sandbox" short:"s" help:"Sandbox mode."`
	Config           []string `name:"config" short:"c" help:"Config overrides."`
	SessionID        string   `arg:"" help:"Session ID to resume."`
	Prompt           string   `arg:"" optional:"" help:"Prompt (- for stdin)."`
}

type codexMockMCPCmd struct {
	Add    codexMockMCPAddCmd    `cmd:"" help:"Add MCP server."`
	Remove codexMockMCPRemoveCmd `cmd:"" help:"Remove MCP server."`
}

type codexMockMCPAddCmd struct {
	Name    string   `arg:"" help:"Server name."`
	Command []string `arg:"" passthrough:"" help:"Command after --."`
}

type codexMockMCPRemoveCmd struct {
	Name string `arg:"" help:"Server name."`
}

type codexMockConfig struct {
	Fail          bool
	ExitCode      int
	Stderr        string
	MalformedJSON bool
	NoResult      bool
	NoThreadStart bool
	Signal        string
	PartialFail   bool

	EmptyLines           bool
	WhitespaceVariations bool
	UnknownEvents        bool
	MultiItem            bool
	EmptyText            bool
	OtherItemType        bool
	EmptyStdin           bool
	MixedOutput          bool

	LargeOutput bool

	VersionFail     bool
	VersionNoSemver bool

	MCPNotFound    bool
	MCPRemoveFail  bool
	MCPAddFail     bool
	MCPAddNoOutput bool

	InvalidSession bool
}

func loadCodexMockConfig() codexMockConfig {
	exitCode := 1
	if code := os.Getenv("MOCK_CODEX_EXIT_CODE"); code != "" {
		if parsed, err := strconv.Atoi(code); err == nil {
			exitCode = parsed
		}
	}

	return codexMockConfig{
		LargeOutput:   envBool("MOCK_CODEX_LARGE_OUTPUT"),
		Fail:          envBool("MOCK_CODEX_FAIL"),
		ExitCode:      exitCode,
		Stderr:        os.Getenv("MOCK_CODEX_STDERR"),
		MalformedJSON: envBool("MOCK_CODEX_MALFORMED_JSON"),
		NoResult:      envBool("MOCK_CODEX_NO_RESULT"),
		NoThreadStart: envBool("MOCK_CODEX_NO_THREAD_STARTED"),
		Signal:        os.Getenv("MOCK_CODEX_SIGNAL"),
		PartialFail:   envBool("MOCK_CODEX_PARTIAL_FAIL"),

		EmptyLines:           envBool("MOCK_CODEX_EMPTY_LINES"),
		WhitespaceVariations: envBool("MOCK_CODEX_WHITESPACE_VARIATIONS"),
		UnknownEvents:        envBool("MOCK_CODEX_UNKNOWN_EVENTS"),
		MultiItem:            envBool("MOCK_CODEX_MULTI_ITEM"),
		EmptyText:            envBool("MOCK_CODEX_EMPTY_TEXT"),
		OtherItemType:        envBool("MOCK_CODEX_OTHER_ITEM_TYPE"),
		EmptyStdin:           envBool("MOCK_CODEX_EMPTY_STDIN"),
		MixedOutput:          envBool("MOCK_CODEX_MIXED_OUTPUT"),

		VersionFail:     envBool("MOCK_CODEX_VERSION_FAIL"),
		VersionNoSemver: envBool("MOCK_CODEX_VERSION_NO_SEMVER"),

		MCPNotFound:    envBool("MOCK_CODEX_MCP_NOT_FOUND"),
		MCPRemoveFail:  envBool("MOCK_CODEX_MCP_REMOVE_FAIL"),
		MCPAddFail:     envBool("MOCK_CODEX_MCP_ADD_FAIL"),
		MCPAddNoOutput: envBool("MOCK_CODEX_MCP_ADD_NO_OUTPUT"),

		InvalidSession: envBool("MOCK_CODEX_INVALID_SESSION"),
	}
}

func envBool(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes"
}

type codexMockItem struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Text string `json:"text"`
}

type codexMockUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type codexMockEvent struct {
	Type     string          `json:"type"`
	ThreadID string          `json:"thread_id,omitempty"`
	Item     *codexMockItem  `json:"item,omitempty"`
	Usage    *codexMockUsage `json:"usage,omitempty"`
	Data     string          `json:"data,omitempty"`
	Status   string          `json:"status,omitempty"`
}

func writeMockThreadStarted(w io.Writer, threadID string, whitespace bool) error {
	event := codexMockEvent{Type: "thread.started", ThreadID: threadID}
	return writeMockEvent(w, event, whitespace)
}

func writeMockTurnStarted(w io.Writer, whitespace bool) error {
	event := codexMockEvent{Type: "turn.started"}
	return writeMockEvent(w, event, whitespace)
}

func writeMockItemCompleted(w io.Writer, itemID, itemType, text string, whitespace bool) error {
	event := codexMockEvent{
		Type: "item.completed",
		Item: &codexMockItem{ID: itemID, Type: itemType, Text: text},
	}
	return writeMockEvent(w, event, whitespace)
}

func writeMockTurnCompleted(w io.Writer, inputTokens, outputTokens int, whitespace bool) error {
	event := codexMockEvent{
		Type:  "turn.completed",
		Usage: &codexMockUsage{InputTokens: inputTokens, OutputTokens: outputTokens},
	}
	return writeMockEvent(w, event, whitespace)
}

func writeMockUnknownEvent(w io.Writer, eventType, data, status string, whitespace bool) error {
	event := codexMockEvent{Type: eventType, Data: data, Status: status}
	return writeMockEvent(w, event, whitespace)
}

func writeMockEvent(w io.Writer, event codexMockEvent, whitespace bool) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	line := string(data)
	if whitespace {
		line = "  " + line + "  \r\n"
	} else {
		line = line + "\n"
	}

	_, err = w.Write([]byte(line))
	return err
}

func writeMockMalformedJSON(w io.Writer) error {
	_, err := fmt.Fprintln(w, `{"type":"item.completed","item":`)
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

func (cmd *codexMockExecDefaultCmd) Run() error {
	config := loadCodexMockConfig()

	prompt := readMockStdin()
	if prompt == "" && !config.EmptyStdin {
		prompt = "MOCK_RESPONSE"
	}

	threadID := uuid.New().String()

	responseText := cmd.buildResponseText(prompt, "")
	if config.LargeOutput {
		responseText = strings.Repeat("x", 100_000)
	}

	if config.Signal != "" {
		handleMockSignal(config, threadID)
		return nil
	}

	_ = writeMockNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.MixedOutput {
		_, _ = fmt.Fprintln(os.Stderr, "Loading configuration...")
	}

	if config.MalformedJSON {
		_ = writeMockThreadStarted(os.Stdout, threadID, config.WhitespaceVariations)
		_ = writeMockMalformedJSON(os.Stdout)
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", responseText, config.WhitespaceVariations)
		return nil
	}

	if !config.NoThreadStart {
		_ = writeMockThreadStarted(os.Stdout, threadID, config.WhitespaceVariations)
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.UnknownEvents {
		_ = writeMockUnknownEvent(os.Stdout, "unknown.event", "ignored", "", config.WhitespaceVariations)
		_ = writeMockUnknownEvent(os.Stdout, "system.status", "", "processing", config.WhitespaceVariations)
	}

	_ = writeMockTurnStarted(os.Stdout, config.WhitespaceVariations)

	if config.MixedOutput {
		_, _ = fmt.Fprintln(os.Stderr, "Processing prompt...")
	}

	time.Sleep(10 * time.Millisecond)

	if config.PartialFail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	if config.NoResult {
		return nil
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.OtherItemType {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "tool_call", "ignored", config.WhitespaceVariations)
		_ = writeMockItemCompleted(os.Stdout, "item_1", "agent_message", responseText, config.WhitespaceVariations)
	} else if config.MultiItem {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", "First response", config.WhitespaceVariations)
		_ = writeMockItemCompleted(os.Stdout, "item_1", "agent_message", "Second response", config.WhitespaceVariations)
	} else if config.EmptyText {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", "", config.WhitespaceVariations)
	} else {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", responseText, config.WhitespaceVariations)
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	_ = writeMockTurnCompleted(os.Stdout, 100, 10, config.WhitespaceVariations)

	if config.Fail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	return nil
}

func (cmd *codexMockExecDefaultCmd) buildResponseText(prompt string, sessionID string) string {
	text := "Mock response to: " + prompt

	if cmd.Model != "" {
		text += fmt.Sprintf(" (Model: %s)", cmd.Model)
	}

	if cmd.Sandbox != "" {
		text += fmt.Sprintf(" (Sandbox: %s)", cmd.Sandbox)
	}

	if sessionID != "" {
		text += fmt.Sprintf(" (Resumed: %s)", sessionID)
	}

	return text
}

func (cmd *codexMockResumeCmd) Run() error {
	config := loadCodexMockConfig()

	prompt := readMockStdin()
	if prompt == "" && !config.EmptyStdin {
		prompt = "MOCK_RESPONSE"
	}

	threadID := cmd.SessionID
	if threadID == "" && !config.InvalidSession {
		threadID = uuid.New().String()
	}

	responseText := "Mock response to: " + prompt + fmt.Sprintf(" (Resumed: %s)", cmd.SessionID)

	if config.Signal != "" {
		handleMockSignal(config, threadID)
		return nil
	}

	_ = writeMockNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.MixedOutput {
		_, _ = fmt.Fprintln(os.Stderr, "Loading configuration...")
	}

	if config.MalformedJSON {
		_ = writeMockThreadStarted(os.Stdout, threadID, config.WhitespaceVariations)
		_ = writeMockMalformedJSON(os.Stdout)
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", responseText, config.WhitespaceVariations)
		return nil
	}

	if !config.NoThreadStart {
		_ = writeMockThreadStarted(os.Stdout, threadID, config.WhitespaceVariations)
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.UnknownEvents {
		_ = writeMockUnknownEvent(os.Stdout, "unknown.event", "ignored", "", config.WhitespaceVariations)
		_ = writeMockUnknownEvent(os.Stdout, "system.status", "", "processing", config.WhitespaceVariations)
	}

	_ = writeMockTurnStarted(os.Stdout, config.WhitespaceVariations)

	if config.MixedOutput {
		_, _ = fmt.Fprintln(os.Stderr, "Processing prompt...")
	}

	time.Sleep(10 * time.Millisecond)

	if config.PartialFail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	if config.NoResult {
		return nil
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.OtherItemType {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "tool_call", "ignored", config.WhitespaceVariations)
		_ = writeMockItemCompleted(os.Stdout, "item_1", "agent_message", responseText, config.WhitespaceVariations)
	} else if config.MultiItem {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", "First response", config.WhitespaceVariations)
		_ = writeMockItemCompleted(os.Stdout, "item_1", "agent_message", "Second response", config.WhitespaceVariations)
	} else if config.EmptyText {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", "", config.WhitespaceVariations)
	} else {
		_ = writeMockItemCompleted(os.Stdout, "item_0", "agent_message", responseText, config.WhitespaceVariations)
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	_ = writeMockTurnCompleted(os.Stdout, 100, 10, config.WhitespaceVariations)

	if config.Fail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	return nil
}

func (cmd *codexMockMCPAddCmd) Run() error {
	config := loadCodexMockConfig()

	if config.MCPAddFail {
		if config.MCPAddNoOutput {
			os.Exit(1)
		}
		fmt.Printf("Error: MCP server '%s' already exists\n", cmd.Name)
		os.Exit(1)
	}

	fmt.Printf("Added global MCP server '%s'.\n", cmd.Name)
	return nil
}

func (cmd *codexMockMCPRemoveCmd) Run() error {
	config := loadCodexMockConfig()

	if config.MCPNotFound {
		fmt.Printf("No MCP server named '%s' found.\n", cmd.Name)
		return nil
	}

	if config.MCPRemoveFail {
		_, _ = fmt.Fprintln(os.Stderr, "Permission denied: cannot modify config")
		os.Exit(1)
	}

	fmt.Printf("Removed global MCP server '%s'.\n", cmd.Name)
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

func handleMockSignal(config codexMockConfig, threadID string) {
	_ = writeMockNonJSONLine(os.Stdout, "Mock initialization info... (non-JSON)")
	_ = writeMockThreadStarted(os.Stdout, threadID, false)

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
