package agent

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/utils"
)

// ExecutionID identifies a single execution in the store.
type ExecutionID string

// ExecutionState describes the lifecycle state of an execution.
type ExecutionState string

const (
	// ExecutionCreated indicates the execution has been created but not yet started.
	ExecutionCreated ExecutionState = "created"

	// ExecutionStarted indicates the execution has been started.
	ExecutionStarted ExecutionState = "started"

	// ExecutionRunning indicates the execution is actively running.
	ExecutionRunning ExecutionState = "running"

	// ExecutionSucceeded indicates the execution has finished successfully.
	ExecutionSucceeded ExecutionState = "succeeded"

	// ExecutionFailed indicates the execution has finished with an error.
	ExecutionFailed ExecutionState = "failed"
)

// ExecutionInput captures the runtime input required to run an execution.
type ExecutionInput struct {
	// WorkingDirectory is the filesystem path where the execution runs.
	// When nil, the runtime uses the current working directory.
	WorkingDirectory *string `json:"workingDirectory"`

	// Timeout defines the maximum allowed duration for the execution.
	Timeout utils.Duration `json:"timeout"`

	// Prompt is the user input sent to the agent.
	Prompt string `json:"prompt"`

	Model *string `json:"model,omitempty"`

	// ConversationID continues an existing agent conversation when provided.
	ConversationID *ConversationID `json:"conversationId,omitempty"`

	// Attachments lists optional files supplied with the prompt.
	Attachments []ExecutionInputAttachment `json:"attachments,omitempty"`
}

// ExecutionInputAttachment describes a single file attached to the execution input.
type ExecutionInputAttachment struct {
	// MimeType is the media type of the attached file.
	MimeType string `json:"mimeType"`

	// Path is the filesystem path to the attached file.
	Path string `json:"path"`
}

// ExecutionResult captures the outcome of an execution.
type ExecutionResult struct {
	// ConversationID reports the conversation identifier returned by the agent.
	ConversationID ConversationID `json:"conversationId,omitempty"`

	// Response carries the final agent response text.
	Response string `json:"response"`
}

// ExecutionStatus tracks lifecycle timestamps and state for an execution.
type ExecutionStatus struct {
	// CreatedAt is the timestamp when the execution was created.
	CreatedAt time.Time `json:"createdAt"`

	// Attempts is the number of times the execution has been started.
	Attempts int `json:"attempts"`

	// UpdatedAt is the timestamp of the most recent status update.
	UpdatedAt time.Time `json:"updatedAt"`

	// FinishedAt is the timestamp when the execution finished.
	FinishedAt *time.Time `json:"finishedAt,omitempty"`

	// State is the current lifecycle state of the execution.
	State ExecutionState `json:"state"`

	// ExitCode reports the process exit code when the execution completes.
	ExitCode *int `json:"exitCode,omitempty"`

	// Error carries a runtime error message when the execution fails.
	Error *string `json:"error,omitempty"`
}

// ExecutionQuery describes filters used to locate executions in a repository.
type ExecutionQuery struct {
}

// ExecutionFilter applies a filter to an execution query.
type ExecutionFilter func(query *ExecutionQuery)

// ExecutionRepository provides access to execution handles in a store.
type ExecutionRepository interface {
	// Create persists a new execution and returns its identifier.
	// Returns ErrExecutionPromptRequired when the input prompt is missing.
	// Returns ErrExecutionTimeoutRequired when the input timeout is missing.
	// Returns ErrExecutionWorkingDirectoryRequired when the input working directory is empty.
	// Returns ErrExecutionWorkingDirectoryInvalid when the input working directory is malformed.
	// Returns ErrExecutionWorkingDirectoryNotAbsolute when the input working directory is not absolute.
	// Returns ErrExecutionAttachmentMimeTypeRequired when an attachment MIME type is missing.
	// Returns ErrExecutionAttachmentPathRequired when an attachment path is missing.
	Create(ctx context.Context, input ExecutionInput, agentConfig Config) (ExecutionID, error)

	// Exists reports whether an execution with the given identifier exists.
	// Returns ErrExecutionIDInvalid when the identifier is missing or malformed.
	Exists(ctx context.Context, id ExecutionID) (bool, error)

	// Get loads the execution handle for the given identifier.
	// Returns ErrExecutionNotFound when the execution does not exist.
	// Returns ErrExecutionIDInvalid when the identifier is missing or malformed.
	Get(ctx context.Context, id ExecutionID) (Execution, error)

	// Find returns execution identifiers matching the provided filters.
	Find(ctx context.Context, filters ...ExecutionFilter) ([]ExecutionID, error)
}

