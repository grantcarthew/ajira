# P-013: Automation Support

- Status: Proposed
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
- `--no-color` - Disable ANSI colours (alternative to --plain)

Batch operations:
- Accept issue keys from stdin for applicable commands
- `--stdin` flag to read from pipe
- Support: assign, move, delete, comment add

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
- Updated commands to support stdin and dry-run
- `docs/cli/automation.md` - Automation guide
- DR-014: Scripting and Automation Conventions
- Integration tests for exit codes
- Tests for rate limit handling

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

## Research Areas

- Go exit code best practices
- HTTP client verbose logging
- Jira API rate limits
- stdin detection in Go

## Questions and Uncertainties

- Should dry-run validate fields against API?
- How to format verbose HTTP output?
- Should batch operations continue on individual failures?
- What's the Jira Cloud rate limit?

## Dependencies

None - enhances existing infrastructure.
