# P-006: Integration Testing

- Status: In Progress
- Started: 2026-01-05
- Completed:

## Overview

Perform end-to-end integration testing of ajira against a real Jira instance. This project validates that all CLI commands work correctly with the Jira API and that Markdown/ADF conversion produces valid results accepted by Jira.

This is an interactive testing project where each feature is tested manually and results are confirmed before proceeding.

## Goals

1. Verify authentication and configuration work correctly
2. Confirm all issue commands function against real Jira API
3. Validate Markdown to ADF conversion is accepted by Jira
4. Validate ADF to Markdown conversion produces readable output
5. Test error handling with invalid inputs
6. Document any issues or edge cases discovered

## Scope

In Scope:

- Authentication via environment variables
- Project listing
- Issue CRUD operations (list, view, create, edit, delete)
- Issue assignment and workflow transitions
- Comment functionality
- JSON output mode
- Error messages for common failure cases

Out of Scope:

- Performance testing
- Load testing
- Security testing
- Automated CI integration tests

## Prerequisites

Before starting, ensure:

- JIRA_BASE_URL environment variable is set
- JIRA_EMAIL environment variable is set
- JIRA_API_TOKEN environment variable is set
- JIRA_PROJECT environment variable is set (or use -p flag)
- User has permission to create/edit/delete issues in the test project

## Success Criteria

Phase 1: Setup and Authentication

- [x] Environment variables are configured
- [x] `ajira me` returns current user info
- [x] `ajira project list` returns accessible projects

Phase 2: Issue Listing and Viewing

- [x] `ajira issue list` returns issues from default project
- [x] `ajira issue list -q "JQL"` filters correctly
- [x] `ajira issue list -s "Status"` filters by status
- [x] `ajira issue list -t "Type"` filters by type
- [x] `ajira issue list -a "user"` filters by assignee
- [x] `ajira issue list -a unassigned` shows unassigned issues
- [x] `ajira issue list -l 5` limits results
- [x] `ajira issue list --json` outputs valid JSON
- [x] `ajira issue view ISSUE-KEY` displays issue details
- [x] `ajira issue view ISSUE-KEY --json` outputs valid JSON
- [x] `ajira issue view ISSUE-KEY -c 0` hides comments
- [x] `ajira issue view ISSUE-KEY -c 10` shows more comments

Phase 3: Issue Creation

- [x] `ajira issue create -s "Summary"` creates issue, returns URL
- [x] `ajira issue create -s "Summary" -b "Description"` includes body
- [ ] `ajira issue create -s "Summary" -t Bug` sets issue type
- [ ] `ajira issue create -s "Summary" --priority Major` sets priority
- [ ] `ajira issue create -s "Summary" --labels a,b` sets labels
- [x] Created issue description displays correctly in Jira UI
- [x] Markdown formatting (bold, italic, lists, code) renders in Jira

Phase 4: Issue Editing

- [ ] `ajira issue edit ISSUE-KEY -s "New Summary"` updates summary
- [ ] `ajira issue edit ISSUE-KEY -b "New Description"` updates description
- [ ] `ajira issue edit ISSUE-KEY -t Task` changes type
- [ ] `ajira issue edit ISSUE-KEY --priority Low` changes priority
- [ ] `ajira issue edit ISSUE-KEY --labels x,y` replaces labels
- [ ] Edited description renders correctly in Jira UI

Phase 5: Issue Assignment

- [ ] `ajira issue assign ISSUE-KEY user@email` assigns by email
- [ ] `ajira issue assign ISSUE-KEY unassigned` removes assignee
- [ ] Assignment reflected in Jira UI

Phase 6: Issue Transitions

- [ ] `ajira issue move ISSUE-KEY` lists available transitions
- [ ] `ajira issue move ISSUE-KEY "Status"` transitions issue
- [ ] Status change reflected in Jira UI

Phase 7: Comments

- [ ] `ajira issue comment add ISSUE-KEY "text"` adds comment
- [ ] `ajira issue comment add ISSUE-KEY -b "text"` adds via flag
- [ ] Comment appears in Jira UI with correct formatting
- [ ] `ajira issue view ISSUE-KEY` displays comments
- [ ] Comment Markdown renders correctly

Phase 8: Issue Deletion

- [ ] `ajira issue delete ISSUE-KEY` deletes issue
- [ ] Issue no longer accessible in Jira

Phase 9: Error Handling

