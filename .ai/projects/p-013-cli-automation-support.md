# p-013: Automation Support

- Status: Pending
- Started:
- Completed:

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
- `--verbose`, `-v` - Show API requests/responses
- `--quiet`, `-q` - Suppress non-essential output
- `--no-color` - Disable ANSI colours for TTY contexts wanting plain output

Batch operations:

- Accept issue keys from stdin for applicable commands
- `--stdin` flag to read from pipe
- Support: assign, move, delete, comment add, sprint add, epic add, epic remove

Rate limiting:

- Detect 429 responses
- Automatic retry with backoff
- `--rate-limit-info` - Show remaining API quota

Out of Scope:

- Interactive retry prompts (non-interactive CLI)
- Configuration file for automation settings
- Webhook integration

## Success Criteria

- [ ] Exit codes are consistent and documented
- [ ] `--dry-run` shows planned actions without executing
- [ ] `--verbose` displays HTTP request/response details
- [ ] `--quiet` reduces output to essentials only
- [ ] Commands accept issue keys from stdin with `--stdin`
- [ ] Rate limiting triggers automatic retry with backoff
- [ ] All exit code scenarios are tested
- [ ] Documentation includes automation examples

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

## Decision Points

1. Verbose short flag conflict

dr-004 assigns `-v` to `--version`. Options for `--verbose`:

- A: Use `-V` (uppercase) for verbose
- B: Use no short flag for verbose
- C: Change version to `-V` and use `-v` for verbose (breaking change)

2. Dry-run validation behaviour

Should `--dry-run` make read-only API calls to validate inputs?
- A: No API calls - just show intended action based on inputs
- B: Validate inputs (user exists, status available) but don't mutate
- Trade-off: Option B is more useful but slower and requires network

3. Batch operation failure behaviour

When processing multiple issue keys via `--stdin`:
- A: Continue all, report all failures at end
- B: Stop on first failure
- C: Add `--continue-on-error` flag (default: stop)

4. Exit code for partial batch failures

If batch has some successes and some failures:
- A: Return failure exit code (non-zero)
- B: Return success if any succeeded
- C: New exit code (e.g., 5) for partial failure

5. Verbose output format

How to format HTTP request/response details:
- A: Full wire format (headers, body, timing)
- B: Simplified (method, URL, status code, duration)
- C: Structured log format with levels

6. --no-color necessity

Non-TTY output is already plain (glamour disabled). Is `--no-color` still needed?
- A: Yes - for TTY contexts that want plain output
- B: No - existing behaviour is sufficient

7. Multi-step command dry-run

Commands like `clone` make multiple API calls (create, then link). In dry-run:
- A: Show all planned actions as a sequence
- B: Show only the primary action

8. Verbose output destination

Where should verbose HTTP logging be written?
- A: stderr (conventional for debug/diagnostic output)
- B: stdout (mixed with command output)

9. Quiet mode and error messages

Should `--quiet` suppress error messages?
- A: No - errors always shown (quiet only affects success output)
- B: Yes - suppress everything except exit code

10. Dry-run output format

Should `--dry-run` output respect `--json` flag?
- A: Yes - dry-run output follows same format rules
- B: No - dry-run always human-readable regardless of --json

11. Flag propagation mechanism

How should global automation flags (dry-run, verbose, quiet) be passed to commands?
- A: Context values (idiomatic Go, thread-safe)
- B: Package-level variables (simpler, matches current --json pattern)

12. Rate limit retry configuration

Should retry behaviour be configurable?
- A: Fixed sensible defaults (3 retries, exponential backoff)
- B: Configurable via flags (--max-retries, --retry-delay)
- C: Configurable via environment variables

13. Stdin conflict with --file flag

Commands using `--file -` for content already consume stdin. When using `--stdin` for batch keys:
- A: Mutual exclusion - error if both are used
- B: Read keys from file with new flag (e.g., `--keys-file`)
- Note: Affects `comment add` which has both `--file` and would need `--stdin`

14. Batch dry-run sequence

How should `--dry-run` behave with batch operations via `--stdin`?
- A: Show all planned actions upfront before any execution
- B: Process sequentially, showing each planned action as it would execute
- Note: Option A gives clearer overview; Option B mirrors actual execution flow

15. Batch result reporting

How should batch operation results be reported?
- A: List each result (success/failure per key)
- B: Summary only (e.g., "3 of 5 succeeded")
- C: Both - individual results plus summary

16. Rate limit info source

The `--rate-limit-info` flag - what should it display?
- A: Make an API call specifically to check remaining quota
- B: Display cached info from the last API request in this session
- C: Make a lightweight API call (e.g., GET /myself) to fetch current limits
- Note: Jira rate limit headers are returned with every response

17. Existing multi-key commands stdin behaviour

Commands already accepting multiple keys via args (sprint add, epic add, epic remove):
- A: `--stdin` replaces all positional keys (args ignored if --stdin present)
- B: `--stdin` supplements positional keys (combine both)
- Note: Option A is simpler and avoids confusion

18. Move command multi-step dry-run

The `issue move` command with `--assignee` makes multiple API calls (resolve user, then transition). In dry-run:
- A: Show both steps - "Would resolve user X to account Y" then "Would transition PROJ-123 to Done"
- B: Show consolidated action - "Would transition PROJ-123 to Done and assign to X"
- Note: Same consideration applies to any command with resolution steps

## Dependencies

None - enhances existing infrastructure.
