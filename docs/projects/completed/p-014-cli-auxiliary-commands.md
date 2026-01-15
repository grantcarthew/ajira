# p-014: Auxiliary Commands

- Status: Complete
- Started: 2026-01-15
- Completed: 2026-01-15

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

- [x] `issue watch` adds current user as watcher
- [x] `issue unwatch` removes current user as watcher
- [x] `release list` shows versions with name, status, dates
- [x] `open` launches browser or prints URL
- [x] `user search` finds users matching query
- [x] `field list` displays fields with IDs
- [x] All commands support `--json` output
- [x] Tests cover all new commands

## Deliverables

- `internal/cli/issue_watch.go` - Watch/unwatch implementation
- `internal/cli/release.go` - Release list command
- `internal/cli/open.go` - Browser open command
- `internal/cli/user.go` - User search command
- `internal/cli/field.go` - Field list command
- dr-016: Utility Command Patterns (if needed)
- Integration tests for new commands

## Dependencies

None - independent utility commands.

## Current State

### Existing Patterns

The codebase follows consistent patterns that this project will adopt:

**Command Structure** (`internal/cli/`):
- Commands use Cobra with `rootCmd.AddCommand()` for root commands or parent subcommands
- Each command file defines a `*cobra.Command` and registers in `init()`
- `SilenceUsage: true` is standard to prevent usage on errors
- Commands accept `--json` via `JSONOutput()` global flag

**API Client** (`internal/api/client.go`):
- `Client` struct with `Get()`, `Post()`, `Put()`, `Delete()` methods
- Base path is `/rest/api/3` for standard Jira API
- Returns `[]byte` and `error`; errors are `*api.APIError` for HTTP failures

**Output Patterns** (`internal/cli/batch.go`):
- `PrintSuccess()`, `PrintSuccessJSON()` for output respecting `--quiet`
- `PrintDryRun()`, `PrintDryRunBatch()` for dry-run mode
- JSON output uses `json.MarshalIndent()` with 2-space indentation

**Testing** (`internal/cli/*_test.go`):
- Uses `httptest.NewServer()` for mock API responses
- `testConfig(serverURL)` helper creates test configuration
- Tests cover success paths, error handling, and edge cases

### Existing Code to Reuse

**Current user** in `internal/cli/me.go:35-73`:
- `runMe()` fetches `/myself` endpoint returning `User` struct
- Contains `AccountID` field needed for watch/unwatch operations
- Pattern: fetch current user, extract accountId for API calls

**User search** already exists in `internal/cli/issue_assign.go:159-187`:
- `resolveUser()` function searches by email or display name
- Uses `/user/search?query=...&maxResults=1` endpoint
- Returns first match only - new `user search` command should show all matches

**URL generation** in `internal/cli/root.go:157-159`:
- `IssueURL(baseURL, key)` returns browse URL for issues
- Can be extended for project URLs

**Tabular output** in `internal/cli/project.go:106-111`:
- Uses `text/tabwriter` for aligned column output
- Pattern: header row, then data rows with tab separators

### Jira API Endpoints

**Watchers** (v3 API):
- `POST /rest/api/3/issue/{issueIdOrKey}/watchers` - Add watcher (body: quoted account ID string, e.g., `"5b10ac8d82e..."`)
- `DELETE /rest/api/3/issue/{issueIdOrKey}/watchers?accountId={accountId}` - Remove watcher (account ID in query param)
- Note: Requires fetching current user's accountId from `/myself` first

**Project Versions** (v3 API):
- `GET /rest/api/3/project/{projectIdOrKey}/version` - Paginated list of versions
- Query params: `status=released|unreleased`, `orderBy`, `maxResults`

**Fields** (v3 API):
- `GET /rest/api/3/field` - List all fields (system and custom)
- Returns array with `id`, `name`, `custom`, `schema` for each field

**User Search** (v3 API):
- `GET /rest/api/3/user/search?query={query}` - Search users
- Returns array of user objects with `accountId`, `displayName`, `emailAddress`

### Browser Opening

Cross-platform browser opening in pure Go using `os/exec`:
- macOS: `exec.Command("open", url)`
- Linux: `exec.Command("xdg-open", url)`
- Windows: `exec.Command("cmd", "/C", "start", "", url)`

Fallback: Print URL to stdout if command fails or not a TTY.

## Decisions

1. **User search result limit**: Use global `--limit` / `-l` flag, default 10
2. **Field list scope**: Show all fields (system + custom)
3. **Open command location**: Both - `ajira open [key]` at root (primary) and `ajira issue open <key>` as alias
4. **Watch batch support**: Yes, support `--stdin` for batch operations (consistent with existing patterns)
5. **Headless/SSH detection**: Use `term.IsTerminal()` check; if not a TTY or browser command fails, print URL to stdout
6. **Custom field ID format**: Display raw field IDs (e.g., `customfield_10001`) as-is - these can be used directly in JQL or API calls
