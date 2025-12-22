# ajira

Atlassian Jira CLI designed for AI agents and automation.

## Overview

`ajira` is a non-interactive command-line tool for Atlassian Jira Cloud, prioritising:

- AI agent compatibility: No interactive prompts, TUI views, or keyboard navigation
- Simplicity: Environment-based configuration, no init wizards or config files
- Scriptability: Plain text and JSON output formats
- Markdown: Human-friendly input/output with automatic ADF conversion

## Installation

```bash
go install github.com/grantcarthew/ajira/cmd/ajira@latest
```

## Configuration

Environment variables only. No config files, no interactive setup.

| Variable | Required | Description |
|----------|----------|-------------|
| `JIRA_BASE_URL` | Yes | Atlassian instance URL (e.g., `https://example.atlassian.net`) |
| `JIRA_EMAIL` | Yes | Your Atlassian account email |
| `JIRA_API_TOKEN` | Yes | API token (fallback: `ATLASSIAN_API_TOKEN`) |
| `JIRA_PROJECT` | No | Default project key (e.g., `PROJ`) |

## Usage

```bash
# Current user info
ajira me

# List projects
ajira project list

# List issues
ajira issue list
ajira issue list -q "assignee = currentUser() AND status != Done"

# View issue
ajira issue view PROJ-123

# Create issue
ajira issue create -s "Implement feature X" -t Task -b "Description here"

# Edit issue
ajira issue edit PROJ-123 -s "Updated summary"

# Transition issue
ajira issue move PROJ-123 "In Progress"

# Assign issue
ajira issue assign PROJ-123 user@example.com

# Add comment
ajira issue comment add PROJ-123 "Comment text"

# Delete issue
ajira issue delete PROJ-123
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |
| `--project` | `-p` | Override default project key |
| `--version` | `-v` | Print version |
| `--help` | `-h` | Print help |

## Design Principles

1. No interactivity - All input via flags, arguments, files, or stdin
2. Predictable output - Consistent plain text format, machine-parseable JSON option
3. Environment config - No config files, no init wizards
4. Fail fast - Clear error messages, non-zero exit codes on failure
5. Unix philosophy - Composable with pipes and scripts
6. Cloud-first - Targets Atlassian Cloud API v3

## License

MIT

---

## Documentation Driven Development (DDD)

This project uses Documentation Driven Development. Design decisions are documented in Design Records (DRs) before or during implementation.

For complete DR writing guidelines: See [docs/design/dr-writing-guide.md](docs/design/dr-writing-guide.md)

For project writing guidelines: See [docs/projects/p-writing-guide.md](docs/projects/p-writing-guide.md)

For feature development workflow: See [docs/workflow.md](docs/workflow.md)

Location: `docs/design/design-records/`
