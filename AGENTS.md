# ajira

Atlassian Jira CLI designed for AI agents and automation. Non-interactive, environment-configured, with Markdown input/output and JSON support.

## Setup

```bash
# Clone and build
git clone https://github.com/grantcarthew/ajira.git
cd ajira
go build -o ajira ./cmd/ajira

# Configure environment
export JIRA_BASE_URL="https://example.atlassian.net"
export JIRA_EMAIL="user@example.com"
export JIRA_API_TOKEN="your-token"
export JIRA_PROJECT="PROJ"

# Verify
./ajira me
```

## Build and Test

```bash
# Build
go build -o ajira ./cmd/ajira

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/cli/...
go test ./internal/jira/...

# Quick validation
./ajira me
./ajira project list
./ajira issue list -l 3
```

## Code Style

- Go standard formatting (`gofmt`)
- No cgo - pure Go only
- Error handling: return errors, don't panic
- Cobra for CLI commands
- Environment variables for configuration (no config files)
- Non-interactive: all input via flags, arguments, or stdin

## Project Structure

```
cmd/ajira/         Main entry point
internal/cli/      Command implementations
internal/jira/     Jira API client
internal/markdown/ Markdown to ADF conversion
.ai/               AI agent working files (projects, design records, tasks)
docs/              Human-facing documentation
```

## Active Project

Projects are stored in `.ai/projects/`. Update this section when starting a new project.

Active Project: None

Completed: p-014 Auxiliary Commands (2026-01-15), p-013 Automation Support (2026-01-15), p-016 Comment Edit (2026-01-15), p-011 Issue Command Enhancements (2026-01-13), p-010 Agile Features (2026-01-13), p-009 Issue Clone (2026-01-12), p-008 Issue List Enhancements (2026-01-09), p-007 Issue Linking (2026-01-08), p-015 CLI Help System (2026-01-07), p-006 Integration Testing (2026-01-07), p-005 Comment Functionality (2026-01-05)

## Development Guidelines

- Read `.ai/workflow.md` for feature development process
- Read `.ai/projects/p-writing-guide.md` for project documentation
- Read `.ai/design/dr-writing-guide.md` for design record format
- Design records are in `.ai/design/design-records/`
