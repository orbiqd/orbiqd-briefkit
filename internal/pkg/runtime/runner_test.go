package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/coder/quartz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

func TestNewRunner_WhenCreated_ThenHasDefaultPollInterval(t *testing.T) {
	// Arrange.
	mockRepo := briefkit.NewMockExecutionRepository(t)

	// Act.
	runner := NewRunner(mockRepo)

	// Assert.
	assert.Equal(t, 500*time.Millisecond, runner.pollInterval)
}

func TestNewRunner_WhenCreated_ThenHasDefaultClock(t *testing.T) {
	// Arrange.
	mockRepo := briefkit.NewMockExecutionRepository(t)

	// Act.
	runner := NewRunner(mockRepo)

	// Assert.
	assert.NotNil(t, runner.clock)
}

func TestNewRunner_WithPollInterval_ThenUseCustomInterval(t *testing.T) {
	// Arrange.
	mockRepo := briefkit.NewMockExecutionRepository(t)
	customInterval := 1 * time.Second

	// Act.
	runner := NewRunner(mockRepo, WithPollInterval(customInterval))

	// Assert.
	assert.Equal(t, customInterval, runner.pollInterval)
}

func TestNewRunner_WithClock_ThenUseCustomClock(t *testing.T) {
	// Arrange.
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockClock := quartz.NewMock(t)

	// Act.
	runner := NewRunner(mockRepo, WithClock(mockClock))

	// Assert.
	assert.Equal(t, mockClock, runner.clock)
}

func TestRunner_Wait_WhenExecutionSucceeds_ThenReturnsNil(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{
		State: briefkit.ExecutionSucceeded,
	}, nil)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	mockClock.Advance(500 * time.Millisecond)

	// Assert.
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenExecutionFails_ThenReturnsError(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{
		State: briefkit.ExecutionFailed,
	}, nil)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	mockClock.Advance(500 * time.Millisecond)

	// Assert.
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, ErrExecutionFailed)
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenExecutionFailsWithErrorMessage_ThenErrorContainsMessage(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")
	errorMessage := "something went wrong"

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{
		State: briefkit.ExecutionFailed,
		Error: &errorMessage,
	}, nil)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	mockClock.Advance(500 * time.Millisecond)

	// Assert.
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, ErrExecutionFailed)
		assert.Contains(t, err.Error(), errorMessage)
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenExecutionFailsWithExitCode_ThenErrorContainsExitCode(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")
	exitCode := 1

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{
		State:    briefkit.ExecutionFailed,
		ExitCode: &exitCode,
	}, nil)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	mockClock.Advance(500 * time.Millisecond)

	// Assert.
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, ErrExecutionFailed)
		assert.Contains(t, err.Error(), "Exit code is 1")
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenExecutionFailsWithErrorAndExitCode_ThenErrorContainsBoth(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")
	errorMessage := "process crashed"
	exitCode := 137

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{
		State:    briefkit.ExecutionFailed,
		Error:    &errorMessage,
		ExitCode: &exitCode,
	}, nil)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	mockClock.Advance(500 * time.Millisecond)

	// Assert.
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, ErrExecutionFailed)
		assert.Contains(t, err.Error(), errorMessage)
		assert.Contains(t, err.Error(), "Exit code is 137")
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	// Arrange.
	ctx, cancel := context.WithCancel(context.Background())
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	cancel()

	// Assert.
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenGetStatusFails_ThenReturnsError(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")
	statusError := errors.New("status error")

	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).Return(briefkit.ExecutionStatus{}, statusError)

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))
	mockClock.Advance(500 * time.Millisecond)

	// Assert.
	select {
	case err := <-errCh:
		require.ErrorIs(t, err, statusError)
	case <-time.After(1 * time.Second):
		t.Fatal("Wait did not return in time")
	}
}

func TestRunner_Wait_WhenRepositoryGetFails_ThenReturnsError(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")
	getError := errors.New("repository error")

	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockRepo.EXPECT().Get(ctx, executionID).Return(nil, getError)

	runner := NewRunner(mockRepo)

	// Act.
	err := runner.Wait(ctx, executionID)

	// Assert.
	require.ErrorIs(t, err, getError)
}

func TestRunner_Wait_WhenExecutionNotFinished_ThenPollsUntilFinished(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")

	callCount := 0
	mockClock := quartz.NewMock(t)
	mockRepo := briefkit.NewMockExecutionRepository(t)
	mockExecution := briefkit.NewMockExecution(t)

	mockRepo.EXPECT().Get(ctx, executionID).Return(mockExecution, nil)
	mockExecution.EXPECT().GetStatus(ctx).RunAndReturn(func(_ context.Context) (briefkit.ExecutionStatus, error) {
		callCount++
		if callCount < 2 {
			return briefkit.ExecutionStatus{State: briefkit.ExecutionRunning}, nil
		}
		return briefkit.ExecutionStatus{State: briefkit.ExecutionSucceeded}, nil
	})

	runner := NewRunner(mockRepo, WithClock(mockClock))

	trap := mockClock.Trap().NewTicker()
	defer trap.Close()

	// Act.
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Wait(ctx, executionID)
	}()

	call := trap.MustWait(ctx)
	require.NoError(t, call.Release(ctx))

	for {
		select {
		case err := <-errCh:
			// Assert.
			require.NoError(t, err)
			require.Equal(t, 2, callCount)
			return
		default:
			mockClock.Advance(500 * time.Millisecond)
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestNewRunner_WhenCreated_ThenHasDefaultExecutableResolver(t *testing.T) {
	// Arrange.
	mockRepo := briefkit.NewMockExecutionRepository(t)

	// Act.
	runner := NewRunner(mockRepo)

	// Assert.
	assert.NotNil(t, runner.executableResolver)
}

func TestNewRunner_WithExecutableResolver_ThenUseCustomResolver(t *testing.T) {
	// Arrange.
	mockRepo := briefkit.NewMockExecutionRepository(t)
	customResolver := func(_ context.Context, _ string) (string, error) {
		return "/custom/path", nil
	}

	// Act.
	runner := NewRunner(mockRepo, WithExecutableResolver(customResolver))

	// Assert.
	path, err := runner.executableResolver(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "/custom/path", path)
}

func TestRunner_Spawn_WhenResolveExecutableFails_ThenReturnsError(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")
	resolveError := errors.New("executable not found")

	mockRepo := briefkit.NewMockExecutionRepository(t)
	resolver := func(_ context.Context, _ string) (string, error) {
		return "", resolveError
	}

	runner := NewRunner(mockRepo, WithExecutableResolver(resolver))

	// Act.
	err := runner.Spawn(ctx, executionID)

	// Assert.
	require.ErrorIs(t, err, resolveError)
	assert.Contains(t, err.Error(), "resolve executable")
}

func TestRunner_Spawn_WhenSuccessful_ThenStartsProcess(t *testing.T) {
	// Arrange.
	ctx := context.Background()
	executionID := briefkit.ExecutionID("550e8400-e29b-41d4-a716-446655440000")

	mockRepo := briefkit.NewMockExecutionRepository(t)
	resolver := func(_ context.Context, _ string) (string, error) {
		return "/usr/bin/true", nil
	}

	runner := NewRunner(mockRepo, WithExecutableResolver(resolver))

	// Act.
	err := runner.Spawn(ctx, executionID)

	// Assert.
	require.NoError(t, err)
}
