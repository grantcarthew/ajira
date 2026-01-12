# ajira Initial Design

Atlassian Jira CLI designed for AI agents and automation.

## Overview

`ajira` is a non-interactive command-line tool for Atlassian Jira, modelled after the `acon` Confluence CLI. It prioritises:

- **AI agent compatibility**: No interactive prompts, TUI views, or keyboard navigation
- **Simplicity**: Environment-based configuration, no init wizards or config files
- **Scriptability**: Plain text and JSON output formats
- **Markdown**: Human-friendly input/output with automatic ADF conversion

## Project Structure

```
ajira/
├── cmd/
│   └── ajira/
│       └── main.go                 # Entry point: cli.Execute()
├── internal/
│   ├── cli/
│   │   ├── root.go                 # NewRootCmd(), Execute(), persistent flags
│   │   ├── issue.go                # addIssueCommand() - parent orchestrator
│   │   ├── issue_list.go           # ajira issue list
│   │   ├── issue_view.go           # ajira issue view
│   │   ├── issue_create.go         # ajira issue create
│   │   ├── issue_edit.go           # ajira issue edit
│   │   ├── issue_delete.go         # ajira issue delete
│   │   ├── issue_assign.go         # ajira issue assign
│   │   ├── issue_move.go           # ajira issue move (transitions)
│   │   ├── issue_comment.go        # ajira issue comment add
│   │   ├── me.go                   # ajira me
│   │   └── project.go              # ajira project list
│   ├── api/
│   │   └── client.go               # Jira REST API client (v3)
│   ├── config/
│   │   └── config.go               # Environment-based configuration
│   └── converter/
│       ├── adf.go                  # Markdown to ADF conversion
│       └── markdown.go             # ADF to Markdown conversion
├── go.mod
└── go.sum
```

## Configuration

Environment variables only. No config files, no interactive setup.

| Variable | Required | Description |
|----------|----------|-------------|
| `JIRA_BASE_URL` | Yes | Atlassian instance URL (e.g., `https://autogeneral-au.atlassian.net`) |
| `JIRA_EMAIL` | Yes | Your Atlassian account email |
| `JIRA_API_TOKEN` | Yes | API token (fallback: `ATLASSIAN_API_TOKEN`) |
| `JIRA_PROJECT` | No | Default project key (e.g., `GCP`) |

## Command Reference

### Root Command

```
ajira [command] [flags]
```

**Global Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |
| `--project` | `-p` | Override default project key |
| `--version` | `-v` | Print version |
| `--help` | `-h` | Print help |

---

### ajira me

Display current user information.

```
ajira me [-j]
```

**Description:**
Returns the authenticated user's account ID and email address. Useful for scripting assignee values.

**Output (plain):**

```
Account ID: 712020:f0aa8349-1860-4d9d-bf31-d12e26b85d84
Email: grant.carthew@autogeneral.com.au
Display Name: Grant Carthew
```

**Output (JSON):**

```json
{
  "accountId": "712020:f0aa8349-1860-4d9d-bf31-d12e26b85d84",
  "emailAddress": "grant.carthew@autogeneral.com.au",
  "displayName": "Grant Carthew"
}
```

---

### ajira project list

List accessible projects.

```
ajira project list [-l LIMIT] [-j]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--limit` | `-l` | 25 | Maximum projects to return |

**Description:**
Lists Jira projects the authenticated user has access to.

**Output (plain):**

```
KEY     NAME
GCP     GCP Cloud Platform
PROJ    Another Project
```

---

### ajira issue list

List and search issues.

```
ajira issue list [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--project` | `-p` | env | Project key |
| `--query` | `-q` | | Raw JQL query |
| `--status` | | | Filter by status |
| `--type` | `-t` | | Filter by issue type |
| `--assignee` | `-a` | | Filter by assignee (email, account ID, or "unassigned") |
| `--reporter` | `-r` | | Filter by reporter |
| `--priority` | `-y` | | Filter by priority |
| `--label` | `-L` | | Filter by label (repeatable) |
| `--created` | | | Filter by created date (e.g., "-7d", "2024-01-01") |
| `--updated` | | | Filter by updated date |
| `--limit` | `-l` | 25 | Maximum issues to return |
| `--offset` | | 0 | Starting offset for pagination |
| `--json` | `-j` | | Output as JSON |

**Description:**
Searches for issues using JQL. Flags are combined into a JQL query. Use `--query` for complex queries.

**Examples:**

```bash
# List issues in default project
ajira issue list

# List issues assigned to me
ajira issue list -a $(ajira me | grep "Account ID" | cut -d' ' -f3)

# List high priority bugs in progress
ajira issue list -t Bug -y High --status "In Progress"

# Raw JQL query
ajira issue list -q "assignee = currentUser() AND status != Done"

# JSON output for scripting
ajira issue list -l 10 -j
```

