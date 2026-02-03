package briefkit

import "context"

type Runner interface {
	Spawn(ctx context.Context, executionId ExecutionID) error
	Wait(ctx context.Context, executionId ExecutionID) error
}
