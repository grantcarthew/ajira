# ajira

[![License: MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/grantcarthew/ajira)](https://goreportcard.com/report/github.com/grantcarthew/ajira)
[![Go Reference](https://pkg.go.dev/badge/github.com/grantcarthew/ajira.svg)](https://pkg.go.dev/github.com/grantcarthew/ajira)
[![GitHub Release](https://img.shields.io/github/v/release/grantcarthew/ajira)](https://github.com/grantcarthew/ajira/releases)

Atlassian Jira CLI designed for AI agents and automation.

## Why ajira?

**Built for scripts, pipelines, and AI agents.**

Most Jira CLIs are designed for humans with interactive prompts, TUI views, and keyboard navigation. ajira takes a different approach:

- **Non-interactive** - All input via flags, arguments, or stdin. No prompts, no wizards
- **Environment-configured** - No config files. Set three environment variables and go
- **Markdown input/output** - Write descriptions in Markdown, automatically converted to Jira's ADF format
- **JSON support** - Machine-parseable output for scripting and automation
- **AI agent friendly** - Token-efficient text output, built-in agent reference (`ajira help agents`)

**Perfect for:**

- CI/CD pipelines creating and updating issues
- AI coding assistants managing Jira tickets
- Shell scripts automating issue workflows
- Automation tools that need predictable, parseable output

## Quick Start

```bash
# Install via Homebrew
brew tap grantcarthew/tap
brew install grantcarthew/tap/ajira

# Configure (add to your shell profile)
export JIRA_BASE_URL="https://your-instance.atlassian.net"
export JIRA_EMAIL="your-email@example.com"
export JIRA_API_TOKEN="your-api-token"
export JIRA_PROJECT="PROJ"  # Optional default project

# Verify authentication
ajira me

# List your issues
ajira issue list -a me
```

## Installation

### Prerequisites

ajira requires a Jira Cloud instance and an API token. Generate a token at:
<https://id.atlassian.com/manage-profile/security/api-tokens>

### Install ajira

**Homebrew (Linux/macOS):**

```bash
brew tap grantcarthew/tap
brew install grantcarthew/tap/ajira
```

**Go Install:**

```bash
go install github.com/grantcarthew/ajira/cmd/ajira@latest
```

**Build from Source:**

```bash
git clone https://github.com/grantcarthew/ajira.git
cd ajira
go build -o ajira ./cmd/ajira
./ajira --version
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

### Verify Authentication

```bash
# Check current user
ajira me

# JSON output for automation
ajira me --json
```

### List Issues

```bash
# List issues in default project
ajira issue list

# Filter by status
ajira issue list -s "In Progress"

# Filter by assignee and type
ajira issue list -a me -t Bug

# Custom JQL query
ajira issue list -q "status = Done AND updated >= -7d"

# Limit results
ajira issue list -l 10
```

### View Issue Details

```bash
# View issue
ajira issue view PROJ-123

# Include recent comments
ajira issue view PROJ-123 -c 5

# JSON output
ajira issue view PROJ-123 --json
```

### Create Issues

```bash
# Create a task
ajira issue create -s "Fix login bug"

# Create a story with description
ajira issue create -s "Add user dashboard" -t Story -d "## Requirements

- Display user stats
- Show recent activity"

# Read description from file
ajira issue create -s "Implement feature" -f description.md

# Read description from stdin
echo "Description here" | ajira issue create -s "From stdin" -f -

# With labels and priority
ajira issue create -s "Critical bug" -t Bug --labels urgent,security --priority High
```

### Edit Issues

```bash
# Update summary
ajira issue edit PROJ-123 -s "Updated summary"

# Update description
ajira issue edit PROJ-123 -d "New description in Markdown"

# Change type and priority
ajira issue edit PROJ-123 -t Bug --priority High
```

### Assign Issues

```bash
# Assign to yourself
ajira issue assign PROJ-123 me

# Assign by email
ajira issue assign PROJ-123 user@example.com

# Unassign
ajira issue assign PROJ-123 unassigned
```

### Transition Issues

```bash
# List available transitions
ajira issue move PROJ-123

# Move to a status
ajira issue move PROJ-123 "In Progress"

# Move to Done
ajira issue move PROJ-123 Done
```

### Comments

```bash
# Add inline comment
ajira issue comment add PROJ-123 "Comment text in Markdown"

# Comment from file
ajira issue comment add PROJ-123 -f comment.md

# Comment from stdin
echo "Automated comment" | ajira issue comment add PROJ-123 -f -
```

### Delete Issues

```bash
# Delete issue (permanent)
ajira issue delete PROJ-123
```

### Discovery Commands

```bash
# List available issue types
ajira issue type

# List available statuses
ajira issue status

# List available priorities
ajira issue priority

# List accessible projects
ajira project list
```

## Automation Examples

### Create and Assign in One Pipeline

```bash
KEY=$(ajira issue create -s "New task" --json | jq -r .key)
ajira issue assign "$KEY" me
ajira issue move "$KEY" "In Progress"
```

### Bulk Transition Issues

```bash
ajira issue list -s "To Do" --json | jq -r '.[].key' | while read key; do
  ajira issue move "$key" "In Progress"
done
```

### CI/CD Integration

```bash
# Create issue on build failure
ajira issue create \
  -s "Build failed: $CI_PIPELINE_ID" \
  -t Bug \
  -d "Pipeline: $CI_PIPELINE_URL" \
  --labels ci-failure
```

## AI Agent Reference

For AI agents and LLMs, ajira includes a token-efficient reference:

```bash
ajira help agents
```

This outputs a compact command reference designed for AI context windows. For JSON schema documentation:

```bash
ajira help schemas
```

## CLI Reference

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--project` | `-p` | Override default project key |
| `--version` | `-v` | Print version |
| `--help` | `-h` | Print help |

### Commands

| Command | Description |
|---------|-------------|
| `me` | Display current user information |
| `project list` | List accessible projects |
| `issue list` | List and search issues |
| `issue view` | View issue details |
| `issue create` | Create a new issue |
| `issue edit` | Edit an existing issue |
| `issue delete` | Delete an issue |
| `issue assign` | Assign an issue to a user |
| `issue move` | Transition an issue to a new status |
| `issue comment add` | Add a comment to an issue |
| `issue link add` | Create a link between two issues |
| `issue link remove` | Remove links between two issues |
| `issue link types` | List available link types |
| `issue link url` | Add a web URL to an issue |
| `issue type` | List available issue types |
| `issue status` | List available statuses |
| `issue priority` | List available priorities |
| `completion` | Generate shell completion scripts |
| `help` | Help for commands and topics |

### Shell Completion

```bash
# Bash (Linux)
ajira completion bash > /etc/bash_completion.d/ajira

# Bash (macOS)
ajira completion bash > $(brew --prefix)/etc/bash_completion.d/ajira

# Zsh (macOS)
ajira completion zsh > $(brew --prefix)/share/zsh/site-functions/_ajira

# Fish
ajira completion fish > ~/.config/fish/completions/ajira.fish
```

## Roadmap

Planned features for future releases:

- **Issue Clone** - Duplicate issues with field modifications
- **Agile Features** - Epic and sprint management
- **Time Tracking** - Worklog support
- **Enhanced Filters** - More filter options, CSV output
- **Automation Support** - Dry-run mode, batch operations, exit codes

## Contributing

Contributions welcome! Please:

1. Check existing issues: <https://github.com/grantcarthew/ajira/issues>
2. Create issue for bugs or feature requests
3. Submit pull requests against `main` branch

### Reporting Issues

Include:

- ajira version: `ajira --version`
- Operating system and version
- Full command and error message

## License

`ajira` is licensed under the [MIT License](LICENSE).

## Author

Grant Carthew
