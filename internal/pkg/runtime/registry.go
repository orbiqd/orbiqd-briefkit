package runtime

import (
	"context"
	"sort"

	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime/claude"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime/codex"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/runtime/gemini"
	"github.com/orbiqd/orbiqd-briefkit/pkg/briefkit"
)

type Registry struct {
	runtime map[briefkit.RuntimeKind]briefkit.Runtime
}

func NewRegistry() *Registry {
	return &Registry{
		runtime: map[briefkit.RuntimeKind]briefkit.Runtime{
			gemini.Gemini: gemini.NewRuntime(),
			claude.Claude: claude.NewRuntime(),
			codex.Codex:   codex.NewRuntime(),
		},
	}
}

func (registry Registry) Get(ctx context.Context, kind briefkit.RuntimeKind) (briefkit.Runtime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	runtime, ok := registry.runtime[kind]
	if !ok {
		return nil, briefkit.ErrRuntimeNotFound
	}

	return runtime, nil
}

func (registry Registry) List(ctx context.Context) ([]briefkit.RuntimeKind, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	kinds := make([]briefkit.RuntimeKind, 0, len(registry.runtime))
	for kind := range registry.runtime {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		kinds = append(kinds, kind)
	}

	sort.Slice(kinds, func(i, j int) bool {
		return kinds[i] < kinds[j]
	})

	return kinds, nil
}
