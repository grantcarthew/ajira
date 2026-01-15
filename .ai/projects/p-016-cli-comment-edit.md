# p-016: Comment Edit Command

- Status: Pending
- Started:
- Completed:
- Design: dr-015-cli-comment-edit.md

## Overview

Add the ability to edit existing comments on Jira issues. This completes the comment management functionality alongside the existing `comment add` command. Users will be able to correct typos, update information, or refine comments after creation.

## Goals

1. Implement `issue comment edit` command
2. Display comment IDs in text output for easier workflow
3. Support same input methods as `comment add` (inline, --body, --file)
4. Maintain consistency with existing CLI patterns

## Scope

In Scope:

- `issue comment edit <issue-key> <comment-id> [text]` command
- `--body` flag for inline text
- `--file` flag for file input (including stdin via `-`)
- `--dry-run` support
- Display comment IDs in `issue view -c N` text output
- Update help documentation

Out of Scope:

- Batch editing (each comment has unique ID)
- Comment deletion (future project)
- Edit by index (1st, 2nd comment) - too fragile
- Interactive comment selection - violates non-interactive design

## Success Criteria

- [ ] `issue comment edit PROJ-123 12345 "new text"` updates comment
- [ ] `issue comment edit PROJ-123 12345 -f comment.md` reads from file
- [ ] `issue comment edit PROJ-123 12345 -f -` reads from stdin
- [ ] `issue view PROJ-123 -c 5` shows comment IDs in text output
- [ ] `--dry-run` shows planned action without executing
- [ ] Error handling for invalid comment ID
- [ ] Help text and examples documented
- [ ] Unit tests for edit functionality

## Deliverables

- Updated `internal/cli/issue_comment.go` - Add edit command
- Updated `internal/cli/issue_view.go` - Show comment IDs in text output
- Updated `internal/cli/help/agents.md` - Add edit example
- Updated `docs/flags-and-arguments.md` - Document edit command
- dr-015-cli-comment-edit.md - Design record
- Unit tests for comment edit

## Technical Approach

API endpoint:

```
PUT /rest/api/3/issue/{issueIdOrKey}/comment/{id}
```

Request body (identical to comment add):

```json
{
  "body": { /* ADF document */ }
}
```

Implementation:

1. Add `issueCommentEditCmd` cobra command
2. Reuse `commentAddRequest` struct for request body
3. Create `editComment()` function using `client.Put()`
4. Add comment ID to text output format: `[date] [id] Author:`
5. Share flag handling with `comment add` where possible

Command structure:

```
ajira issue comment edit <issue-key> <comment-id> [text]
                         --body, -b    Comment text in Markdown
                         --file, -f    Read from file (- for stdin)
```

## Current State

Comment viewing (`issue view -c N`):

- Text output: `[2026-01-15 11:43] Grant Carthew: comment body`
- JSON output includes: `{"id": "2599838", "author": "...", ...}`
- Comment IDs only visible in JSON output

Comment adding (`issue comment add`):

- Flags: `--body`, `--file`, `--stdin` (for batch keys)
- Uses POST to `/issue/{key}/comment`
- Returns comment ID in response

## Dependencies

None - builds on existing comment infrastructure.