- [ ] Invalid credentials show clear error
- [ ] Non-existent issue shows "not found" error
- [ ] Invalid JQL shows clear error
- [ ] Missing required fields show clear error
- [ ] Permission denied shows clear error

Phase 10: Edge Cases

- [ ] Long issue summaries handled correctly
- [ ] Special characters in text handled correctly
- [ ] Empty description handled correctly
- [ ] Complex Markdown (tables, code blocks, nested lists) converts correctly

## Deliverables

- All success criteria checkboxes completed
- List of discovered issues (if any)
- Recommendations for improvements (if any)

## Testing Approach

Each test will be performed interactively:

1. Run the command
2. Verify CLI output
3. Verify result in Jira UI (where applicable)
4. Confirm with user before marking complete
5. Document any issues encountered

## Dependencies

- P-003: Markdown/ADF Conversion (completed)
- P-004: Issue Commands (completed)
- P-005: Comment Functionality (completed)
- Access to a Jira Cloud instance with API access

## Notes

Test issues created during this project should be deleted after testing is complete to avoid clutter in the Jira project.

## Issues Found and Fixed

### Phase 1

1. **Project list missing LEAD data** - The project search API was not returning lead information. Fixed by adding `expand=lead` parameter to the API call in `internal/cli/project.go:122`.

2. **Column rename TYPE to STYLE** - Renamed the TYPE column to STYLE to accurately reflect the field name from the Jira API (`style` field indicates classic vs next-gen).

### Phase 2

1. **Jira search API deprecated** - The `/search` endpoint was removed (410 Gone). Migrated to `/search/jql` endpoint with token-based pagination (`nextPageToken`) instead of offset-based (`startAt`).

2. **Added color support** - Added `github.com/fatih/color` for terminal colors. Colors auto-disable when piped. Respects `NO_COLOR` env var.

3. **Status coloring by category** - Uses Jira's `statusCategory.key` for automatic coloring: `done` (green), `indeterminate` (blue), `new` (faint). Override for "Blocked", "On Hold" → yellow.

4. **Column alignment with colors** - Replaced tabwriter with manual padding to fix alignment issues caused by ANSI escape codes.

5. **Added "me" alias for assignee filter** - Added support for `-a me` which uses Jira's `currentUser()` function. Case insensitive. When `-q` (raw JQL) is provided, other filters including `-a me` are silently ignored (expected behaviour).

6. **Added glamour markdown rendering** - Integrated `github.com/charmbracelet/glamour` for terminal-styled markdown output. Description and comments now render with syntax highlighting, styled headers, and proper formatting. Auto-detects terminal width and dark/light theme. Falls back to plain text when piped.

7. **Changed default comment count to 0** - Issue view now hides comments by default for cleaner output. Use `-c N` to show N recent comments.

8. **Output clickable URLs** - Commands that modify issues (create, edit, assign, move, comment add) now output the issue URL instead of just the key. Added `IssueURL()` helper function. Delete still outputs `KEY deleted` since the URL won't work after deletion.

9. **Added field metadata commands** - Implemented `ajira issue priority`, `ajira issue type`, and `ajira issue status` for discovering valid field values. See DR-008.

### Phase 3

1. **ADF code mark compatibility** - Per ADF spec, the `code` mark can ONLY combine with `link` mark. Combinations like `**`code`**` (bold+code) are invalid. Fixed converter to skip incompatible marks when a node has a `code` mark. Added tests for this behavior.

2. **ADF taskItem structure** - Per ADF spec, `taskItem` content must be inline nodes directly, not wrapped in paragraphs. Fixed `convertTaskItem()` to output inline nodes correctly. Task lists now work in Jira.

3. **ADF nested blockquotes not supported** - Per ADF spec, blockquote content can only contain paragraphs, lists, code blocks, and media - NOT other blockquotes. Nested blockquotes like `> > text` are invalid ADF. Documented limitation in testdata file.

4. **Round-trip conversion improvements** - Fixed multiple escaping issues in ADF→MD conversion:
   - Removed over-escaping of pipes and backslashes
   - Fixed underscore escaping to only escape at word boundaries
   - Added merging of adjacent text nodes to prevent goldmark's underscore splitting from causing over-escaping
   - Fixed detection of already-escaped characters to prevent double-escaping
   - Round-trip now preserves content accurately with only cosmetic differences (indentation style, table separators)
   - Created `testdata/comprehensive-markdown.md` for testing all Markdown features
   - Updated DR-007 with full ADF specification constraints and escaping strategy
