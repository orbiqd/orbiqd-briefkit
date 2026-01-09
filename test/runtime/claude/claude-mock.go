package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kong"
)

var CLI struct {
	Version         bool   `help:"Print version information and exit."`
	Print           bool   `help:"Print output."`
	Verbose         bool   `help:"Verbose output."`
	OutputFormat    string `help:"Output format."`
	Model           string `help:"Model to use."`
	Resume          string `help:"Conversation ID to resume."`
	Settings        string `help:"Settings JSON."`
	DisallowedTools string `help:"Disallowed tools."`

	// Allow for extra args that might be passed but not explicitly handled yet
	Extra []string `arg:"" optional:""`
}

// Struktury JSON (Events)
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

func main() {
	ctx := kong.Parse(&CLI)

	// 1. Obsługa wersji
	if CLI.Version {
		fmt.Println("claude-code 1.0.0-mock")
		ctx.Exit(0)
	}

	// 2. Odczyt promptu ze Stdin
	scanner := bufio.NewScanner(os.Stdin)
	var inputBuilder strings.Builder
	for scanner.Scan() {
		inputBuilder.WriteString(scanner.Text())
		inputBuilder.WriteString("\n")
	}
	prompt := strings.TrimSpace(inputBuilder.String())

	if prompt == "" {
		prompt = "MOCK_RESPONSE"
	}

	// 3. Generowanie strumienia JSON
	encoder := json.NewEncoder(os.Stdout)

	// Ustalanie SessionID (nowy lub wznowiony)
	sessionID := "mock-session-id-12345"
	if CLI.Resume != "" {
		sessionID = CLI.Resume
	}

	// Symulacja linii nie będącej JSON-em (np. logi, update checker itp.)
	fmt.Println("Mock initialization info... (non-JSON)")

	// Event 1: Init
	_ = encoder.Encode(event{
		Type:      "system",
		Subtype:   "init",
		SessionID: sessionID,
	})

	time.Sleep(10 * time.Millisecond)

	// Event 2: Assistant response
	responseVisibleText := "Mock response to: " + prompt

	// Jeśli podano model, dodajemy info
	if CLI.Model != "" {
		responseVisibleText += fmt.Sprintf(" (Model: %s)", CLI.Model)
	}

	_ = encoder.Encode(event{
		Type: "assistant",
		Message: &message{
			Content: []content{
				{
					Type: "text",
					Text: responseVisibleText,
				},
			},
		},
	})

	// Event 3: Result
	_ = encoder.Encode(event{
		Type:    "result",
		Subtype: "success",
		Result:  responseVisibleText,
	})
}