**Output (plain):**

```
KEY        STATUS        ASSIGNEE              SUMMARY
GCP-123    In Progress   Grant Carthew         Implement feature X
GCP-122    To Do         Unassigned            Fix bug in login
```

---

### ajira issue view

View issue details.

```
ajira issue view ISSUE-KEY [flags]
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `ISSUE-KEY` | Yes | Issue key (e.g., GCP-123) |

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--comments` | | 1 | Number of comments to display |
| `--json` | `-j` | | Output as JSON |

**Description:**
Displays issue details including summary, description (converted to Markdown), status, assignee, and recent comments.

**Examples:**

```bash
# View issue
ajira issue view GCP-123

# View with 5 recent comments
ajira issue view GCP-123 --comments 5

# JSON output
ajira issue view GCP-123 -j
```

**Output (plain):**

```
Key: GCP-123
Summary: Implement feature X
Status: In Progress
Type: Task
Priority: High
Assignee: Grant Carthew
Reporter: Jane Smith
Created: 2024-01-15T10:30:00+10:00
Updated: 2024-01-16T14:22:00+10:00

Description:
This task involves implementing feature X as per the requirements.

## Acceptance Criteria
- Criterion 1
- Criterion 2

---
Comments (1):

[2024-01-16 14:22] Grant Carthew:
Started working on this today.
```

---

### ajira issue create

Create a new issue.

```
ajira issue create -s SUMMARY [flags]
```

**Flags:**

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--summary` | `-s` | Yes | | Issue summary |
| `--type` | `-t` | No | Task | Issue type (Task, Bug, Story, Epic) |
| `--body` | `-b` | No | | Description (Markdown, inline) |
| `--file` | `-f` | No | | Read description from file (use `-` for stdin) |
| `--priority` | `-y` | No | | Priority (Highest, High, Medium, Low, Lowest) |
| `--assignee` | `-a` | No | | Assignee (email or account ID) |
| `--reporter` | `-r` | No | | Reporter (email or account ID) |
| `--label` | `-L` | No | | Labels (repeatable) |
| `--component` | `-C` | No | | Components (repeatable) |
| `--parent` | `-P` | No | | Parent issue key (for subtasks or epic link) |
| `--project` | `-p` | No | env | Project key |
| `--json` | `-j` | No | | Output as JSON |

**Description:**
Creates a new Jira issue. Description can be provided inline with `--body`, from a file with `--file`, or piped via stdin.

**Examples:**

```bash
# Simple task
ajira issue create -s "Implement login feature"

# Bug with description
ajira issue create -t Bug -s "Login fails on Safari" -b "Users report 500 error when logging in via Safari browser."

# From file
ajira issue create -s "New feature" -f description.md

# From stdin
echo "Description here" | ajira issue create -s "New task" -f -

# Full example
ajira issue create \
  -t Story \
  -s "User authentication" \
  -b "Implement OAuth2 authentication flow" \
  -y High \
  -a "grant.carthew@autogeneral.com.au" \
  -L backend \
  -L security
```

**Output (plain):**

```
Issue created successfully
Key: GCP-456
URL: https://autogeneral-au.atlassian.net/browse/GCP-456
```

---

### ajira issue edit

Edit an existing issue.

```
ajira issue edit ISSUE-KEY [flags]
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `ISSUE-KEY` | Yes | Issue key (e.g., GCP-123) |

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--summary` | `-s` | New summary |
| `--body` | `-b` | New description (Markdown, inline) |
| `--file` | `-f` | Read description from file |
| `--priority` | `-y` | New priority |
| `--json` | `-j` | Output as JSON |

**Description:**
Updates issue fields. Only specified fields are updated.

**Examples:**

```bash
# Update summary
ajira issue edit GCP-123 -s "Updated summary"

# Update description from file
ajira issue edit GCP-123 -f new-description.md

# Update priority
ajira issue edit GCP-123 -y High
```

**Output (plain):**

```
Issue updated successfully
Key: GCP-123
URL: https://autogeneral-au.atlassian.net/browse/GCP-123
```

---

### ajira issue delete

Delete an issue.

```
ajira issue delete ISSUE-KEY
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `ISSUE-KEY` | Yes | Issue key (e.g., GCP-123) |

**Description:**
Permanently deletes an issue. This action cannot be undone.

**Examples:**

```bash
ajira issue delete GCP-123
```

**Output (plain):**

```
Issue GCP-123 deleted successfully
```

---

### ajira issue assign

Assign an issue to a user.

