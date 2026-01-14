# p-014: Auxiliary Commands

- Status: Proposed
- Started:
- Completed:

## Overview

Add utility commands that complement the core issue functionality. This includes watch management, release/version listing, browser navigation, and user search. These are lower-priority but useful features that round out the CLI.

## Goals

1. Implement watch/unwatch commands
2. Implement release list command
3. Implement open command for browser navigation
4. Implement user search command
5. Implement field discovery command

## Scope

In Scope:

Watch commands:

- `ajira issue watch <key>` - Start watching an issue
- `ajira issue unwatch <key>` - Stop watching an issue

Release commands:

- `ajira release list` - List project versions/releases
- Filter by released/unreleased status

Open command:

- `ajira open` - Open project in browser
- `ajira open <key>` - Open issue in browser
- Print URL if browser cannot be opened

User commands:

- `ajira user search <query>` - Search users by name/email
- Useful for finding account IDs for assignment

Field discovery:

- `ajira field list` - List available fields
- `ajira field list <project>` - List fields for project
- Include custom fields with their IDs

Out of Scope:

- User management (admin function)
- Release creation/management (typically done in UI)
- Field creation (admin function)

## Success Criteria

- [ ] `issue watch` adds current user as watcher
- [ ] `issue unwatch` removes current user as watcher
- [ ] `release list` shows versions with name, status, dates
- [ ] `open` launches browser or prints URL
- [ ] `user search` finds users matching query
- [ ] `field list` displays fields with IDs
- [ ] All commands support `--json` output
- [ ] Tests cover all new commands

## Deliverables

- `internal/cli/issue_watch.go` - Watch/unwatch implementation
- `internal/cli/release.go` - Release list command
- `internal/cli/open.go` - Browser open command
- `internal/cli/user.go` - User search command
- `internal/cli/field.go` - Field list command
- dr-015: Utility Command Patterns (if needed)
- Integration tests for new commands

## Research Areas

- Jira watcher API
- Project versions API
- Cross-platform browser opening in Go
- Field metadata API (including custom fields)

## Questions and Uncertainties

- How to detect if running in headless/SSH environment for open?
- Should field list include all fields or just editable ones?
- How to format custom field IDs for use in other commands?
- Should user search support multiple results or just first?

## Dependencies

None - independent utility commands.
