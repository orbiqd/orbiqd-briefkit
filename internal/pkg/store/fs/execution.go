package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
	"github.com/spf13/afero"
)

const (
	executionAgentConfigFileName = "agent.json"
	executionInputFileName       = "input.json"
	executionResultFileName      = "result.json"
	executionStatusFileName      = "status.json"
)

// Repository is an implementation of agent.ExecutionRepository that stores
// execution data on the file system.
type Repository struct {
	basePath string
	fs       afero.Fs
}

// NewExecutionRepository creates a new file system-based execution repository and
// ensures the base path exists.
func NewExecutionRepository(basePath string, fs afero.Fs) (*Repository, error) {
	if err := fs.MkdirAll(basePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("create execution repository path: %w", err)
	}

	return &Repository{
		basePath: basePath,
		fs:       fs,
	}, nil
}

// Create persists a new execution and returns its identifier.
func (r *Repository) Create(ctx context.Context, input briefkit.ExecutionInput, agentConfig briefkit.Config) (briefkit.ExecutionID, error) {
	if err := input.Validate(); err != nil {
		return briefkit.EmptyExecutionID, err
	}

	id := briefkit.NewExecutionID()
	executionPath := filepath.Join(r.basePath, string(id))

	if err := r.fs.MkdirAll(executionPath, os.ModePerm); err != nil { // os.ModePerm needed
		return briefkit.EmptyExecutionID, err
	}

	inputFilePath := filepath.Join(executionPath, executionInputFileName)
	if err := writeJSON(r.fs, inputFilePath, input); err != nil {
		return briefkit.EmptyExecutionID, err
	}

	agentConfigFilePath := filepath.Join(executionPath, executionAgentConfigFileName)
	if err := writeJSON(r.fs, agentConfigFilePath, agentConfig); err != nil {
		return briefkit.EmptyExecutionID, err
	}

	now := time.Now()
	status := briefkit.ExecutionStatus{
		CreatedAt: now,
		UpdatedAt: now,
		State:     briefkit.ExecutionCreated,
		Attempts:  0,
	}
	statusFilePath := filepath.Join(executionPath, executionStatusFileName)
	if err := writeJSON(r.fs, statusFilePath, status); err != nil {
		return briefkit.EmptyExecutionID, err
	}

	return id, nil
}

// Exists reports whether an execution with the given identifier exists.
func (r *Repository) Exists(ctx context.Context, id briefkit.ExecutionID) (bool, error) {
	if err := id.Validate(); err != nil {
		return false, err
	}

	executionPath := filepath.Join(r.basePath, string(id))
	return afero.Exists(r.fs, executionPath)
}

// Get loads the execution handle for the given identifier.
func (r *Repository) Get(ctx context.Context, id briefkit.ExecutionID) (briefkit.Execution, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}

	exists, err := r.Exists(ctx, id)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, briefkit.ErrExecutionNotFound
	}

	return &Execution{
		id:       id,
		basePath: r.basePath,
		fs:       r.fs,
	}, nil
}

// Find returns execution identifiers matching the provided filters.
func (r *Repository) Find(ctx context.Context, filters ...briefkit.ExecutionFilter) ([]briefkit.ExecutionID, error) {
	_ = filters

	entries, err := afero.ReadDir(r.fs, r.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []briefkit.ExecutionID{}, nil
		}
		return nil, err
	}

	ids := make([]briefkit.ExecutionID, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if !entry.IsDir() {
			continue
		}

		id := briefkit.ExecutionID(entry.Name())
		if err := id.Validate(); err != nil {
			continue
		}

		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})

	return ids, nil
}

// Execution is an implementation of agent.Execution for the file system store.
type Execution struct {
	id       briefkit.ExecutionID
	basePath string
	fs       afero.Fs
}

func (e *Execution) executionDirPath() string {
	return filepath.Join(e.basePath, string(e.id))
}

func (e *Execution) inputFilePath() string {
	return filepath.Join(e.executionDirPath(), executionInputFileName)
}

func (e *Execution) agentConfigFilePath() string {
	return filepath.Join(e.executionDirPath(), executionAgentConfigFileName)
}

func (e *Execution) resultFilePath() string {
	return filepath.Join(e.executionDirPath(), executionResultFileName)
}

func (e *Execution) statusFilePath() string {
	return filepath.Join(e.executionDirPath(), executionStatusFileName)
}

// GetInput returns the stored input for the execution.
func (e *Execution) GetInput(ctx context.Context) (briefkit.ExecutionInput, error) {
	return readJSON[briefkit.ExecutionInput](e.fs, e.inputFilePath())
}

// GetAgentConfig returns the stored agent config for the execution.
func (e *Execution) GetAgentConfig(ctx context.Context) (briefkit.Config, error) {
	exists, err := hasJSON(e.fs, e.agentConfigFilePath())
	if err != nil {
		return briefkit.Config{}, err
	}

	if !exists {
		return briefkit.Config{}, briefkit.ErrExecutionAgentConfigNotFound
	}

	return readJSON[briefkit.Config](e.fs, e.agentConfigFilePath())
}

// GetResult returns the stored result for the execution.
func (e *Execution) GetResult(ctx context.Context) (briefkit.ExecutionResult, error) {
	exists, err := hasJSON(e.fs, e.resultFilePath())
	if err != nil {
		return briefkit.ExecutionResult{}, err
	}

	if !exists {
		return briefkit.ExecutionResult{}, briefkit.ErrExecutionNoResult
	}

	return readJSON[briefkit.ExecutionResult](e.fs, e.resultFilePath())
}

// HasResult reports whether the execution has a stored result.
func (e *Execution) HasResult(ctx context.Context) (bool, error) {
	return hasJSON(e.fs, e.resultFilePath())
}

// SetResult stores the result for the execution.
func (e *Execution) SetResult(ctx context.Context, result briefkit.ExecutionResult) error {
	status, err := e.GetStatus(ctx)
	if err != nil {
		return err
	}

	if err := writeJSON(e.fs, e.resultFilePath(), result); err != nil {
		return err
	}

	now := time.Now()
	status.State = briefkit.ExecutionSucceeded
	status.FinishedAt = &now
	status.UpdatedAt = now
	if err := writeJSON(e.fs, e.statusFilePath(), status); err != nil {
		return err
	}

	return nil
}

// GetStatus returns the lifecycle status for the execution.
func (e *Execution) GetStatus(ctx context.Context) (briefkit.ExecutionStatus, error) {
	return readJSON[briefkit.ExecutionStatus](e.fs, e.statusFilePath())
}

// UpdateStatus stores the lifecycle status for the execution.
func (e *Execution) UpdateStatus(ctx context.Context, status briefkit.ExecutionStatus) error {
	status.UpdatedAt = time.Now()
	if err := writeJSON(e.fs, e.statusFilePath(), status); err != nil {
		return err
	}

	return nil
}
