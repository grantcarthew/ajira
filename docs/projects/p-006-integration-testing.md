# P-006: Integration Testing

- Status: Proposed
- Started:
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

- JIRA_URL environment variable is set
- JIRA_EMAIL environment variable is set
- JIRA_API_TOKEN environment variable is set
- JIRA_PROJECT environment variable is set (or use -p flag)
- User has permission to create/edit/delete issues in the test project

## Success Criteria

Phase 1: Setup and Authentication

- [ ] Environment variables are configured
- [ ] `ajira me` returns current user info
- [ ] `ajira project list` returns accessible projects

Phase 2: Issue Listing and Viewing

- [ ] `ajira issue list` returns issues from default project
- [ ] `ajira issue list -q "JQL"` filters correctly
- [ ] `ajira issue list -s "Status"` filters by status
- [ ] `ajira issue list -t "Type"` filters by type
- [ ] `ajira issue list -a "user"` filters by assignee
- [ ] `ajira issue list -a unassigned` shows unassigned issues
- [ ] `ajira issue list -l 5` limits results
- [ ] `ajira issue list --json` outputs valid JSON
- [ ] `ajira issue view ISSUE-KEY` displays issue details
- [ ] `ajira issue view ISSUE-KEY --json` outputs valid JSON
- [ ] `ajira issue view ISSUE-KEY -c 0` hides comments
- [ ] `ajira issue view ISSUE-KEY -c 10` shows more comments

Phase 3: Issue Creation

- [ ] `ajira issue create -s "Summary"` creates issue, returns key
- [ ] `ajira issue create -s "Summary" -b "Description"` includes body
- [ ] `ajira issue create -s "Summary" -t Bug` sets issue type
- [ ] `ajira issue create -s "Summary" --priority High` sets priority
- [ ] `ajira issue create -s "Summary" --labels a,b` sets labels
- [ ] Created issue description displays correctly in Jira UI
- [ ] Markdown formatting (bold, italic, lists, code) renders in Jira

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
