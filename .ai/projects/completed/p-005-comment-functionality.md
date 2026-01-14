# p-005: Comment Functionality

- Status: Completed
- Started: 2026-01-05
- Completed: 2026-01-05

## Overview

Implement comment management for Jira issues. Comments are a critical part of issue collaboration, and this project enables adding comments with Markdown formatting and viewing existing comments.

## Goals

1. Add comments to issues with Markdown support
2. Display comments when viewing issues
3. Support multiple input methods (inline, file, stdin)

## Scope

In Scope:

- `ajira issue comment add` - add a comment to an issue
- Comment display in `ajira issue view` output
- Markdown input for comment body (via p-003 conversion)
- Plain text and JSON output

Out of Scope:

- Editing existing comments
- Deleting comments
- Comment visibility restrictions (internal vs external)
- Mentioning users in comments

## Success Criteria

- [x] `ajira issue comment add ISSUE-KEY "text"` adds a comment
- [x] `ajira issue comment add ISSUE-KEY -b "text"` adds via --body flag
- [x] `ajira issue comment add ISSUE-KEY -f file.md` adds from file
- [x] `echo "text" | ajira issue comment add ISSUE-KEY -f -` adds from stdin
- [x] `ajira issue view ISSUE-KEY` displays recent comments
- [x] `ajira issue view ISSUE-KEY --comments N` controls comment count
- [x] Comments display author, timestamp, and Markdown-formatted body
- [x] All commands support --json output
- [x] Unit tests for comment functionality

## Deliverables

- `internal/cli/issue_comment.go` - comment subcommand
- Updates to `internal/cli/issue_view.go` for comment display
- API client method for adding comments
- Unit tests

## Technical Approach

### API Endpoints

| Operation | Method | Endpoint |
|-----------|--------|----------|
| Add comment | POST | `/rest/api/3/issue/{key}/comment` |
| Get comments | GET | `/rest/api/3/issue/{key}/comment` |

### Comment Input Priority

When adding a comment, input is resolved in this order:

1. `--file` flag with path or `-` for stdin
2. `--body` flag with inline text
3. Positional argument after issue key

### Comment Display Format

Plain text format for `issue view`:

```
---
Comments (2):

[2024-01-16 14:22] Grant Carthew:
Started working on this today.

[2024-01-16 15:30] Jane Smith:
Looks good, please add tests.
```

## Dependencies

- p-003 (Markdown/ADF Conversion) for comment body handling
- p-004 (Issue Commands) for issue view integration

## Questions and Uncertainties

- Should we support listing comments separately from issue view?
- How many comments should be displayed by default?
- Should we support comment threading/replies?
