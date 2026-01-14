# p-004: Issue Commands

- Status: Completed
- Started: 2026-01-05
- Completed: 2026-01-05

## Overview

Implement the core issue management commands for ajira. These commands provide full CRUD operations for Jira issues plus workflow transitions and assignment, enabling AI agents and scripts to manage issues programmatically.

## Goals

1. List and search issues with JQL support
2. View issue details with Markdown-formatted descriptions
3. Create issues with Markdown descriptions
4. Edit existing issues
5. Delete issues
6. Assign issues to users
7. Transition issues between statuses

## Scope

In Scope:

- `ajira issue list` - search with JQL and convenience filters
- `ajira issue view` - display issue details
- `ajira issue create` - create new issues
- `ajira issue edit` - update issue fields
- `ajira issue delete` - delete issues
- `ajira issue assign` - assign/unassign issues
- `ajira issue move` - transition issue status
- Plain text and JSON output formats
- Markdown input for descriptions (via p-003 conversion)

Out of Scope:

- Comments (covered in p-005)
- Attachments and images
- Worklogs
- Issue linking
- Watchers
- Subtasks (may be added later)

## Success Criteria

- [x] `ajira issue list` returns issues with JQL filtering
- [x] `ajira issue list` supports convenience flags (--status, --type, --assignee)
- [x] `ajira issue view ISSUE-KEY` displays formatted issue details
- [x] `ajira issue create -s "Summary"` creates an issue and returns key
- [x] `ajira issue create` accepts Markdown body via --body, --file, or stdin
- [x] `ajira issue edit ISSUE-KEY` updates specified fields
- [x] `ajira issue delete ISSUE-KEY` deletes the issue
- [x] `ajira issue assign ISSUE-KEY USER` assigns the issue
- [x] `ajira issue assign ISSUE-KEY unassigned` removes assignee
- [x] `ajira issue move ISSUE-KEY STATUS` transitions the issue
- [x] All commands support --json output
- [x] Unit tests for command logic
- [x] Error messages are clear and actionable

## Deliverables

- `internal/cli/issue.go` - parent command
- `internal/cli/issue_list.go`
- `internal/cli/issue_view.go`
- `internal/cli/issue_create.go`
- `internal/cli/issue_edit.go`
- `internal/cli/issue_delete.go`
- `internal/cli/issue_assign.go`
- `internal/cli/issue_move.go`
- API client methods in `internal/api/client.go`
- Unit tests for each command

## Technical Approach

### API Endpoints

| Command | Method | Endpoint |
|---------|--------|----------|
| list | POST | `/rest/api/3/search` |
| view | GET | `/rest/api/3/issue/{key}` |
| create | POST | `/rest/api/3/issue` |
| edit | PUT | `/rest/api/3/issue/{key}` |
| delete | DELETE | `/rest/api/3/issue/{key}` |
| assign | PUT | `/rest/api/3/issue/{key}/assignee` |
| move | GET | `/rest/api/3/issue/{key}/transitions` |
| move | POST | `/rest/api/3/issue/{key}/transitions` |

### JQL Building

The `issue list` command builds JQL from convenience flags:

- `--project` → `project = X`
- `--status` → `status = "X"`
- `--type` → `issuetype = X`
- `--assignee` → `assignee = X` or `assignee IS EMPTY`
- `--query` → raw JQL (overrides other filters)

Flags are combined with AND.

### User Resolution

For assignee, accept:

- Email address → resolve to accountId via user search
- Account ID → use directly
- "unassigned" → set to null

## Dependencies

- p-003 (Markdown/ADF Conversion) for description handling

## Questions and Uncertainties

- How should we handle pagination for large result sets in issue list?
- Should we support --columns flag for customising list output?
- How do we handle custom fields?
