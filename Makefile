.PHONY: build build-local build-release build-all build-mocks lint lint-go lint-goreleaser test clean run-briefkit-runner debug-briefkit-mcp setup

# Local build (current platform via go build)
build: build-local

build-local: lint-go
	mkdir -p bin
	go build -o bin/briefkit-ctl ./cmd/briefkit-ctl
	go build -o bin/briefkit-mcp ./cmd/briefkit-mcp
	go build -o bin/briefkit-runner ./cmd/briefkit-runner

# Release build (all platforms via GoReleaser)
build-release: lint
	goreleaser build --snapshot --clean

# Backwards-compatible alias
build-all: build-release

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
run-briefkit-runner: build-local
	./bin/briefkit-runner --log-level=debug --retry $(filter-out $@,$(MAKECMDGOALS))

debug-briefkit-mcp: build-local
	DANGEROUSLY_OMIT_AUTH=true npx @modelcontextprotocol/inspector ./bin/briefkit-mcp

setup: build
	./bin/briefkit-ctl setup --force

# Clean
clean:
	rm -rf bin/ dist/

%:
	@:
