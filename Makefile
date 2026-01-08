.PHONY: build build-mocks build-briefkit-runner build-briefkit-ctl build-briefkit-mcp build-runner run-briefkit-runner debug-briefkit-mcp test

BIN_DIR := bin

build: build-briefkit-runner build-briefkit-ctl build-briefkit-mcp

build-mocks: build-claude-mock build-codex-mock build-gemini-mock

build-claude-mock:
	go build -o test/runtime/claude/claude-mock ./test/runtime/claude/claude-mock.go

build-codex-mock:
	go build -o test/runtime/codex/codex-mock ./test/runtime/codex/codex-mock.go

build-gemini-mock:
	go build -o test/runtime/gemini/gemini-mock ./test/runtime/gemini/gemini-mock.go

test: build-mocks
	go test -coverprofile=coverage.out ./...

build-briefkit-runner:
	go build -o $(BIN_DIR)/briefkit-runner ./cmd/briefkit-runner/main.go

build-runner: build-briefkit-runner

build-briefkit-ctl:
	go build -o $(BIN_DIR)/briefkit-ctl ./cmd/briefkit-ctl/main.go

build-briefkit-mcp:
	go build -o $(BIN_DIR)/briefkit-mcp ./cmd/briefkit-mcp/main.go

run-briefkit-runner: build-runner
	$(BIN_DIR)/briefkit-runner --log-level=debug --retry $(filter-out $@,$(MAKECMDGOALS))

debug-briefkit-mcp: build
	DANGEROUSLY_OMIT_AUTH=true npx @modelcontextprotocol/inspector ./bin/briefkit-mcp

%:
	@:
