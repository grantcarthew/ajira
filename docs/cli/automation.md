# Automation Guide

This guide covers ajira features designed for scripting and CI/CD pipelines.

## Exit Codes

ajira uses standardised exit codes to help scripts handle different error conditions:

| Code | Name | Description |
|------|------|-------------|
| 0 | Success | Successful execution |
| 1 | UserError | User/input error (invalid arguments, missing values) |
| 2 | APIError | API error (4xx/5xx responses, except auth) |
| 3 | NetError | Network/connection error |
| 4 | AuthError | Authentication error (401, 403) |
| 5 | Partial | Partial failure in batch operations |

Example usage in bash:

```bash
ajira issue move PROJ-123 "Done"
case $? in
  0) echo "Success" ;;
  1) echo "Invalid input - check arguments" ;;
  2) echo "API error - issue may not exist" ;;
  3) echo "Network error - retry later" ;;
  4) echo "Auth failed - check credentials" ;;
  5) echo "Partial failure - check output" ;;
esac
```

## Global Automation Flags

### --dry-run

Preview what would happen without making any changes:

```bash
ajira issue delete PROJ-123 --dry-run
# Output: Would delete PROJ-123

ajira issue assign PROJ-123 me --dry-run
# Output: Would assign PROJ-123 to me
```

Dry-run respects the --json flag:

```bash
ajira issue delete PROJ-123 --dry-run --json
# Output: {"action": "delete PROJ-123"}
```

### --verbose

Show HTTP request/response details for debugging:

```bash
ajira issue view PROJ-123 --verbose
# Output to stderr: GET /rest/api/3/issue/PROJ-123 200 OK (142ms)
# Output to stdout: (issue details)
```

Verbose output goes to stderr, keeping stdout clean for parsing:

```bash
ajira issue list -l 5 --json --verbose 2>/dev/null | jq '.[].key'
```

### --quiet

Suppress success output (errors still shown):

```bash
ajira issue assign PROJ-123 me --quiet
# No output on success

ajira issue assign INVALID-999 me --quiet
# Error: API error: ...
```

### --no-color

Disable coloured output even in TTY contexts:

```bash
ajira issue view PROJ-123 --no-color > issue.txt
```

## Batch Operations

Several commands support reading issue keys from stdin using --stdin:

- `issue assign`
- `issue move`
- `issue delete`
- `issue comment add`
- `sprint add`
- `epic add`
- `epic remove`

### Basic Batch Usage

```bash
echo -e "PROJ-1\nPROJ-2\nPROJ-3" | ajira issue assign --stdin me
```

Output format:

```
PROJ-1: success
PROJ-2: success
PROJ-3: failed - Issue does not exist

3 processed: 2 succeeded, 1 failed
```

### Combining with JQL

```bash
ajira issue list -q "status = 'To Do'" --json | \
  jq -r '.[].key' | \
  ajira issue move --stdin "In Progress"
```

### Batch with Dry-Run

Preview batch operations before executing:

```bash
ajira issue list -q "assignee = me" --json | \
  jq -r '.[].key' | \
  ajira issue move --stdin Done --dry-run
```

### Batch Delete

```bash
ajira issue list -q "status = Done AND updated < -30d" --json | \
  jq -r '.[].key' | \
  ajira issue delete --stdin --cascade
```

## Rate Limiting

ajira automatically handles Jira Cloud rate limits:

- Detects 429 Too Many Requests responses
- Retries up to 3 times with exponential backoff
- Uses Retry-After header when present
- Verbose mode shows retry attempts

Example with verbose:

```bash
ajira issue list -l 1000 --verbose 2>&1 | grep -E "(GET|Rate limited)"
# GET /rest/api/3/search 429 Too Many Requests (45ms)
# Rate limited, retrying in 1s (attempt 1/3)
# GET /rest/api/3/search 200 OK (234ms)
```

## CI/CD Examples

### GitHub Actions

```yaml
- name: Move issues to Done
  run: |
    echo "$ISSUE_KEYS" | ajira issue move --stdin Done --quiet
  env:
    JIRA_BASE_URL: ${{ secrets.JIRA_BASE_URL }}
    JIRA_EMAIL: ${{ secrets.JIRA_EMAIL }}
    JIRA_API_TOKEN: ${{ secrets.JIRA_API_TOKEN }}
```

### Shell Script with Error Handling

```bash
#!/bin/bash
set -e

# Assign issues, handle partial failures
if ! ajira issue list -q "sprint = currentSprint()" --json | \
     jq -r '.[].key' | \
     ajira issue assign --stdin me; then
  exit_code=$?
  if [ $exit_code -eq 5 ]; then
    echo "Some assignments failed - check output"
  else
    echo "Assignment failed with code $exit_code"
    exit $exit_code
  fi
fi
```

### Combining Multiple Operations

```bash
# Create issue and capture key
KEY=$(ajira issue create -s "Automated task" --json | jq -r .key)

# Assign and move
ajira issue assign "$KEY" me --quiet
ajira issue move "$KEY" "In Progress" --quiet

echo "Created and started: $KEY"
```

## stdin Conflict

Commands that use --file - for content input cannot also use --stdin for batch keys (both read from stdin). Error is returned if both are used.

Valid:
```bash
echo -e "PROJ-1\nPROJ-2" | ajira issue comment add --stdin "Comment text"
ajira issue comment add PROJ-123 -f comment.md
```

Invalid:
```bash
echo "keys" | ajira issue comment add --stdin -f -
# Error: cannot use --stdin with --file - (both read from stdin)
```

Workaround: Save content to a file first, then use --stdin for keys.
