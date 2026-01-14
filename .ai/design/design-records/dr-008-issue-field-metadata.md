# dr-008: Issue Field Metadata Commands

- Date: 2026-01-06
- Status: Accepted
- Category: CLI

## Problem

Users need to discover valid values for issue fields when creating or editing issues:

- Priority names vary by Jira instance (e.g., "Major" not "High")
- Issue types are project-specific (Task, Bug, Story, Epic, etc.)
- Statuses are project and workflow-specific

Without this information, users must guess field values or consult the Jira web UI, breaking the CLI workflow.

## Decision

Add three commands under `ajira issue` for querying field metadata:

```
ajira issue priority    # List available priorities
ajira issue type        # List issue types for project
ajira issue status      # List statuses for project
```

These are simple query commands without subcommands since they are read-only lookups.

## Command Specifications

### ajira issue priority

Lists all priorities available in the Jira instance.

API: `GET /rest/api/3/priority`

Scope: Instance-wide (priorities are typically global)

Output columns: NAME, DESCRIPTION

### ajira issue type

Lists issue types available for the current project.

API: `GET /rest/api/3/project/{projectKey}/statuses` or `GET /rest/api/3/issuetype/project?projectId=...`

Scope: Project-specific (requires `-p` flag or `JIRA_PROJECT`)

Output columns: NAME, DESCRIPTION

### ajira issue status

Lists statuses available for the current project.

API: `GET /rest/api/3/project/{projectKey}/statuses`

Scope: Project-specific (requires `-p` flag or `JIRA_PROJECT`)

Output columns: NAME, CATEGORY

## Output Formats

Plain text (default):

```
NAME      DESCRIPTION
Critical  100+ staff affected
Major     50+ users affected
Minor     Single user impact
```

JSON (`-j` flag):

```json
[
  {"name": "Critical", "description": "100+ staff affected"},
  {"name": "Major", "description": "50+ users affected"}
]
```

## Why

- Enables self-service discovery of valid field values
- Keeps users in the CLI workflow without switching to web UI
- Simple noun commands (no `list` subcommand) since these are pure lookups with no CRUD operations
- Follows existing pattern of `ajira issue` subcommands
- Project-scoped commands respect `-p` flag for consistency

## Trade-offs

Accept:

- Three new commands increase `ajira issue --help` output
- Status list may be long for complex workflows

Gain:

- Users can discover valid values without leaving CLI
- Reduces errors from invalid field values
- Consistent with existing CLI patterns

## Alternatives

Subcommand pattern (`ajira issue priority list`):

- Pro: Allows future subcommands
- Con: No realistic CRUD operations for these read-only resources
- Con: Extra word for simple lookup
- Rejected: Simpler is better for pure query commands

Top-level commands (`ajira priority`, `ajira type`):

- Pro: Shorter commands
- Con: Pollutes top-level namespace
- Con: These are issue-related, not standalone concepts
- Rejected: Belongs under `issue` for logical grouping
