.PHONY: build build-all build-mocks lint lint-go lint-goreleaser test clean run-briefkit-runner debug-briefkit-mcp

# Dev build (current platform via GoReleaser)
build: lint
	goreleaser build --snapshot --clean --single-target

# Release build (all platforms)
build-all: lint
	goreleaser build --snapshot --clean

# Lint
lint: lint-go lint-goreleaser

lint-go:
	golangci-lint run --fix

lint-goreleaser:
	goreleaser check

# Test
build-mocks: build-claude-mock build-codex-mock build-gemini-mock

build-claude-mock:
	go build -o test/runtime/claude/claude-mock ./test/runtime/claude/claude-mock.go

build-codex-mock:
	go build -o test/runtime/codex/codex-mock ./test/runtime/codex/codex-mock.go

build-gemini-mock:
	go build -o test/runtime/gemini/gemini-mock ./test/runtime/gemini/gemini-mock.go

test: build-mocks lint
	go test -coverprofile=coverage.out ./...

# Run targets
run-briefkit-runner: build
	./dist/briefkit-runner_*/briefkit-runner --log-level=debug --retry $(filter-out $@,$(MAKECMDGOALS))

debug-briefkit-mcp: build
	DANGEROUSLY_OMIT_AUTH=true npx @modelcontextprotocol/inspector ./dist/briefkit-mcp_*/briefkit-mcp

# Clean
clean:
	rm -rf dist/

%:
	@:
