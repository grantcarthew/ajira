# p-002: CLI Core Infrastructure

- Status: Completed
- Started: 2025-12-22
- Completed: 2025-12-22

## Overview

Establish the foundational infrastructure for ajira: project structure, configuration, API client, and initial commands. This project creates the skeleton upon which all subsequent features are built.

## Goals

1. Set up Go project structure following standard layout conventions
2. Implement environment-based configuration
3. Create Jira REST API v3 client with authentication
4. Build Cobra CLI framework with root command and global flags
5. Implement `ajira me` command to validate end-to-end functionality
6. Implement `ajira project list` command

## Scope

In Scope:

- Go module initialization and directory structure
- Configuration package reading environment variables
- HTTP client with Basic Auth (email + API token)
- Root command with `--json`, `--project`, `--version`, `--help` flags
- `ajira me` command (GET /rest/api/3/myself)
- `ajira project list` command (GET /rest/api/3/project/search)
- Plain text and JSON output formatting
- Error handling with clear messages and exit codes

Out of Scope:

- Issue commands (p-003)
- Markdown/ADF conversion (p-004)
- Comment functionality (p-005)

## Success Criteria

- [x] `go build ./cmd/ajira` compiles without errors
- [x] `ajira --version` displays version information
- [x] `ajira me` returns current user info (plain text)
- [x] `ajira me -j` returns current user info (JSON)
- [x] `ajira project list` returns accessible projects
- [x] Missing environment variables produce clear error messages
- [x] Invalid credentials produce clear error messages
- [x] Unit tests pass for config and API packages

## Deliverables

- `go.mod` and `go.sum`
- `cmd/ajira/main.go`
- `internal/cli/root.go`
- `internal/cli/me.go`
- `internal/cli/project.go`
- `internal/config/config.go`
- `internal/api/client.go`
- Unit tests for config and API packages

## Technical Approach

### Project Structure

```
ajira/
├── cmd/ajira/main.go
├── internal/
│   ├── cli/
│   │   ├── root.go
│   │   ├── me.go
│   │   └── project.go
│   ├── api/
│   │   └── client.go
│   └── config/
│       └── config.go
├── go.mod
└── go.sum
```

### Configuration

Read from environment variables:

- `JIRA_BASE_URL` (required)
- `JIRA_EMAIL` (required)
- `JIRA_API_TOKEN` (required, fallback to `ATLASSIAN_API_TOKEN`)
- `JIRA_PROJECT` (optional default project)

### API Client

- HTTP client with configurable timeout
- Basic Auth header: base64(email:token)
- JSON request/response handling
- Error response parsing

### CLI Framework

- Cobra for command structure
- Persistent flags on root command
- Consistent output formatting (text vs JSON)

## Workflow

Follow the Feature Development Workflow (docs/workflow.md) for each command:

1. Identify the feature
2. Read required documentation
3. Discuss implementation options
4. Create Design Record if significant decisions arise
5. Implement
6. Test and fix
7. Update this project document

## Dependencies

- Cobra library for CLI framework
- Standard library for HTTP and JSON

## Questions and Uncertainties

- Should we use a custom HTTP client or standard library?
- How should API errors be structured for consistent handling?
- What timeout value is appropriate for API calls?