// Execution encapsulates per-execution state accessors and updates.
type Execution interface {
	// GetInput returns the stored input for the execution.
	// Returns ErrExecutionNotFound when the execution does not exist.
	GetInput(ctx context.Context) (ExecutionInput, error)

	// GetAgentConfig returns the stored agent configuration snapshot for the execution.
	// Returns ErrExecutionNotFound when the execution does not exist.
	// Returns ErrExecutionAgentConfigNotFound when the agent config snapshot is missing.
	GetAgentConfig(ctx context.Context) (Config, error)

	// GetResult returns the stored result for the execution.
	// Returns ErrExecutionNotFound when the execution does not exist.
	// Returns ErrExecutionNoResult when the execution has no result yet.
	GetResult(ctx context.Context) (ExecutionResult, error)

	// HasResult reports whether the execution has a stored result.
	// Returns ErrExecutionNotFound when the execution does not exist.
	HasResult(ctx context.Context) (bool, error)

	// SetResult stores the result for the execution.
	// Returns ErrExecutionNotFound when the execution does not exist.
	SetResult(ctx context.Context, result ExecutionResult) error

	// GetStatus returns the lifecycle status for the execution.
	// Returns ErrExecutionNotFound when the execution does not exist.
	GetStatus(ctx context.Context) (ExecutionStatus, error)

	// UpdateStatus stores the lifecycle status for the execution.
	// Returns ErrExecutionNotFound when the execution does not exist.
	UpdateStatus(ctx context.Context, status ExecutionStatus) error
}

// NewExecutionID generates a new execution identifier.
func NewExecutionID() ExecutionID {
	return ExecutionID(uuid.NewString())
}

// Validate checks whether the execution identifier is non-empty and a valid UUID.
func (id ExecutionID) Validate() error {
	if id == "" {
		return ErrExecutionIDInvalid
	}

	_, err := uuid.Parse(string(id))
	if err != nil {
		return ErrExecutionIDInvalid
	}

	return nil
}

// IsFinished reports whether the execution state is terminal.
func (state ExecutionState) IsFinished() bool {
	return state == ExecutionSucceeded || state == ExecutionFailed
}

// Validate checks whether the attachment contains the required metadata.
func (attachment ExecutionInputAttachment) Validate() error {
	if strings.TrimSpace(attachment.MimeType) == "" {
		return ErrExecutionAttachmentMimeTypeRequired
	}

	if strings.TrimSpace(attachment.Path) == "" {
		return ErrExecutionAttachmentPathRequired
	}

	return nil
}

// Validate checks whether the execution input is complete and well-formed.
func (input ExecutionInput) Validate() error {
	if strings.TrimSpace(input.Prompt) == "" {
		return ErrExecutionPromptRequired
	}

	if input.Timeout <= 0 {
		return ErrExecutionTimeoutRequired
	}

	if input.WorkingDirectory != nil {
		if strings.TrimSpace(*input.WorkingDirectory) == "" {
			return ErrExecutionWorkingDirectoryRequired
		}

		expanded, err := homedir.Expand(*input.WorkingDirectory)
		if err != nil {
			return ErrExecutionWorkingDirectoryInvalid
		}

		if !filepath.IsAbs(expanded) {
			return ErrExecutionWorkingDirectoryNotAbsolute
		}
	}

	for _, attachment := range input.Attachments {
		if err := attachment.Validate(); err != nil {
			return err
		}
	}

	return nil
}

var (
	// EmptyExecutionID is a zero-value ExecutionID.
	EmptyExecutionID ExecutionID = ""

	// ErrExecutionNotFound indicates the execution does not exist in the repository.
	ErrExecutionNotFound = errors.New("execution not found")

	// ErrExecutionNoResult indicates the execution exists but has no stored result yet.
	ErrExecutionNoResult = errors.New("execution result not found")

	// ErrExecutionAgentConfigNotFound indicates the execution exists but has no stored agent config yet.
	ErrExecutionAgentConfigNotFound = errors.New("execution agent config not found")

	// ErrExecutionIDInvalid indicates the execution identifier is missing or malformed.
	ErrExecutionIDInvalid = errors.New("execution id invalid")

	// ErrExecutionPromptRequired indicates the execution prompt is missing.
	ErrExecutionPromptRequired = errors.New("execution prompt required")

	// ErrExecutionTimeoutRequired indicates the execution timeout is missing or invalid.
	ErrExecutionTimeoutRequired = errors.New("execution timeout required")

	// ErrExecutionWorkingDirectoryRequired indicates the execution working directory is empty.
	ErrExecutionWorkingDirectoryRequired = errors.New("execution working directory required")

	// ErrExecutionWorkingDirectoryInvalid indicates the execution working directory is malformed.
	ErrExecutionWorkingDirectoryInvalid = errors.New("execution working directory invalid")

	// ErrExecutionWorkingDirectoryNotAbsolute indicates the execution working directory is not absolute.
	ErrExecutionWorkingDirectoryNotAbsolute = errors.New("execution working directory not absolute")

	// ErrExecutionAttachmentMimeTypeRequired indicates the attachment MIME type is missing.
	ErrExecutionAttachmentMimeTypeRequired = errors.New("execution attachment mime type required")

	// ErrExecutionAttachmentPathRequired indicates the attachment path is missing.
	ErrExecutionAttachmentPathRequired = errors.New("execution attachment path required")
)
