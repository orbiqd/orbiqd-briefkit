package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/agent"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime/claude"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime/codex"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime/gemini"
)

// TestNewRegistry_WhenCreated_ThenRegistersDefaultRuntimes verifies default runtimes are registered.
func TestNewRegistry_WhenCreated_ThenRegistersDefaultRuntimes(t *testing.T) {
	// Arrange.
	registry := NewRegistry()

	// Act.
	runtimes := registry.runtime

	// Assert.
	require.NotNil(t, runtimes)
	assert.Len(t, runtimes, 3)
	assert.Contains(t, runtimes, claude.Claude)
	assert.Contains(t, runtimes, codex.Codex)
	assert.Contains(t, runtimes, gemini.Gemini)
}

// TestRegistry_Get_WhenKindKnown_ThenReturnsRuntime returns a runtime for a known kind.
func TestRegistry_Get_WhenKindKnown_ThenReturnsRuntime(t *testing.T) {
	// Arrange.
	registry := NewRegistry()
	ctx := context.Background()

	// Act.
	runtime, err := registry.Get(ctx, codex.Codex)

	// Assert.
	require.NoError(t, err)
	assert.NotNil(t, runtime)
}

// TestRegistry_Get_WhenKindUnknown_ThenReturnsErrRuntimeNotFound returns ErrRuntimeNotFound for unknown kind.
func TestRegistry_Get_WhenKindUnknown_ThenReturnsErrRuntimeNotFound(t *testing.T) {
	// Arrange.
	registry := NewRegistry()
	ctx := context.Background()
	kind := agent.RuntimeKind("unknown")

	// Act.
	runtime, err := registry.Get(ctx, kind)

	// Assert.
	require.ErrorIs(t, err, agent.ErrRuntimeNotFound)
	assert.Nil(t, runtime)
}

// TestRegistry_Get_WhenContextCanceled_ThenReturnsContextError returns ctx error when canceled.
func TestRegistry_Get_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	// Arrange.
	registry := NewRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act.
	runtime, err := registry.Get(ctx, codex.Codex)

	// Assert.
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, runtime)
}

// TestRegistry_List_WhenCalled_ThenReturnsSortedKinds returns sorted runtime kinds.
func TestRegistry_List_WhenCalled_ThenReturnsSortedKinds(t *testing.T) {
	// Arrange.
	registry := NewRegistry()
	ctx := context.Background()
	expected := []agent.RuntimeKind{claude.Claude, codex.Codex, gemini.Gemini}

	// Act.
	kinds, err := registry.List(ctx)

	// Assert.
	require.NoError(t, err)
	assert.Equal(t, expected, kinds)
}

// TestRegistry_List_WhenContextCanceled_ThenReturnsContextError returns ctx error when canceled.
func TestRegistry_List_WhenContextCanceled_ThenReturnsContextError(t *testing.T) {
	// Arrange.
	registry := NewRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act.
	kinds, err := registry.List(ctx)

	// Assert.
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, kinds)
}