```
ajira issue assign ISSUE-KEY ASSIGNEE
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `ISSUE-KEY` | Yes | Issue key (e.g., GCP-123) |
| `ASSIGNEE` | Yes | Email, account ID, or "unassigned" |

**Description:**
Assigns an issue to a user. Use "unassigned" to remove the assignee.

**Examples:**

```bash
# Assign to user by email
ajira issue assign GCP-123 grant.carthew@autogeneral.com.au

# Assign to self
ajira issue assign GCP-123 $(ajira me | grep "Account ID" | cut -d' ' -f3)

# Unassign
ajira issue assign GCP-123 unassigned
```

**Output (plain):**

```
Issue GCP-123 assigned to Grant Carthew
```

---

### ajira issue move

Transition an issue to a new status.

```
ajira issue move ISSUE-KEY STATUS
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `ISSUE-KEY` | Yes | Issue key (e.g., GCP-123) |
| `STATUS` | Yes | Target status name (e.g., "In Progress", "Done") |

**Description:**
Transitions an issue to a new status. The status must be a valid transition from the current state.

**Examples:**

```bash
# Move to In Progress
ajira issue move GCP-123 "In Progress"

# Move to Done
ajira issue move GCP-123 Done
```

**Output (plain):**

```
Issue GCP-123 moved to "In Progress"
```

---

### ajira issue comment add

Add a comment to an issue.

```
ajira issue comment add ISSUE-KEY [BODY] [flags]
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `ISSUE-KEY` | Yes | Issue key (e.g., GCP-123) |
| `BODY` | No | Comment text (if not using --file) |

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--body` | `-b` | Comment text (alternative to positional arg) |
| `--file` | `-f` | Read comment from file (use `-` for stdin) |

**Description:**
Adds a comment to an issue. Comment text can be provided as a positional argument, via `--body`, from a file, or piped via stdin.

**Examples:**

```bash
# Inline comment
ajira issue comment add GCP-123 "This is my comment"

# Using --body flag
ajira issue comment add GCP-123 -b "This is my comment"

# From file
ajira issue comment add GCP-123 -f comment.md

# From stdin
echo "Comment from stdin" | ajira issue comment add GCP-123 -f -
```

**Output (plain):**

```
Comment added to GCP-123
```

---

## API Client Methods

The `internal/api/client.go` will implement these methods:

| Method | HTTP | Endpoint | Description |
|--------|------|----------|-------------|
| `GetMyself` | GET | `/rest/api/3/myself` | Current user info |
| `ListProjects` | GET | `/rest/api/3/project/search` | List projects |
| `Search` | POST | `/rest/api/3/search` | Search issues with JQL |
| `GetIssue` | GET | `/rest/api/3/issue/{key}` | Get issue details |
| `CreateIssue` | POST | `/rest/api/3/issue` | Create issue |
| `UpdateIssue` | PUT | `/rest/api/3/issue/{key}` | Update issue fields |
| `DeleteIssue` | DELETE | `/rest/api/3/issue/{key}` | Delete issue |
| `AssignIssue` | PUT | `/rest/api/3/issue/{key}/assignee` | Assign issue |
| `GetTransitions` | GET | `/rest/api/3/issue/{key}/transitions` | Get valid transitions |
| `DoTransition` | POST | `/rest/api/3/issue/{key}/transitions` | Perform transition |
| `AddComment` | POST | `/rest/api/3/issue/{key}/comment` | Add comment |

## Content Conversion

### Markdown to ADF (input)

When creating or editing issues/comments, Markdown input is converted to Atlassian Document Format (ADF) for the API.

**Supported Markdown:**

- Headings (# to ######)
- Bold, italic, strikethrough
- Ordered and unordered lists
- Code blocks with language hints
- Inline code
- Links
- Tables

### ADF to Markdown (output)

When viewing issues, ADF content from the API is converted to Markdown for terminal display.

## Design Principles

1. **No interactivity**: All input via flags, arguments, files, or stdin
2. **Predictable output**: Consistent plain text format, machine-parseable JSON option
3. **Environment config**: No config files, no init wizards
4. **Fail fast**: Clear error messages, non-zero exit codes on failure
5. **Unix philosophy**: Composable with pipes and scripts
6. **Cloud-first**: Targets Atlassian Cloud API v3

## Comparison with jira-cli

| Feature | jira-cli | ajira |
|---------|----------|-------|
| Interactive TUI | Yes | No |
| Config wizard | Yes | No |
| Config files | Yes (.jira.yml) | No (env vars only) |
| Interactive prompts | Yes | No |
| API support | v2 + v3 | v3 only |
| Installation types | Cloud + Local | Cloud only |
| AI agent friendly | No | Yes |

## References

- [acon](https://github.com/grantcarthew/acon) - Confluence CLI (design reference)
- [jira-cli](https://github.com/ankitpokhrel/jira-cli) - Existing Jira CLI (feature reference)
- [Jira REST API v3](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/)
- [Atlassian Document Format](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
