package workspace

import (
	"context"
	"net/url"
)

// Provider provisions an isolated working directory from a workspace source URI.
type Provider interface {
	// Provision creates an isolated copy of the workspace and returns the result.
	Provision(ctx context.Context, uri url.URL) (ProvisionResult, error)
}

// ProvisionResult holds the outcome of a successful workspace provisioning.
type ProvisionResult struct {
	// WorkDir is the absolute path to the provisioned workspace directory.
	WorkDir string

	// Cleanup removes the provisioned workspace directory.
	// Cleanup errors are non-fatal and should be logged, not propagated.
	Cleanup func() error
}
