# dr-010: Issue List Filtering and Sorting

- Date: 2026-01-09
- Status: Accepted
- Category: CLI

## Problem

The `issue list` command has basic filters (status, type, assignee) but lacks:

- Reporter filter
- Priority filter
- Labels filter
- Watching filter
- Sorting control

Agents can use raw JQL via `--query`, but convenience flags reduce token usage and simplify common queries.

## Decision

Add new filter and sorting flags to `issue list`:

Filter flags:

| Flag | Short | Type | JQL Mapping |
|------|-------|------|-------------|
| `--reporter` | `-r` | string | `reporter = "value"` |
| `--priority` | `-P` | string | `priority = "value"` |
| `--labels` | `-L` | []string | `labels IN (...)` |
| `--watching` | `-w` | bool | `watcher = currentUser()` |

Sorting flags:

| Flag | Type | Description |
|------|------|-------------|
| `--order-by` | string | Sort field: created, updated, priority, key, rank |
| `--reverse` | bool | Use ASC instead of DESC |

## Why

- Convenience flags reduce JQL syntax errors
- Shorter commands save tokens for AI agents
- Consistent with existing filter pattern (status, type, assignee)
- Labels as slice allows `--labels bug,urgent` syntax

## JQL Generation

Filters combine with AND logic:

```
ajira issue list --status "In Progress" -r me --labels bug
```

Generates:

```
project = PROJ AND status = "In Progress" AND reporter = currentUser() AND labels IN ("bug") ORDER BY updated DESC
```

Sorting replaces default ORDER BY:

```
ajira issue list --order-by created --reverse
```

Generates:

```
project = PROJ ORDER BY created ASC
```

## Trade-offs

Accept:

- More flags to document and maintain
- Uppercase short flags (-P, -L) less discoverable than lowercase

Gain:

- Reduced token usage for common queries
- Fewer JQL syntax errors
- Consistent filtering interface

## Alternatives

Date filters (--created, --updated):

- Considered relative formats (-7d, week, month)
- Rejected: Adds complexity; agents can use JQL for date filtering
- JQL date functions (startOfWeek, -7d) are well-documented

Output format flags (--plain, --csv, --columns):

- Considered for scripting flexibility
- Rejected: Agents use --json; colour auto-disables on pipe
- Keep CLI focused on agent use case
