# DR-014: CLI Automation Support

- Date: 2026-01-14
- Status: Accepted
- Category: cli

## Problem

ajira is designed for AI agents and automation, but lacks features that make it robust for scripting and CI/CD pipelines:

- Single exit code (1) for all errors - scripts cannot distinguish error types
- No way to preview actions before execution
- No debug output for troubleshooting API issues
- No batch processing for multiple issues
- No rate limit handling - fails immediately on 429

## Decision

Add automation-focused features: standardised exit codes, global flags for dry-run/verbose/quiet/no-color, batch stdin processing, and automatic rate limit retry.

## Exit Codes

| Code | Name | Description |
| ---- | ---- | ----------- |
| 0 | Success | Successful execution |
| 1 | UserError | User/input error (invalid args, missing values) |
| 2 | APIError | API error (4xx/5xx responses, except auth) |
| 3 | NetError | Network/connection error |
| 4 | AuthError | Authentication error (401, 403) |
| 5 | Partial | Partial failure in batch operations |

## Global Flags

| Flag | Short | Description |
| ---- | ----- | ----------- |
| --dry-run | - | Show planned actions without executing |
| --verbose | - | Show HTTP request/response details to stderr |
| --quiet | - | Suppress non-essential output (errors still shown) |
| --no-color | - | Disable ANSI colours even in TTY contexts |

Note: --verbose has no short flag because -v is used by --version. --quiet has no short flag because -q is used by --query in issue list and project list commands.

## Dry-Run Behaviour

- No API calls made - shows intended action based on inputs only
- Respects --json flag for output format
- Multi-step commands show consolidated action (e.g., "Would transition PROJ-123 to Done and assign to user@example.com")
- Batch operations show all planned actions upfront before any would execute

## Verbose Output

- Written to stderr (keeps stdout clean for parsing)
- Simplified format: method, URL, status code, duration
- Example: `GET /rest/api/3/issue/PROJ-123 200 OK (142ms)`

## Quiet Mode

- Suppresses success output (URLs, confirmations)
- Errors always shown - scripts need to know why failures occurred
- Exit codes still set correctly

## Batch Stdin Processing

Commands supporting --stdin flag:

- issue assign
- issue move
- issue delete
- issue comment add
- sprint add
- epic add
- epic remove

Behaviour:

- Reads issue keys from stdin (one per line)
- Processes all keys, continues on failure
- Reports individual results plus summary
- Returns exit code 5 (Partial) if some succeed and some fail
- Error if both --stdin and positional key arguments provided
- Error if --stdin combined with --file - (mutual exclusion)

Output format:

```
PROJ-123: success
PROJ-124: failed - Issue not found
PROJ-125: success

3 processed: 2 succeeded, 1 failed
```

## Rate Limiting

- Detects 429 Too Many Requests responses
- Automatic retry with exponential backoff
- Fixed defaults: 3 retries, starting at 1 second
- Uses Retry-After header when present
- Verbose mode shows retry attempts

## Why

Exit codes: Scripts need to handle different error types differently. Auth errors may need credential refresh, network errors may warrant retry, user errors need input correction.

Dry-run: Safe testing of automation scripts before production execution. Essential for CI/CD pipelines and destructive operations like delete.

Verbose: Debugging API issues without adding application-level logging. Shows exactly what HTTP calls are made.

Quiet: CI/CD logs should show only failures. Success output clutters logs and slows parsing.

Batch stdin: Processing hundreds of issues one command at a time is slow and awkward. Piping from other tools (grep, jq) enables powerful workflows.

Rate limiting: Jira Cloud enforces rate limits. Failing immediately wastes previous work and requires manual retry. Automatic backoff handles transient limits gracefully.

## Trade-offs

Accept:

- More complex error handling in main.go
- Package-level variables for global flags (matches existing pattern)
- Fixed retry configuration (not user-configurable)
- No validation in dry-run mode (may miss some errors until actual execution)

Gain:

- Scripts can handle errors appropriately by type
- Safe testing of destructive operations
- Debug output without code changes
- Efficient batch processing
- Resilient to rate limiting

## Alternatives

Verbose short flag options:

- -V (uppercase): Distinguishes from version but unconventional
- No short flag: Chosen for simplicity, --verbose is clear
- Change version to -V: Breaking change, rejected

Dry-run with validation:

- Make read-only API calls to validate inputs exist
- Rejected: Adds complexity, requires network, violates KISS principle

Batch failure modes:

- Stop on first failure: Rejected - wastes previous successes, requires multiple runs
- --continue-on-error flag: Rejected - adds complexity, continue-all is better default for automation

Rate limit configuration:

- Flags (--max-retries, --retry-delay): Rejected - flag bloat, most users don't need tuning
- Environment variables: Rejected - unnecessary complexity
- Fixed defaults: Chosen - sensible defaults cover most cases

Flag propagation:

- Context values: Idiomatic Go, thread-safe
- Package-level variables: Chosen - matches existing jsonOutput/project/board pattern, simpler

## Usage Examples

Exit code handling in bash:

```bash
ajira issue move PROJ-123 "Done"
case $? in
  0) echo "Success" ;;
  1) echo "Invalid input" ;;
  2) echo "API error" ;;
  3) echo "Network error - retry later" ;;
  4) echo "Auth failed - check credentials" ;;
esac
```

Dry-run before bulk operation:

```bash
echo -e "PROJ-123\nPROJ-124\nPROJ-125" | ajira issue move --stdin "Done" --dry-run
# Review output, then run without --dry-run
```

Debug API issues:

```bash
ajira issue view PROJ-123 --verbose 2>&1 | grep "GET\|POST"
```

Batch assign from JQL:

```bash
ajira issue list -q "project = PROJ AND status = 'To Do'" --json | \
  jq -r '.[].key' | \
  ajira issue assign --stdin me
```

CI/CD quiet mode:

```bash
ajira issue move PROJ-123 "Done" --quiet || echo "Move failed"
```
