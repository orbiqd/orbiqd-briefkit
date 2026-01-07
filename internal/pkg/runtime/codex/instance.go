package codex

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
	"slices"
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

type codexEvent struct {
	Type     string `json:"type"`
	ThreadID string `json:"thread_id"`
	Item     struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"item"`
}

func newInstance(ctx context.Context, executionId agent.ExecutionID, executionInput agent.ExecutionInput, runtimeConfig Config, runtimeFeatures agent.RuntimeFeatures, logDir string) (*Instance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	path, err := process.LookupExecutable(ctx, []string{"codex"})
	if err != nil {
		return nil, fmt.Errorf("lookup codex executable: %w", err)
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
	runtimeArguments.SetFlag("json")

	instanceArgumentsList := slices.Concat(
		[]string{"exec"},
		runtimeArguments.ToList(),
	)

	if executionInput.ConversationID != nil {
		instanceArgumentsList = append(instanceArgumentsList, "resume", string(*executionInput.ConversationID), "-")
	} else {
		instanceArgumentsList = append(instanceArgumentsList, "-")
	}

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
		events: make(chan agent.RuntimeEvent, 2),
		done:   make(chan struct{}),
	}

	// Setup logging
	sessionLogDir := filepath.Join(logDir, "codex", string(executionId), time.Now().Format("2006-01-02_15-04-05"))
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

	cmd.Stdin = io.TeeReader(strings.NewReader(executionInput.Prompt), stdinLog)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("capture codex stdout: %w", err)
	}
	instance.stdout = pipe
	// We wrap the pipe in a TeeReader to log its content as it's being read by the decoder.
	// But sinceStdoutPipe returns a ReadCloser, we need to handle closing correctly.
	// Actually we will wrap the read side in watchCodexEvents.

	cmd.Stderr = io.MultiWriter(&instance.stderr, stderrLog)

	if err := instance.cmd.Start(); err != nil {
		return nil, fmt.Errorf("start codex: %w", err)
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

	parseErr := instance.watchCodexEvents(stdoutLog)
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

func (instance *Instance) watchCodexEvents(stdoutLog io.Writer) error {
	scanner := bufio.NewScanner(io.TeeReader(instance.stdout, stdoutLog))

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "{") {
			slog.Debug("Skipping non-JSON line from Codex CLI", slog.String("line", line))
			continue
		}

		var event codexEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			slog.Warn("Failed to unmarshal JSON candidate from Codex CLI", slog.String("line", line), slog.Any("error", err))
			continue
		}

		slog.Debug("Codex event received.", slog.String("eventType", event.Type))

		switch event.Type {
		case "thread.started":
			if event.ThreadID != "" {
				instance.result.ConversationID = agent.ConversationID(event.ThreadID)
			}
		case "item.completed":
			if event.Item.Type == "agent_message" {
				instance.result.Response = event.Item.Text
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read codex output: %w", err)
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
