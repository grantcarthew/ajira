# dr-011: Issue Clone Command

- Date: 2026-01-12
- Status: Accepted
- Category: CLI

## Problem

Users need to duplicate Jira issues for common workflows: creating similar issues, copying templates, or duplicating work items across projects. Jira's REST API has no native clone endpoint, requiring a multi-step process to fetch and recreate issues.

## Decision

Implement `ajira issue clone <key>` command that:

1. Fetches the source issue with all fields
2. Creates a new issue with copied fields
3. Applies any field overrides from flags
4. Optionally creates a link to the original issue
5. Returns the new issue URL (text) or full details (JSON)

## Flags

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--summary` | `-s` | string | Override summary |
| `--assignee` | `-a` | string | Override assignee |
| `--reporter` | `-r` | string | Override reporter |
| `--priority` | `-P` | string | Override priority |
| `--type` | `-t` | string | Override issue type |
| `--labels` | `-L` | string | Override labels (comma-separated) |
| `--link` | | string | Link to original (default: "Clones", or specify type) |
| `--project` | `-p` | string | Target project (global flag) |
| `--json` | `-j` | bool | JSON output (global flag) |

## Why

- No native Jira clone API exists; must implement via GET + POST
- Flags reuse existing patterns from create/edit/list commands for consistency
- Default "clone all fields" matches user expectations of what "clone" means
- Optional `--link` with sensible default ("Clones") mirrors Jira UI behaviour
- Fail-fast validation for cross-project cloning prevents confusing API errors

## Execution Flow

When `ajira issue clone <key>` is executed:

1. Fetch source issue via GET /issue/{key}
2. If `--project` specified and differs from source:
   - Validate issue type exists in target project
   - Fail with error listing valid types if not found
3. Build create request with source fields:
   - summary, description, priority, labels, issue type
4. Apply flag overrides (summary, assignee, reporter, priority, type, labels)
5. POST to create new issue in target project
6. If `--link` specified:
   - Validate link type (default "Clones" if no value)
   - Create link: new issue "clones" original
   - Link direction: inwardIssue = new, outwardIssue = original
7. Output new issue URL (text) or full response (JSON)

## Link Behaviour

| Usage | Behaviour |
|-------|-----------|
| No `--link` flag | No link created |
| `--link` | Link with "Clones" type |
| `--link Duplicate` | Link with specified type (validated) |

If "Clones" type does not exist, fail with error listing available types.

## Trade-offs

Accept:

- No attachment cloning (API complexity, rarely needed)
- No comment cloning (YAGNI, rarely needed)
- No subtask cloning (future consideration)
- No text substitution flag (can use edit command after clone)

Gain:

- Simple, predictable clone behaviour
- Consistent flag patterns with existing commands
- Cross-project cloning with validation
- Optional linking matches Jira UI behaviour

## Alternatives

Text substitution flag (`--replace "find:replace"`):

- Pro: Enables template-style cloning in one command
- Con: Delimiter ambiguity, regex complexity, multiple replacements
- Rejected: Two-step workflow (clone then edit) is sufficient; YAGNI

Comment cloning (`--comments`):

- Pro: Matches Jira UI option
- Con: Rarely needed via CLI, adds complexity
- Rejected: YAGNI; can be added later if demand emerges

Minimal clone (explicit opt-in for each field):

- Pro: Maximum control
- Con: Tedious for common case, violates "clone" semantics
- Rejected: Clone should copy everything by default
