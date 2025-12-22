# DR-006: Project List Command

- Date: 2025-12-22
- Status: Accepted
- Category: CLI

## Problem

Users need to discover available Jira projects to:

- Find project keys for use with other commands
- Verify project access permissions
- Get project metadata (lead, type)

The API returns paginated results, so the command must handle pagination to return all projects.

## Decision

Implement `ajira project list` command using GET /rest/api/3/project/search with automatic pagination.

Command structure:

```
ajira project list [flags]
```

Flags:

| Flag | Short | Type | Default | Description |
| ---- | ----- | ---- | ------- | ----------- |
| --query | -q | string | "" | Filter by project name/key |
| --limit | -l | int | 0 | Maximum projects to return (0 = all) |

## Pagination Behaviour

The command fetches all pages automatically:

1. Request first page with `startAt=0`, `maxResults=50`
2. Check `isLast` field in response
3. If `isLast` is false, request next page with `startAt += maxResults`
4. Repeat until `isLast` is true or `--limit` reached
5. Aggregate all results and output once

This provides a simple user experience - one command returns all projects.

## Output Formats

Plain text (default):

```
KEY     NAME              LEAD         TYPE
EX      Example           Jane Doe     classic
ABC     Alphabetical      John Smith   classic
```

JSON (`ajira project list -j`):

```json
[
  {
    "id": "10000",
    "key": "EX",
    "name": "Example",
    "lead": "Jane Doe",
    "type": "classic"
  }
]
```

## Why

- Automatic pagination hides API complexity from users
- `--limit` provides escape hatch for large instances
- `--query` leverages server-side filtering to reduce API calls
- Tabular plain text is scannable for humans
- JSON array is easy to process programmatically

## Trade-offs

Accept:

- Multiple API calls for large project lists
- Memory usage scales with project count

Gain:

- Simple user experience (no manual pagination)
- Complete results by default
- Optional limit for performance control
