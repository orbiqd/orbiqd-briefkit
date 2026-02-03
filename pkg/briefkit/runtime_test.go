package briefkit

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingRuntimeEvent struct{}

func (failingRuntimeEvent) Kind() RuntimeEventKind {
	return RuntimeEventStarted
}

func (failingRuntimeEvent) At() time.Time {
	return time.Time{}
}

func (failingRuntimeEvent) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func TestRuntimeEventEnvelope_WhenRuntimeStartedEvent_ThenKindAndPayloadMatch(t *testing.T) {
	now := time.Date(2024, time.March, 1, 12, 30, 0, 0, time.UTC)
	event := RuntimeStartedEvent{Timestamp: now}

	envelope, err := NewRuntimeEventEnvelope(event)
	require.NoError(t, err)
	assert.Equal(t, RuntimeEventStarted, envelope.Kind)

	var decoded RuntimeStartedEvent
	err = json.Unmarshal(envelope.Payload, &decoded)
	require.NoError(t, err)
	assert.Equal(t, now, decoded.Timestamp)
}

func TestRuntimeEventEnvelope_WhenRuntimeFinishedEvent_ThenKindAndPayloadMatch(t *testing.T) {
	now := time.Date(2024, time.April, 10, 9, 15, 0, 0, time.UTC)
	event := RuntimeFinishedEvent{Timestamp: now}

	envelope, err := NewRuntimeEventEnvelope(event)
	require.NoError(t, err)
	assert.Equal(t, RuntimeEventFinished, envelope.Kind)

	var decoded RuntimeFinishedEvent
	err = json.Unmarshal(envelope.Payload, &decoded)
	require.NoError(t, err)
	assert.Equal(t, now, decoded.Timestamp)
}

func TestRuntimeEventEnvelope_WhenMarshalFails_ThenReturnsError(t *testing.T) {
	_, err := NewRuntimeEventEnvelope(failingRuntimeEvent{})
	require.Error(t, err)
	require.ErrorContains(t, err, "marshal runtime event")
}

func TestRuntimeStartedEvent_WhenKindCalled_ThenReturnsRuntimeEventStarted(t *testing.T) {
	event := RuntimeStartedEvent{}

	kind := event.Kind()

	assert.Equal(t, RuntimeEventStarted, kind)
}

func TestRuntimeStartedEvent_WhenAtCalled_ThenReturnsTimestamp(t *testing.T) {
	now := time.Date(2024, time.May, 5, 8, 0, 0, 0, time.UTC)
	event := RuntimeStartedEvent{Timestamp: now}

	at := event.At()

	assert.Equal(t, now, at)
}

func TestRuntimeFinishedEvent_WhenKindCalled_ThenReturnsRuntimeEventFinished(t *testing.T) {
	event := RuntimeFinishedEvent{}

	kind := event.Kind()

	assert.Equal(t, RuntimeEventFinished, kind)
}

func TestRuntimeFinishedEvent_WhenAtCalled_ThenReturnsTimestamp(t *testing.T) {
	now := time.Date(2024, time.June, 12, 15, 45, 0, 0, time.UTC)
	event := RuntimeFinishedEvent{Timestamp: now}

	at := event.At()

	assert.Equal(t, now, at)
}

func TestRuntimeExecutionError_WhenNil_ThenErrorReturnsEmptyString(t *testing.T) {
	var runtimeErr *RuntimeExecutionError

	message := runtimeErr.Error()

	assert.Empty(t, message)
}

func TestRuntimeExecutionError_WhenMessageSet_ThenErrorReturnsMessage(t *testing.T) {
	runtimeErr := &RuntimeExecutionError{Message: "custom message"}

	message := runtimeErr.Error()

	assert.Equal(t, "custom message", message)
}

func TestRuntimeExecutionError_WhenOnlyCauseSet_ThenErrorReturnsCauseMessage(t *testing.T) {
	cause := errors.New("root cause")
	runtimeErr := &RuntimeExecutionError{Cause: cause}

	message := runtimeErr.Error()

	assert.Equal(t, "root cause", message)
}

func TestRuntimeExecutionError_WhenMessageAndCauseSet_ThenErrorReturnsMessage(t *testing.T) {
	cause := errors.New("root cause")
	runtimeErr := &RuntimeExecutionError{Message: "custom message", Cause: cause}

	message := runtimeErr.Error()

	assert.Equal(t, "custom message", message)
}

func TestRuntimeExecutionError_WhenNoMessageAndNoCause_ThenErrorReturnsDefault(t *testing.T) {
	runtimeErr := &RuntimeExecutionError{}

	message := runtimeErr.Error()

	assert.Equal(t, "runtime execution error", message)
}

func TestRuntimeExecutionError_Unwrap_WhenNil_ThenReturnsNil(t *testing.T) {
	var runtimeErr *RuntimeExecutionError

	cause := runtimeErr.Unwrap()

	assert.NoError(t, cause)
}

func TestRuntimeExecutionError_Unwrap_WhenCauseSet_ThenReturnsCause(t *testing.T) {
	cause := errors.New("root cause")
	runtimeErr := &RuntimeExecutionError{Cause: cause}

	unwrapped := runtimeErr.Unwrap()

	assert.Equal(t, cause, unwrapped)
}

func TestRuntimeExecutionError_Unwrap_WhenNoCause_ThenReturnsNil(t *testing.T) {
	runtimeErr := &RuntimeExecutionError{}

	unwrapped := runtimeErr.Unwrap()

	assert.NoError(t, unwrapped)
}
