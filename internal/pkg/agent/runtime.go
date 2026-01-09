package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type RuntimeKind string

// RuntimeConfig is an opaque, runtime-specific configuration payload.
type RuntimeConfig any

type RuntimeFeatures struct {
	// EnableWebSearch allows the agent to use internal web-search tool without accessing the internet.
	EnableWebSearch *bool `json:"enableWebSearch"`

	// EnableNetworkAccess allows the agent to access external network resources.
	EnableNetworkAccess *bool `json:"enableNetworkAccess"`

	// EnableSandbox overrides the agent's default sandbox configuration, forcing it to run in (or out of) sandbox mode.
	EnableSandbox *bool `json:"enableSandbox"`
}

// RuntimeEventKind describes the type of runtime event emitted by a runtime instance.
type RuntimeEventKind string

const (
	// RuntimeEventStarted indicates the runtime instance has started.
	RuntimeEventStarted RuntimeEventKind = "runtime-started"

	// RuntimeEventFinished indicates the runtime instance has finished.
	RuntimeEventFinished RuntimeEventKind = "runtime-finished"
)

// RuntimeEvent represents a runtime event emitted by a runtime instance.
type RuntimeEvent interface {
	// Kind returns the kind of the runtime event.
	Kind() RuntimeEventKind

	// At returns the timestamp when the event occurred.
	At() time.Time
}

// RuntimeEventEnvelope wraps a runtime event for cross-process transport.
type RuntimeEventEnvelope struct {
	Kind RuntimeEventKind `json:"kind"`

	Payload json.RawMessage `json:"payload"`
}

// NewRuntimeEventEnvelope builds an envelope for the provided runtime event.
func NewRuntimeEventEnvelope(event RuntimeEvent) (RuntimeEventEnvelope, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return RuntimeEventEnvelope{}, fmt.Errorf("marshal runtime event: %w", err)
	}

	return RuntimeEventEnvelope{
		Kind:    event.Kind(),
		Payload: json.RawMessage(payload),
	}, nil
}

// RuntimeStartedEvent describes a runtime start event.
type RuntimeStartedEvent struct {
	Timestamp time.Time `json:"timestamp"`
}

// Kind returns the runtime event kind.
func (RuntimeStartedEvent) Kind() RuntimeEventKind {
	return RuntimeEventStarted
}

// At returns the timestamp when the event occurred.
func (event RuntimeStartedEvent) At() time.Time {
	return event.Timestamp
}

// RuntimeFinishedEvent describes a runtime finish event.
type RuntimeFinishedEvent struct {
	Timestamp time.Time `json:"timestamp"`
}

// Kind returns the runtime event kind.
func (RuntimeFinishedEvent) Kind() RuntimeEventKind {
	return RuntimeEventFinished
}

// At returns the timestamp when the event occurred.
func (event RuntimeFinishedEvent) At() time.Time {
	return event.Timestamp
}

// RuntimeRegistry provides access to available runtimes.
type RuntimeRegistry interface {
	// Get returns the runtime implementation for the provided kind.
	// Returns ErrRuntimeNotFound when the kind is not registered.
	Get(ctx context.Context, kind RuntimeKind) (Runtime, error)

	// List returns the kinds supported by the registry.
	List(ctx context.Context) ([]RuntimeKind, error)
}

// Runtime describes a runtime implementation that can execute agent workloads.
type Runtime interface {
	// Execute starts a runtime instance for the given configuration and input.
	Execute(ctx context.Context, id ExecutionID, input ExecutionInput, config Config) (RuntimeInstance, error)

	// Discovery checks whether the runtime is available on the system.
	Discovery(ctx context.Context) (bool, error)

	// GetDefaultConfig returns the default configuration for the runtime.
	GetDefaultConfig(ctx context.Context) (RuntimeConfig, error)

	// GetDefaultFeatures returns list of default features supported by the runtime.
	GetDefaultFeatures(ctx context.Context) (RuntimeFeatures, error)

	// GetInfo returns metadata about the runtime implementation.
	GetInfo(ctx context.Context) (RuntimeInfo, error)

	AddMCPServer(ctx context.Context, mcpServerName RuntimeMCPServerName, mcpServer RuntimeMCPServer) error

	ListMCPServers(ctx context.Context) (map[RuntimeMCPServerName]RuntimeMCPServer, error)

	RemoveMCPServer(ctx context.Context, mcpServerName RuntimeMCPServerName) error
}

type RuntimeMCPServerName string

type RuntimeSTDIOMCPServer struct {
	Command   string
	Arguments []string
}

type RuntimeMCPServer struct {
	STDIO *RuntimeSTDIOMCPServer
}

// RuntimeInfo describes runtime metadata.
type RuntimeInfo struct {
	Version string `json:"version"`
}

// RuntimeResult captures the output of a runtime instance.
type RuntimeResult struct {
	Response string `json:"response,omitempty"`

	ConversationID ConversationID `json:"conversationId,omitempty"`
}

// RuntimeExecutionError reports a runtime execution failure.
type RuntimeExecutionError struct {
	// Message is the user-facing description of the failure.
	Message string

	// ExitCode is the process exit code when available.
	ExitCode *int

	// Cause is the underlying error, when available.
	Cause error
}

// Error returns the error message for the runtime execution error.
func (err *RuntimeExecutionError) Error() string {
	if err == nil {
		return ""
	}

	if err.Message != "" {
		return err.Message
	}

	if err.Cause != nil {
		return err.Cause.Error()
	}

	return "runtime execution error"
}

// Unwrap returns the underlying error, if any.
func (err *RuntimeExecutionError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Cause
}

// RuntimeInstance represents a running runtime process.
type RuntimeInstance interface {
	// Events returns a channel of runtime events emitted by the instance.
	Events() <-chan RuntimeEvent

	// Wait blocks until the runtime instance completes.
	// Returns RuntimeExecutionError when the runtime execution fails.
	Wait(ctx context.Context) (RuntimeResult, error)
}

// ErrRuntimeNotFound indicates the requested runtime is not registered.
var ErrRuntimeNotFound = fmt.Errorf("runtime not found")
