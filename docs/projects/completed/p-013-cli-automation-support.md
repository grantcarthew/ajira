# p-013: Automation Support

- Status: Completed
- Started: 2026-01-14
- Completed: 2026-01-15
- Design: dr-014-cli-automation-support.md

## Overview

Add features specifically designed for scripting and automation use cases. This includes standardised exit codes, dry-run mode, verbose output, and batch operations. These features make ajira more robust for CI/CD pipelines and automated workflows.

## Goals

1. Implement standardised exit codes
2. Add dry-run mode for safe testing
3. Add verbose mode for debugging
4. Implement batch operations via stdin
5. Add rate limit awareness
6. Document automation best practices

## Scope

In Scope:

Exit codes:

- 0: Success
- 1: User/input error
- 2: API error
- 3: Network/connection error
- 4: Authentication error
- Document exit codes in help and README

Global flags:

- `--dry-run` - Show what would happen without executing
- `--verbose` - Show API requests/responses (no short flag, -v is version)
- `--quiet` - Suppress non-essential output (no short flag, -q is used by --query)
- `--no-color` - Disable ANSI colours for TTY contexts wanting plain output

Batch operations:

- Accept issue keys from stdin for applicable commands
- `--stdin` flag to read from pipe
- Support: assign, move, delete, comment add, sprint add, epic add, epic remove

Rate limiting:

- Detect 429 responses
- Automatic retry with backoff (3 retries, exponential)

Out of Scope:

- Interactive retry prompts (non-interactive CLI)
- Configuration file for automation settings
- Webhook integration

## Success Criteria

- [x] Exit codes are consistent and documented
- [x] `--dry-run` shows planned actions without executing
- [x] `--verbose` displays HTTP request/response details
- [x] `--quiet` reduces output to essentials only
- [x] Commands accept issue keys from stdin with `--stdin`
- [x] Rate limiting triggers automatic retry with backoff
- [x] All exit code scenarios are tested
- [x] Documentation includes automation examples

## Deliverables

- Updated `internal/cli/root.go` - Global flags
- `internal/cli/exit_codes.go` - Exit code constants and handling
- Updated `internal/api/client.go` - Verbose logging, rate limiting
- Updated commands to support stdin and dry-run:
  - `issue assign` - batch assign via stdin
  - `issue move` - batch move via stdin
  - `issue delete` - batch delete via stdin
  - `issue comment add` - batch comment via stdin
  - `sprint add` - already supports multiple keys, add stdin
  - `epic add` - already supports multiple keys, add stdin
  - `epic remove` - already supports multiple keys, add stdin
- Updated `docs/flags-and-arguments.md` - Document new global flags
- `docs/cli/automation.md` - Automation guide with examples
- dr-014: Scripting and Automation Conventions
- Integration tests for exit codes
- Tests for rate limit handling
- Tests for batch operations

## Technical Approach

Exit code implementation:

1. Define constants for exit codes
2. Create typed errors that carry exit codes
3. Handle at root command level
4. Map API errors to appropriate codes

Dry-run implementation:

1. Pass dry-run flag through context
2. Commands check flag before API calls
3. Print intended action instead of executing
4. Return success without side effects

Stdin batch processing:

1. Check if stdin is a pipe
2. Read issue keys (one per line)
3. Process each with same operation
4. Report success/failure per issue

## Current State

Exit codes:

- `main.go` uses `os.Exit(1)` for any error from `cli.Execute()`
- No differentiation between error types
- Commands return `error` which Cobra handles generically

Verbose/debug mode:

- No verbose logging capability exists
- HTTP client (`internal/api/client.go`) has no request/response logging
- No way to see API calls being made

Quiet mode:

- Not implemented
- All output goes to stdout via `fmt.Println`
- No abstraction layer for suppressible output

Dry-run mode:

- Not implemented
- Commands execute API calls directly
- No mechanism to preview actions

Stdin usage:

- Currently used for content input via `--file -` (description, comment text)
- NOT used for batch issue keys
- The `--stdin` flag for batch keys would be distinct from existing `--file -` usage

Multiple issue key arguments:

- `sprint add <sprint-id> <issue-keys...>` already accepts multiple keys
- `epic add <epic-key> <issue-keys...>` already accepts multiple keys
- `epic remove <issue-keys...>` already accepts multiple keys
- These commands could also benefit from `--stdin` for large batches

Rate limiting:

- No rate limit handling
- HTTP client doesn't check for 429 responses
- No retry logic
- Note: Both v3 API and Agile API share the same `doRequest` method, so rate limiting implementation will apply to all endpoints

Colour output:

- Uses `glamour` for markdown with terminal styling
- Has TTY detection - non-TTY output is already plain
- No explicit `--no-color` flag

Global flags (from dr-004):

- `--json, -j` - JSON output format
- `--project, -p` - Default project key
- `--board` - Default board ID (no short flag)
- `--version, -v` - Version information (CONFLICT with proposed --verbose)
- `--help, -h` - Help

API error structure (`internal/api/client.go`):

- `APIError` type contains: StatusCode, Status, Messages, Errors, RawBody, Method, Path
- Can inspect `StatusCode` to differentiate error types:
  - 401: Authentication error (exit code 4)
  - 403: Permission denied (could be exit code 4 or 2)
  - 404: Not found (exit code 2)
  - 429: Rate limited (handled with retry, or exit code 2 if exhausted)
  - 5xx: Server error (exit code 2)
- Network errors from `http.Client.Do()` are wrapped errors (exit code 3)
- Config errors from `config.Load()` are validation errors (exit code 1)

Short flag availability (from `docs/flags-and-arguments.md`):

- Reserved globally: `-j`, `-p`, `-v` (version), `-h`
- `-q` is available globally (proposed for --quiet)
- `-V` (uppercase) is available globally (option for --verbose)

## Research Areas

Completed research:

Go exit codes:

- Reserve `os.Exit` for main function only
- Return errors up the call stack; let main decide if fatal
- Use typed errors that carry exit code information

stdin pipe detection:

- Use `os.Stdin.Stat()` and check `os.ModeNamedPipe` bit
- `(stat.Mode() & os.ModeNamedPipe) != 0` indicates piped input

Jira Cloud rate limit headers:

- `X-RateLimit-Limit` - Total requests allowed in current window
- `X-RateLimit-Remaining` - Requests left before limit reached
- `X-RateLimit-Reset` - ISO 8601 timestamp when window resets
- `Retry-After` - Seconds (or ISO 8601) until safe to retry
- `RateLimit-Reason` - Context about why limit was hit (burst, quota, etc.)
- Returns `429 Too Many Requests` when limit exceeded
- Note: Headers may not appear on every response; only guaranteed on 429

## Dependencies

None - enhances existing infrastructure.
