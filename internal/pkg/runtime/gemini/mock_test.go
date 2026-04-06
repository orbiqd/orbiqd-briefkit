package gemini

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

const geminiMockEnvKey = "BRIEFKIT_GEMINI_MOCK"

func init() {
	if os.Getenv(geminiMockEnvKey) != "1" {
		return
	}

	flag.Bool("version", false, "")
	flag.String("output-format", "", "")
	flag.String("model", "", "")
	flag.String("resume", "", "")
	flag.String("sandbox", "", "")
	flag.String("scope", "", "")
}

func TestMain(m *testing.M) {
	if os.Getenv(geminiMockEnvKey) == "1" {
		os.Exit(runGeminiMock())
	}

	os.Exit(m.Run())
}

func runGeminiMock() int {
	ctx := kong.Parse(&geminiMockCLI,
		kong.Name("gemini"),
		kong.Description("Mock Gemini CLI for testing."),
		kong.NoDefaultHelp(),
	)

	if err := ctx.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

var geminiMockCLI struct {
	MCP     geminiMockMCPCmd     `cmd:"" help:"MCP server management."`
	Default geminiMockDefaultCmd `cmd:"" default:"withargs" help:"Execute prompt."`

	Version geminiMockVersionFlag `name:"version" help:"Print version information."`
}

type geminiMockVersionFlag bool

func (v geminiMockVersionFlag) BeforeApply(app *kong.Kong) error {
	config := loadGeminiMockConfig()

	if config.VersionFail {
		_, _ = fmt.Fprintln(os.Stderr, "version check failed")
		app.Exit(1)
	}

	if config.VersionNoSemver {
		fmt.Println("gemini-cli (development build)")
	} else {
		fmt.Println("gemini-cli 1.0.0-mock")
	}

	app.Exit(0)
	return nil
}

type geminiMockDefaultCmd struct {
	OutputFormat string `name:"output-format" help:"Output format." default:"stream-json"`
	Model        string `help:"Model to use."`
	Resume       string `help:"Session ID to resume."`
	Sandbox      string `help:"Sandbox mode."`
}

type geminiMockMCPCmd struct {
	Add    geminiMockMCPAddCmd    `cmd:"" help:"Add MCP server."`
	Remove geminiMockMCPRemoveCmd `cmd:"" help:"Remove MCP server."`
}

type geminiMockMCPAddCmd struct {
	Scope string   `name:"scope" help:"Scope for the server." default:"user"`
	Name  string   `arg:"" help:"Server name."`
	Cmd   string   `arg:"" help:"Command to run."`
	Args  []string `arg:"" optional:"" help:"Command arguments."`
}

type geminiMockMCPRemoveCmd struct {
	Scope string `name:"scope" help:"Scope for the server." default:"user"`
	Name  string `arg:"" help:"Server name."`
}

type geminiMockConfig struct {
	Fail          bool
	ExitCode      int
	Stderr        string
	MalformedJSON bool
	NoInit        bool
	NoMessage     bool
	Signal        string
	PartialFail   bool
	MultiMessage  bool
	EmptyLines    bool
	UnknownEvents bool
	ErrorResult   bool

	LargeOutput bool

	VersionFail     bool
	VersionNoSemver bool

	MCPNotFound    bool
	MCPRemoveFail  bool
	MCPAddFail     bool
	MCPAddNoOutput bool

	InvalidSession bool
}

func loadGeminiMockConfig() geminiMockConfig {
	exitCode := 1
	if code := os.Getenv("MOCK_GEMINI_EXIT_CODE"); code != "" {
		if parsed, err := strconv.Atoi(code); err == nil {
			exitCode = parsed
		}
	}

	return geminiMockConfig{
		LargeOutput:   envBool("MOCK_GEMINI_LARGE_OUTPUT"),
		Fail:          envBool("MOCK_GEMINI_FAIL"),
		ExitCode:      exitCode,
		Stderr:        os.Getenv("MOCK_GEMINI_STDERR"),
		MalformedJSON: envBool("MOCK_GEMINI_MALFORMED_JSON"),
		NoInit:        envBool("MOCK_GEMINI_NO_INIT"),
		NoMessage:     envBool("MOCK_GEMINI_NO_MESSAGE"),
		Signal:        os.Getenv("MOCK_GEMINI_SIGNAL"),
		PartialFail:   envBool("MOCK_GEMINI_PARTIAL_FAIL"),
		MultiMessage:  envBool("MOCK_GEMINI_MULTI_MESSAGE"),
		EmptyLines:    envBool("MOCK_GEMINI_EMPTY_LINES"),
		UnknownEvents: envBool("MOCK_GEMINI_UNKNOWN_EVENTS"),
		ErrorResult:   envBool("MOCK_GEMINI_ERROR_RESULT"),

		VersionFail:     envBool("MOCK_GEMINI_VERSION_FAIL"),
		VersionNoSemver: envBool("MOCK_GEMINI_VERSION_NO_SEMVER"),

		MCPNotFound:    envBool("MOCK_GEMINI_MCP_NOT_FOUND"),
		MCPRemoveFail:  envBool("MOCK_GEMINI_MCP_REMOVE_FAIL"),
		MCPAddFail:     envBool("MOCK_GEMINI_MCP_ADD_FAIL"),
		MCPAddNoOutput: envBool("MOCK_GEMINI_MCP_ADD_NO_OUTPUT"),

		InvalidSession: envBool("MOCK_GEMINI_INVALID_SESSION"),
	}
}

func envBool(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes"
}

type geminiMockEvent struct {
	Type      string           `json:"type"`
	SessionID string           `json:"session_id,omitempty"`
	Model     string           `json:"model,omitempty"`
	Role      string           `json:"role,omitempty"`
	Content   string           `json:"content,omitempty"`
	Delta     bool             `json:"delta,omitempty"`
	Status    string           `json:"status,omitempty"`
	Error     *geminiMockError `json:"error,omitempty"`
	Stats     *geminiMockStats `json:"stats,omitempty"`
	Data      string           `json:"data,omitempty"`
}

type geminiMockError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type geminiMockStats struct {
	TotalTokens  int `json:"total_tokens"`
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func writeMockInit(w io.Writer, sessionID, model string) error {
	return json.NewEncoder(w).Encode(geminiMockEvent{
		Type:      "init",
		SessionID: sessionID,
		Model:     model,
	})
}

func writeMockAssistantMessage(w io.Writer, content string) error {
	return json.NewEncoder(w).Encode(geminiMockEvent{
		Type:    "message",
		Role:    "assistant",
		Content: content,
		Delta:   true,
	})
}

func writeMockResultSuccess(w io.Writer) error {
	return json.NewEncoder(w).Encode(geminiMockEvent{
		Type:   "result",
		Status: "success",
		Stats: &geminiMockStats{
			TotalTokens:  110,
			InputTokens:  100,
			OutputTokens: 10,
		},
	})
}

func writeMockResultError(w io.Writer, errMsg string) error {
	return json.NewEncoder(w).Encode(geminiMockEvent{
		Type:   "result",
		Status: "error",
		Error: &geminiMockError{
			Type:    "Error",
			Message: errMsg,
		},
	})
}

func writeMockUnknownEvent(w io.Writer, eventType, data string) error {
	return json.NewEncoder(w).Encode(geminiMockEvent{
		Type: eventType,
		Data: data,
	})
}

func writeMockEmptyLine(w io.Writer) error {
	_, err := fmt.Fprintln(w)
	return err
}

func writeMockNonJSONLine(w io.Writer, text string) error {
	_, err := fmt.Fprintln(w, text)
	return err
}

func writeMockMalformedJSON(w io.Writer) error {
	_, err := fmt.Fprintln(w, `{"type":"init","session_id":`)
	return err
}

func (cmd *geminiMockDefaultCmd) Run() error {
	config := loadGeminiMockConfig()

	prompt := readMockStdin()
	if prompt == "" {
		prompt = "MOCK_RESPONSE"
	}

	sessionID := uuid.New().String()
	if cmd.Resume != "" {
		sessionID = cmd.Resume
	}

	model := "gemini-2.0-flash"
	if cmd.Model != "" {
		model = cmd.Model
	}

	responseText := cmd.buildResponseText(prompt)
	if config.LargeOutput {
		responseText = strings.Repeat("x", 100_000)
	}

	if config.Signal != "" {
		handleMockSignal(config, sessionID, model)
		return nil
	}

	_ = writeMockNonJSONLine(os.Stdout, "Loaded cached credentials.")

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.MalformedJSON {
		_ = writeMockMalformedJSON(os.Stdout)
		return nil
	}

	if !config.NoInit {
		_ = writeMockInit(os.Stdout, sessionID, model)
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.UnknownEvents {
		_ = writeMockUnknownEvent(os.Stdout, "unknown.event", "ignored")
		_ = writeMockUnknownEvent(os.Stdout, "system.status", "processing")
	}

	time.Sleep(10 * time.Millisecond)

	if config.PartialFail {
		if !config.NoMessage {
			_ = writeMockAssistantMessage(os.Stdout, responseText)
		}
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	if !config.NoMessage {
		if config.MultiMessage {
			_ = writeMockAssistantMessage(os.Stdout, "First part of response")
			if config.EmptyLines {
				_ = writeMockEmptyLine(os.Stdout)
			}
			_ = writeMockAssistantMessage(os.Stdout, "Second part of response")
			if config.EmptyLines {
				_ = writeMockEmptyLine(os.Stdout)
			}
			_ = writeMockAssistantMessage(os.Stdout, "Third part of response")
		} else {
			_ = writeMockAssistantMessage(os.Stdout, responseText)
		}
	}

	if config.EmptyLines {
		_ = writeMockEmptyLine(os.Stdout)
	}

	if config.ErrorResult {
		_ = writeMockResultError(os.Stdout, "execution error occurred")
		return nil
	}

	_ = writeMockResultSuccess(os.Stdout)

	if config.Fail {
		if config.Stderr != "" {
			_, _ = fmt.Fprintln(os.Stderr, config.Stderr)
		}
		os.Exit(config.ExitCode)
	}

	return nil
}

func (cmd *geminiMockDefaultCmd) buildResponseText(prompt string) string {
	text := "Mock response to: " + prompt

	if cmd.Model != "" {
		text += fmt.Sprintf(" (Model: %s)", cmd.Model)
	}

	if cmd.Resume != "" {
		text += fmt.Sprintf(" (Resumed: %s)", cmd.Resume)
	}

	if cmd.Sandbox != "" {
		text += fmt.Sprintf(" (Sandbox: %s)", cmd.Sandbox)
	}

	return text
}

func (cmd *geminiMockMCPAddCmd) Run() error {
	config := loadGeminiMockConfig()

	if config.MCPAddFail {
		if config.MCPAddNoOutput {
			os.Exit(1)
		}
		fmt.Printf("MCP server \"%s\" registration failed\n", cmd.Name)
		os.Exit(1)
	}

	fmt.Printf("MCP server \"%s\" added to user settings. (stdio)\n", cmd.Name)
	return nil
}

func (cmd *geminiMockMCPRemoveCmd) Run() error {
	config := loadGeminiMockConfig()

	if config.MCPNotFound {
		fmt.Printf("Server \"%s\" not found in user settings.\n", cmd.Name)
		return nil
	}

	if config.MCPRemoveFail {
		_, _ = fmt.Fprintln(os.Stderr, "MCP server removal failed")
		os.Exit(1)
	}

	fmt.Printf("Server \"%s\" removed from user settings.\n", cmd.Name)
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

func handleMockSignal(config geminiMockConfig, sessionID, model string) {
	_ = writeMockNonJSONLine(os.Stdout, "Loaded cached credentials.")
	_ = writeMockInit(os.Stdout, sessionID, model)

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
