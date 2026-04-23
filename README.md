# ajira

[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](https://www.mozilla.org/en-US/MPL/2.0/)
[![Go Report Card](https://goreportcard.com/badge/github.com/grantcarthew/ajira)](https://goreportcard.com/report/github.com/grantcarthew/ajira)
[![Go Reference](https://pkg.go.dev/badge/github.com/grantcarthew/ajira.svg)](https://pkg.go.dev/github.com/grantcarthew/ajira)
[![GitHub Release](https://img.shields.io/github/v/release/grantcarthew/ajira)](https://github.com/grantcarthew/ajira/releases)

Atlassian Jira Cloud CLI designed for AI agents and automation. Jira Server and Data Center are not supported.

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
export JIRA_PROJECT="PROJ"    # Optional default project
export JIRA_BOARD="42"        # Optional default board for agile commands

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

Homebrew (Linux/macOS):

```bash
brew tap grantcarthew/tap
brew install grantcarthew/tap/ajira
```

Go Install:

```bash
go install github.com/grantcarthew/ajira/cmd/ajira@latest
```

Build from Source:

```bash
git clone https://github.com/grantcarthew/ajira.git
cd ajira
go build -o ajira ./cmd/ajira
./ajira --version
```

## Configuration

Environment variables only. No config files, no interactive setup. `ATLASSIAN_*` variables are shared with the sibling `acon` (Confluence) tool; `JIRA_*` variables override them when set.

| Variable | Required | Description |
|----------|----------|-------------|
| `ATLASSIAN_BASE_URL` | Yes | Atlassian instance URL (e.g., `https://example.atlassian.net`) |
| `ATLASSIAN_EMAIL` | Yes | Atlassian account email |
| `ATLASSIAN_API_TOKEN` | Yes | API token |
| `JIRA_BASE_URL` | No | Overrides `ATLASSIAN_BASE_URL` |
| `JIRA_EMAIL` | No | Overrides `ATLASSIAN_EMAIL` |
| `JIRA_API_TOKEN` | No | Overrides `ATLASSIAN_API_TOKEN` |
| `JIRA_PROJECT` | No | Default project key (e.g., `PROJ`) |
| `JIRA_BOARD` | No | Default board ID for agile commands |

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
ajira issue list --status "In Progress"

# Filter by assignee and type
ajira issue list -a me -t Bug

# Custom JQL query
ajira issue list -q "status = Done AND updated >= -7d"

# Limit results
ajira issue list -l 10
```

### View Issue Details

```bash
# View issue (shows 5 most recent comments by default)
ajira issue view PROJ-123

# Show more comments
ajira issue view PROJ-123 -c 10

# Hide comments
ajira issue view PROJ-123 -c 0

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

### Clone Issues

```bash
# Clone with same fields
ajira issue clone PROJ-123

# Override summary and link to original
ajira issue clone PROJ-123 -s "New summary" --link

# Clone to a different project
ajira issue clone PROJ-123 -p OTHER
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

# List comments
ajira issue comment list PROJ-123

# Edit existing comment (use issue view -c N to find comment IDs)
ajira issue comment edit PROJ-123 12345 "Updated text"
```

### Attachments

```bash
# Upload one or more files
ajira issue attachment add PROJ-123 screenshot.png
ajira issue attachment add PROJ-123 *.log

# List attachments
ajira issue attachment list PROJ-123

# Download an attachment
ajira issue attachment download PROJ-123 <attachment-id>

# Remove an attachment
ajira issue attachment remove PROJ-123 <attachment-id>
```

### Links

```bash
# Link two issues
ajira issue link add PROJ-123 PROJ-456 "Blocks"

# List links on an issue
ajira issue link list PROJ-123

# List available link types
ajira issue link types

# Add a web URL as a remote link
ajira issue link url PROJ-123 https://example.com "Design doc"

# Remove a link
ajira issue link remove PROJ-123 PROJ-456
```

### Watch Issues

```bash
# Watch an issue
ajira issue watch PROJ-123

# Unwatch
ajira issue unwatch PROJ-123

# Batch watch from stdin
echo -e "PROJ-1\nPROJ-2" | ajira issue watch --stdin
```

### Delete Issues

```bash
# Delete issue (permanent)
ajira issue delete PROJ-123
```

### Open in Browser

```bash
# Open project in browser
ajira open

# Open a specific issue
ajira open PROJ-123
```

### Discovery Commands

```bash
# List available issue types, statuses, priorities
ajira issue type
ajira issue status
ajira issue priority

# List accessible projects
ajira project list

# Search users (returns account IDs for assign)
ajira user search john
ajira user search john@example.com -l 20

# List Jira fields
ajira field list
```

## Agile Commands

Epic, sprint, board, and release commands. Sprint operations require `JIRA_BOARD` or `--board`.

```bash
# Boards
ajira board list

# Sprints (need a board id)
ajira sprint list
ajira sprint list --current
ajira sprint list --state closed -l 5
ajira sprint add 42 PROJ-123 PROJ-124
echo -e "PROJ-1\nPROJ-2" | ajira sprint add 42 --stdin

# Epics
ajira epic list
ajira epic list --status "In Progress"
ajira epic create -s "Auth Epic"
ajira epic create -s "API" -d "Description" -P Major -a me
ajira epic add EPIC-1 PROJ-123 PROJ-124
ajira epic remove PROJ-123 PROJ-124

# Releases
ajira release list
```

See `ajira help agile` for the full reference.

## Automation Examples

### Preview Before Acting

Most mutating commands accept `--dry-run` to show what would happen without calling the API.

```bash
ajira issue create -s "Test" --dry-run
ajira issue move PROJ-123 Done --dry-run
```

### Create and Assign in One Pipeline

```bash
KEY=$(ajira issue create -s "New task" --json | jq -r .key)
ajira issue assign "$KEY" me
ajira issue move "$KEY" "In Progress"
```

### Bulk Transition Issues

```bash
ajira issue list --status "To Do" --json | jq -r '.[].key' | while read key; do
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

For AI agents and LLMs, ajira includes token-efficient help topics:

```bash
ajira help agents      # Compact command reference for AI context windows
ajira help agile       # Epic, sprint, board reference
ajira help markdown    # Markdown-to-ADF formatting rules
ajira help schemas     # JSON output schemas
```

Exit codes are stable and documented — see `internal/cli/exitcodes.go`.

## CLI Reference

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--project` | `-p` | Override default project key |
| `--board` |  | Override default board ID |
| `--dry-run` |  | Preview actions without executing |
| `--quiet` |  | Suppress non-essential output |
| `--no-color` |  | Disable coloured output |
| `--verbose` |  | Show HTTP request/response details |
| `--version` | `-v` | Print version |
| `--help` | `-h` | Print help |

### Commands

| Command | Description |
|---------|-------------|
| `me` | Display current user information |
| `open [issue]` | Open project or issue in browser |
| `project list` | List accessible projects |
| `board list` | List boards |
| `sprint list` | List sprints |
| `sprint add` | Add issues to a sprint |
| `epic list` | List epics |
| `epic create` | Create a new epic |
| `epic add` | Add issues to an epic |
| `epic remove` | Remove issues from their epic |
| `release list` | List project releases / versions |
| `issue list` | List and search issues |
| `issue view` | View issue details |
| `issue create` | Create a new issue |
| `issue edit` | Edit an existing issue |
| `issue clone` | Clone an issue |
| `issue delete` | Delete an issue |
| `issue assign` | Assign an issue to a user |
| `issue move` | Transition an issue to a new status |
| `issue watch` / `unwatch` | Add or remove yourself as a watcher |
| `issue comment add` / `edit` / `list` | Manage comments |
| `issue attachment add` / `list` / `download` / `remove` | Manage attachments |
| `issue link add` / `remove` / `list` / `types` / `url` | Manage issue links and remote URLs |
| `issue type` / `status` / `priority` | List metadata options |
| `user search` | Search users by name or email |
| `field list` | List Jira fields |
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

`ajira` is licensed under the [Mozilla Public License 2.0](LICENSE).

## Author

Grant Carthew
