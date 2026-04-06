package workspace

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Manager selects a workspace provider by URI scheme and delegates provisioning.
type Manager struct {
	providers map[string]Provider
}

// NewManager creates a Manager with the given scheme-to-provider mapping.
func NewManager(providers map[string]Provider) *Manager {
	return &Manager{providers: providers}
}

// Schemes returns the URI schemes supported by registered providers (e.g. ["dir"]).
func (m *Manager) Schemes() []string {
	schemes := make([]string, 0, len(m.providers))
	for scheme := range m.providers {
		schemes = append(schemes, scheme)
	}
	return schemes
}

// Provision parses rawURI, selects a provider by scheme, and provisions an isolated workspace.
// Returns ErrProviderNotFound when no provider is registered for the URI scheme.
func (m *Manager) Provision(ctx context.Context, rawURI string) (ProvisionResult, error) {
	parsed, err := url.Parse(rawURI)
	if err != nil {
		return ProvisionResult{}, fmt.Errorf("workspace URI parse: %w", err)
	}

	provider, ok := m.providers[parsed.Scheme]
	if !ok {
		return ProvisionResult{}, fmt.Errorf("%w: %s", ErrProviderNotFound, parsed.Scheme)
	}

	result, err := provider.Provision(ctx, *parsed)
	if err != nil {
		return ProvisionResult{}, fmt.Errorf("workspace provision: %w", err)
	}

	return result, nil
}

var (
	// ErrProviderNotFound indicates no provider is registered for the requested URI scheme.
	ErrProviderNotFound = errors.New("workspace provider not found")
)
