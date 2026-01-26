package claude

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

type claudeEvent struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Message   struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text,omitempty"`
		} `json:"content,omitempty"`
	} `json:"message,omitempty"`
	Result string `json:"result,omitempty"`
}

func newInstance(ctx context.Context, executionId agent.ExecutionID, executionInput agent.ExecutionInput, runtimeConfig RuntimeConfig, runtimeFeatures agent.RuntimeFeatures, logDir string) (*Instance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path, err := locateExecutable(ctx)
	if err != nil {
		return nil, err
	}

	runtimeArguments := NewClaudeArguments()

	err = runtimeArguments.ApplyRuntimeConfig(runtimeConfig)
	if err != nil {
		return nil, fmt.Errorf("apply runtime config: %w", err)
	}

	err = runtimeArguments.ApplyRuntimeFeatures(runtimeFeatures)
	if err != nil {
		return nil, fmt.Errorf("apply runtime features: %w", err)
	}

	err = runtimeArguments.ApplyExecutionInput(executionInput)
	if err != nil {
		return nil, fmt.Errorf("apply execution input: %w", err)
	}

	instanceArgumentsList := runtimeArguments.ToSlice()

	// #nosec G204 - path comes from LookupExecutable with hardcoded name, arguments are constructed internally
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

	sessionLogDir := filepath.Join(logDir, "claude", string(executionId), time.Now().Format("2006-01-02_15-04-05"))
	if err := os.MkdirAll(sessionLogDir, 0750); err != nil {
		return nil, fmt.Errorf("create session log directory: %w", err)
	}

	// #nosec G304 - sessionLogDir is constructed from controlled values
	stdinLog, err := os.Create(filepath.Join(sessionLogDir, "stdin.log"))
	if err != nil {
		return nil, fmt.Errorf("create stdin log: %w", err)
	}
	instance.closers = append(instance.closers, stdinLog)

	// #nosec G304 - sessionLogDir is constructed from controlled values
	stdoutLog, err := os.Create(filepath.Join(sessionLogDir, "stdout.log"))
	if err != nil {
		return nil, fmt.Errorf("create stdout log: %w", err)
	}
	instance.closers = append(instance.closers, stdoutLog)

	// #nosec G304 - sessionLogDir is constructed from controlled values
	stderrLog, err := os.Create(filepath.Join(sessionLogDir, "stderr.log"))
	if err != nil {
		return nil, fmt.Errorf("create stderr log: %w", err)
	}
	instance.closers = append(instance.closers, stderrLog)

	cmd.Stdin = io.TeeReader(strings.NewReader(executionInput.Prompt), stdinLog)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("capture claude stdout: %w", err)
	}
	instance.stdout = pipe

	cmd.Stderr = io.MultiWriter(&instance.stderr, stderrLog)

	if err := instance.cmd.Start(); err != nil {
		return nil, fmt.Errorf("start claude: %w", err)
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

	parseErr := instance.watchClaudeEvents(stdoutLog)

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

func (instance *Instance) watchClaudeEvents(stdoutLog io.Writer) error {
	scanner := bufio.NewScanner(io.TeeReader(instance.stdout, stdoutLog))

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "{") {
			slog.Debug("Skipping non-JSON line from Claude CLI", slog.String("line", line))
			continue
		}

		var event claudeEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			slog.Warn("Failed to unmarshal JSON candidate from Claude CLI", slog.String("line", line), slog.Any("error", err))
			continue
		}

		slog.Debug("Claude event received.", slog.String("eventType", event.Type), slog.String("eventSubtype", event.Subtype))

		switch event.Type {
		case "system":
			if event.Subtype == "init" && event.SessionID != "" {
				instance.result.ConversationID = agent.ConversationID(event.SessionID)
			}
		case "assistant":
			for _, content := range event.Message.Content {
				if content.Type == "text" {
					instance.result.Response += content.Text
				}
			}
		case "result":
			if event.Subtype == "success" && event.Result != "" {
				instance.result.Response = event.Result
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read claude output: %w", err)
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
