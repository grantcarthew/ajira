# dr-015: Comment Edit Command

- Status: Proposed
- Date: 2026-01-15
- Project: p-016-cli-comment-edit

## Context

Users need the ability to edit existing comments on Jira issues. Currently ajira only supports adding comments via `issue comment add`. The Jira REST API v3 provides a PUT endpoint for updating comments.

Current state:

- `issue comment add` creates new comments using POST to `/issue/{key}/comment`
- `issue view -c N --json` returns comments with their IDs
- Text output does not display comment IDs
- Comment IDs are required to edit comments via API

## Decision

Implement `issue comment edit` command with the following design:

Command syntax:
```
ajira issue comment edit <issue-key> <comment-id> [text]
ajira issue comment edit <issue-key> <comment-id> --body "text"
ajira issue comment edit <issue-key> <comment-id> --file comment.md
```

API endpoint:
```
PUT /rest/api/3/issue/{issueIdOrKey}/comment/{id}
```

Request body (same as comment add):
```json
{
  "body": { /* ADF document */ }
}
```

Additionally, modify text output of `issue view -c N` to display comment IDs for easier workflow.

## Alternatives Considered

### Alternative 1: Edit by comment index (1, 2, 3...)

```
ajira issue comment edit PROJ-123 1 "new text"  # Edit first comment
```

Pros:
- Simpler for users (no need to look up IDs)
- Works with text output

Cons:
- Requires fetching all comments to map index to ID
- Index can change if comments are deleted
- Ambiguous ordering (oldest first? newest first?)
- Extra API call to resolve index

Rejected: Too fragile and requires extra API calls.

### Alternative 2: Interactive selection

```
ajira issue comment edit PROJ-123  # Lists comments, prompts for selection
```

Pros:
- User-friendly

Cons:
- Violates non-interactive design principle
- Not suitable for automation

Rejected: CLI is designed for non-interactive use.

### Alternative 3: Edit last comment

```
ajira issue comment edit PROJ-123 --last "new text"
```

Pros:
- Common use case (fixing typo in just-added comment)
- No ID lookup needed

Cons:
- Only works for last comment
- Still need ID-based edit for other comments

Considered: Could be added as convenience flag, but not essential for initial implementation.

## Consequences

Positive:
- Complete comment lifecycle (add, edit, delete would be next)
- Consistent with Jira web UI capabilities
- Follows existing patterns for text input (--body, --file)

Negative:
- Users must use --json to get comment IDs (mitigated by adding IDs to text output)
- Comment IDs are opaque numeric strings

Implementation notes:
- Reuse `commentAddRequest` struct (same body format)
- Add `editComment` function using `client.Put()`
- Add comment ID to text output format: `[date] [id] Author:`
- No batch mode (each comment has unique ID)

## Related

- p-016-cli-comment-edit (implementation project)
- dr-007-adf-markdown-conversion (comment body format)
