package gemini

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/process"
)

type Instance struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser

	events chan agent.RuntimeEvent
	done   chan struct{}

	result agent.RuntimeResult
	err    error

	stderr strings.Builder

	closers []io.Closer
}

// geminiEvent represents the structure of JSON events emitted by the Gemini CLI.
type geminiEvent struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id,omitempty"`
	Role      string `json:"role,omitempty"`
	Content   string `json:"content,omitempty"`
	Delta     bool   `json:"delta,omitempty"`
}

func newInstance(ctx context.Context, executionId agent.ExecutionID, executionInput agent.ExecutionInput, runtimeConfig Config, runtimeFeatures agent.RuntimeFeatures, logDir string) (*Instance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path, err := process.LookupExecutable(ctx, []string{"gemini"})
	if err != nil {
		return nil, fmt.Errorf("lookup gemini executable: %w", err)
	}

	runtimeArguments := defaultArguments()

	err = applyRuntimeConfigArguments(runtimeArguments, runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("apply runtime config: %w", err)
	}

	err = applyRuntimeFeaturesArguments(runtimeArguments, runtimeFeatures)
	if err != nil {
		return nil, fmt.Errorf("apply runtime features: %w", err)
	}

	err = applyExecutionInputArguments(runtimeArguments, executionInput)
	if err != nil {
		return nil, fmt.Errorf("apply execution input: %w", err)
	}

	// Force JSON stream output for parsing
	if err = runtimeArguments.SetValue("output-format", "stream-json"); err != nil {
		return nil, fmt.Errorf("set output-format: %w", err)
	}

	// Construct arguments
	instanceArgumentsList := runtimeArguments.ToList()

	cmd := exec.CommandContext(ctx, path, instanceArgumentsList...)

	if executionInput.WorkingDirectory != nil && strings.TrimSpace(*executionInput.WorkingDirectory) != "" {
		cmd.Dir = *executionInput.WorkingDirectory
	} else {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("resolve working directory: %w", err)
		}
		cmd.Dir = workingDir
	}

	instance := &Instance{
		cmd:    cmd,
		events: make(chan agent.RuntimeEvent, 10),
		done:   make(chan struct{}),
	}

	// Setup logging
	sessionLogDir := filepath.Join(logDir, "gemini", string(executionId), time.Now().Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(sessionLogDir, 0755); err != nil {
		return nil, fmt.Errorf("create session log directory: %w", err)
	}

	stdinLog, err := os.Create(filepath.Join(sessionLogDir, "stdin.log"))
	if err != nil {
		return nil, fmt.Errorf("create stdin log: %w", err)
	}
	instance.closers = append(instance.closers, stdinLog)

	stdoutLog, err := os.Create(filepath.Join(sessionLogDir, "stdout.log"))
	if err != nil {
		return nil, fmt.Errorf("create stdout log: %w", err)
	}
	instance.closers = append(instance.closers, stdoutLog)

	stderrLog, err := os.Create(filepath.Join(sessionLogDir, "stderr.log"))
	if err != nil {
		return nil, fmt.Errorf("create stderr log: %w", err)
	}
	instance.closers = append(instance.closers, stderrLog)

	// Pipe Prompt to Stdin and log it
	cmd.Stdin = io.TeeReader(strings.NewReader(executionInput.Prompt), stdinLog)

	// Capture stdout for parsing
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("capture gemini stdout: %w", err)
	}
	instance.stdout = pipe

	// Capture stderr to buffer and log
	cmd.Stderr = io.MultiWriter(&instance.stderr, stderrLog)

	if err := instance.cmd.Start(); err != nil {
		return nil, fmt.Errorf("start gemini: %w", err)
	}

	instance.emitRuntimeEvent(agent.RuntimeStartedEvent{Timestamp: time.Now()})
	go instance.run(stdoutLog)

	return instance, nil
}

func (instance *Instance) run(stdoutLog io.Writer) {
	defer close(instance.done)
	defer close(instance.events)
	defer instance.emitRuntimeEvent(agent.RuntimeFinishedEvent{Timestamp: time.Now()})
	defer func() {
		for _, closer := range instance.closers {
			_ = closer.Close()
		}
	}()

	parseErr := instance.watchGeminiEvents(stdoutLog)

	if parseErr != nil {
		_, _ = io.Copy(io.Discard, instance.stdout)
	}

	waitErr := instance.cmd.Wait()

	if parseErr != nil {
		instance.err = &agent.RuntimeExecutionError{
			Message: parseErr.Error(),
			Cause:   parseErr,
		}
		return
	}

	if waitErr != nil {
		instance.err = instance.runtimeError(waitErr)
	}
}

func (instance *Instance) Events() <-chan agent.RuntimeEvent {
	return instance.events
}

func (instance *Instance) Wait(ctx context.Context) (agent.RuntimeResult, error) {
	select {
	case <-instance.done:
		return instance.result, instance.err
	case <-ctx.Done():
		return agent.RuntimeResult{}, ctx.Err()
	}
}

func (instance *Instance) watchGeminiEvents(stdoutLog io.Writer) error {
	// We use a scanner to read line by line because Gemini CLI might output non-JSON text
	// (e.g. "Loaded cached credentials.") mixed with JSON lines.
	scanner := bufio.NewScanner(io.TeeReader(instance.stdout, stdoutLog))

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Optimization: Try to unmarshal only if it looks like a JSON object
		if !strings.HasPrefix(line, "{") {
			slog.Debug("Skipping non-JSON line from Gemini CLI", slog.String("line", line))
			continue
		}

		var event geminiEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// If unmarshal fails, it might be a false positive or malformed JSON.
			// We log it and continue instead of failing the whole runtime.
			slog.Warn("Failed to unmarshal JSON candidate from Gemini CLI", slog.String("line", line), slog.Any("error", err))
			continue
		}

		slog.Debug("Gemini event received.", slog.String("eventType", event.Type))

		switch event.Type {
		case "init":
			if event.SessionID != "" {
				instance.result.ConversationID = agent.ConversationID(event.SessionID)
			}
		case "message":
			if event.Role == "assistant" {
				// We append content.
				instance.result.Response += event.Content
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read gemini output: %w", err)
	}

	return nil
}

func (instance *Instance) runtimeError(err error) error {
	message := strings.TrimSpace(instance.stderr.String())
	if message == "" {
		message = err.Error()
	}

	runtimeErr := &agent.RuntimeExecutionError{
		Message: message,
		Cause:   err,
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		runtimeErr.ExitCode = &code
	}

	return runtimeErr
}

func (instance *Instance) emitRuntimeEvent(event agent.RuntimeEvent) {
	if instance.events == nil {
		return
	}

	select {
	case instance.events <- event:
		slog.Debug("Runtime event emitted.", slog.String("eventKind", string(event.Kind())))
	default:
		slog.Warn("Runtime event dropped because the channel is full.", slog.String("eventKind", string(event.Kind())))
	}
}
