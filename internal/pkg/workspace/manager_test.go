package workspace

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubProvider struct {
	result ProvisionResult
	err    error
}

func (s *stubProvider) Provision(_ context.Context, _ url.URL) (ProvisionResult, error) {
	return s.result, s.err
}

func TestManager_Schemes_WhenNoProviders_ThenReturnsEmpty(t *testing.T) {
	manager := NewManager(map[string]Provider{})

	schemes := manager.Schemes()

	assert.Empty(t, schemes)
}

func TestManager_Schemes_WhenMultipleProviders_ThenReturnsAllSchemes(t *testing.T) {
	manager := NewManager(map[string]Provider{
		"dir": &stubProvider{},
		"git": &stubProvider{},
	})

	schemes := manager.Schemes()

	assert.ElementsMatch(t, []string{"dir", "git"}, schemes)
}

func TestManager_Provision_WhenInvalidURI_ThenReturnsError(t *testing.T) {
	manager := NewManager(map[string]Provider{})

	_, err := manager.Provision(context.Background(), "://bad")

	require.Error(t, err)
}

func TestManager_Provision_WhenUnknownScheme_ThenReturnsErrProviderNotFound(t *testing.T) {
	manager := NewManager(map[string]Provider{})

	_, err := manager.Provision(context.Background(), "unknown:///path")

	require.ErrorIs(t, err, ErrProviderNotFound)
}

func TestManager_Provision_WhenProviderFails_ThenReturnsWrappedError(t *testing.T) {
	providerErr := errors.New("provision failed")
	manager := NewManager(map[string]Provider{
		"dir": &stubProvider{err: providerErr},
	})

	_, err := manager.Provision(context.Background(), "dir:///tmp/project")

	require.Error(t, err)
	require.ErrorIs(t, err, providerErr)
}

func TestManager_Provision_WhenProviderSucceeds_ThenReturnsResult(t *testing.T) {
	expected := ProvisionResult{WorkDir: "/tmp/runs/abc123", Cleanup: func() error { return nil }}
	manager := NewManager(map[string]Provider{
		"dir": &stubProvider{result: expected},
	})

	result, err := manager.Provision(context.Background(), "dir:///tmp/project")

	require.NoError(t, err)
	assert.Equal(t, expected.WorkDir, result.WorkDir)
	assert.NotNil(t, result.Cleanup)
}
