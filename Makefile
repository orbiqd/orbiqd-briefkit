.PHONY: build build-local build-release build-all lint lint-go lint-goreleaser lint-codecov lint-docs test clean run-briefkit-runner debug-briefkit-mcp setup generate-mocks

# Local build (current platform via go build)
build: build-local

build-local: lint-go generate-mocks
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
lint: lint-go lint-goreleaser lint-codecov lint-docs

lint-go:
	golangci-lint run --fix

lint-goreleaser:
	goreleaser check

lint-codecov:
	curl --fail --silent --show-error --data-binary @codecov.yml https://codecov.io/validate >/dev/null

lint-docs:
	vale sync
	find . -name '*.md' -not -path './.tmp/*' | xargs vale

# Test
test: lint generate-mocks
	go test -tags=coverage -coverprofile=coverage.out ./...

# Generate mocks
generate-mocks:
	mockery

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
