# P-020: List Subcommands for Comments and Links

- Status: Complete
- Started: 2026-02-19
- Completed: 2026-02-19

## Overview

Add dedicated `list` subcommands to the `comment` and `link` command groups. Currently comments and links are only viewable through `ajira issue view`, which fetches the entire issue. Dedicated list commands provide direct access, better formatting, and consistent CLI patterns matching the existing `ajira issue attachment list` command.

## Goals

1. Implement `ajira issue comment list` with limit and JSON support
2. Implement `ajira issue link list` with JSON support
3. Follow established patterns from `ajira issue attachment list`
4. Reuse existing API calls and data structures
5. Maintain backward compatibility with `issue view` flags

## Scope

In Scope:

- `ajira issue comment list <issue-key>` command
- `ajira issue comment list <issue-key> -l <count>` for limiting results
- `ajira issue comment list <issue-key> --json` for JSON output
- `ajira issue link list <issue-key>` command
- `ajira issue link list <issue-key> --json` for JSON output
- Unit tests for both commands
- CLI help documentation for both commands

Out of Scope:

- Changes to `ajira issue view` behaviour
- Pagination beyond simple limit flag
- Filtering or search within comments/links
- New API client methods (reuse existing)

## Success Criteria

- [x] `ajira issue comment list PROJ-123` displays comments in readable format
- [x] `ajira issue comment list PROJ-123 -l 10` limits to 10 most recent comments
- [x] `ajira issue comment list PROJ-123 --json` outputs valid JSON
- [x] `ajira issue link list PROJ-123` displays links with direction (inward/outward)
- [x] `ajira issue link list PROJ-123 --json` outputs valid JSON
- [x] All existing tests pass
- [x] New unit tests cover both commands
- [x] Help text follows existing CLI help conventions

## Deliverables

Files to create:

- `internal/cli/issue_comment_list.go` - Comment list command
- `internal/cli/issue_link_list.go` - Link list command
- `internal/cli/issue_comment_list_test.go` - Tests for comment list

Files to modify:

- `internal/cli/issue_link_test.go` - Add tests for link list
- `internal/cli/help/agents.md` - Add comment list and link list entries

## Current State

Existing patterns and reusable functions:

- `issue_attachment_list.go` is the canonical pattern: Cobra command with `cobra.ExactArgs(1)`, `SilenceUsage: true`, `RunE` function that loads config, creates client, fetches data, outputs via tabwriter or `json.MarshalIndent`
- `getComments()` in `issue_view.go` (line 365) already fetches comments via `GET /issue/{key}/comment?maxResults=N&orderBy=-created`, converts ADF to Markdown, returns `([]CommentInfo, int, error)`
- `getIssueLinks()` in `issue_link_remove.go` already fetches links via `GET /issue/{key}?fields=issuelinks`, returns `([]issueLink, error)`
- `CommentInfo` struct defined in `issue_view.go`: ID, Author, Created, Body (Markdown)
- `LinkInfo` struct defined in `issue_view.go`: Direction, Key, Status, Summary (used by `issue view` display)
- `issueLink` struct defined in `issue_link_remove.go`: raw API type with InwardIssue/OutwardIssue
- Subcommand registration uses `init()` in the new file (e.g., `issueCommentCmd.AddCommand(issueCommentListCmd)`)

Test file status:

- `issue_comment_test.go` does not exist, needs to be created
- `issue_link_test.go` exists with tests for `getIssueLinks()`, `createIssueLink()`, etc.
- `issue_attachment_test.go` provides the list test pattern: `httptest.NewServer`, `testConfig()`, direct function calls

Flag conventions for list commands:

- `issue list`, `epic list`, `sprint list` all use `--limit` / `-l` (not `-n`)
- `issue view` uses `--comments` / `-c` for comment count

CLI help:

- Help lives in `internal/cli/help/agents.md` (not `docs/cli/`)
- `agents.md` needs new entries for `comment list` and `link list`

JSON output:

- List commands use `json.MarshalIndent` + `fmt.Println` directly (not `PrintSuccessJSON`)
- `PrintSuccessJSON` is for mutation commands and respects `--quiet`

## Technical Approach

Follow the patterns established by `ajira issue attachment list`:

- Reuse `getComments()` from `issue_view.go` for comment list
- Reuse `getIssueLinks()` from `issue_link_remove.go` for link list, convert to `LinkInfo` for display
- Use `tabwriter` for formatted table output (links)
- Use `json.MarshalIndent` + `fmt.Println` for JSON output (not `PrintSuccessJSON`)
- Register as subcommand under existing parent commands via `init()`

## Notes

- Created from `list.md` enhancement proposal (2026-01-23)
- Pattern: all `issue <resource>` commands should have a `list` subcommand
- Priority is low as current workarounds via `issue view` are functional
