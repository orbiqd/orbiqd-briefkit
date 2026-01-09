# Development Guide

This document provides instructions for setting up the development environment, running tests, and contributing to the BriefKit project. It also outlines the core architectural concepts to help developers and AI agents understand the system's design.

## 1. Getting Started

Follow these steps to get the project running on your local machine.

### Prerequisites

- Go (version 1.22 or later)
- Basic familiarity with the command line

### Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/orbiqd/orbiqd-briefkit
    cd orbiqd-briefkit
    ```

2.  **Install dependencies:**
    Run `go mod tidy` to ensure all Go modules are correctly downloaded and verified.
    ```bash
    go mod tidy
    ```

3.  **Build the project:**
    Compile the application binaries with Makefile targets.
    ```bash
    make build
    ```

4.  **Run the application:**
    Execute the compiled binary.
    ```bash
    ./bin/briefkit-ctl
    ```

### Running Tests

To run the entire test suite, use the following command:
```bash
go test ./...
```

## 2. Core Concepts

To understand how BriefKit works, it's essential to grasp these fundamental domain concepts.

-   **Session**: A shared collaboration context that groups multiple `Turns` across same agent. It represents a continuous conversation or task.
-   **Turn**: A single user-to-agent exchange. It contains the user's request and the agent's final response, stored chronologically in the `Session` transcript.
-   **Execution**: The technical runtime that processes a `Turn`. It takes a request, orchestrates the necessary logic (e.g., running an agent), and records its status and results. Executions are ephemeral and perform the work of a `Turn`.
-   **File System Abstraction**: To facilitate easier testing and provide a consistent interface for file operations, we utilize the `spf13/afero` library. This allows us to work with various file systems (e.g., in-memory, OS-level) through a unified API.

## 3. Project Structure

The project follows the standard Go project layout.

-   `/cmd/briefkit-ctl`: Main application entry point for the CLI.
-   `/internal/app/briefkit-ctl`: Core application logic, business rules, and services.
-   `/internal/pkg/agent`: Packages related to agent definition, configuration, and execution.
-   `/go.mod`, `/go.sum`: Go module dependency management files.

### State and Data Storage

BriefKit is designed to be stateless at the application level. All state is persisted to the local filesystem to ensure data integrity and allow for observability.

-   **Single Source of Truth**: The application state is stored by default in `~/.orbiqd/briefkit`. This location can be overridden via configuration.
-   **Configuration**:
    -   Global config: `~/.orbiqd/briefkit/config.yaml`
    -   Agent definitions: `~/.orbiqd/briefkit/agents/*.yaml` (the agent ID is derived from the filename).
-   **Runtime State Directory**: The `state/` subdirectory contains data related to ongoing and completed operations. This structure allows different processes to communicate and share results reliably.
    ```
    ~/.orbiqd/briefkit/state/
      executions/<execution-id>/
        input.json
        result.json
        status.json
      turns/<turn-id>/
        request.json
        response.json
        status.json
      sessions/<session-id>/
        transcript.ndjson
    ```

## 4. Development Workflow & Conventions

Please adhere to the following rules when contributing.

-   **Code Formatting**: All Go code must be formatted with `gofmt`.
-   **Dependency Management**: After adding or removing Go dependencies, always run `go mod tidy`.
-   **Builds**: After making changes, rebuild the binaries to ensure local runs use the latest code.


