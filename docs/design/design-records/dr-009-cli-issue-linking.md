# DR-009: Issue Linking Commands

- Date: 2026-01-08
- Status: Accepted
- Category: CLI

## Problem

Users need to manage relationships between Jira issues and add external web links without leaving the CLI:

- Link issues to show dependencies, duplicates, or related work
- Remove links when relationships change
- Add web URLs to reference external resources (GitHub PRs, docs, etc.)
- Discover available link types for the Jira instance

Without these capabilities, users must switch to the Jira web UI to manage issue relationships.

## Decision

Add a `link` subcommand under `ajira issue` with four operations:

```
ajira issue link add KEY1 TYPE KEY2     # Create link (KEY1 blocks KEY2)
ajira issue link remove KEY1 KEY2       # Remove all links between issues
ajira issue link url KEY URL [TITLE]    # Add web URL to issue
ajira issue link types                  # List available link types
```

Additionally:

- Display linked issues in `ajira issue view` output
- Add singular/plural aliases to existing metadata commands

## Command Specifications

### ajira issue link add KEY1 TYPE KEY2

Creates a directional link between two issues.

API: `POST /rest/api/3/issueLink`

Request body:

```json
{
  "outwardIssue": { "key": "KEY1" },
  "inwardIssue": { "key": "KEY2" },
  "type": { "name": "TYPE" }
}
```

Argument order reads as a sentence: "KEY1 blocks KEY2"

Validation: Pre-fetch link types, validate TYPE before API call. Error format:

```
error: link type not found: elephant (available: Blocks, Cloners, Duplicate, Relates)
```

### ajira issue link remove KEY1 KEY2

Removes all links between two issues regardless of type.

API: `DELETE /rest/api/3/issueLink/{linkId}`

Process:

1. Fetch issue KEY1 with `issuelinks` field
2. Find all links where the other issue is KEY2
3. Delete each link by ID

### ajira issue link url KEY URL [TITLE]

Adds a web URL to an issue. Alias: `web`

API: `POST /rest/api/3/issue/{key}/remotelink`

Request body:

```json
{
  "object": {
    "url": "URL",
    "title": "TITLE"
  }
}
```

If TITLE is omitted, the URL is used as the title.

### ajira issue link types

Lists available link types. Alias: `type`

API: `GET /rest/api/3/issueLinkType`

Output columns: NAME, OUTWARD, INWARD

```
NAME            OUTWARD              INWARD
Blocks          blocks               is blocked by
Cloners         clones               is cloned by
Duplicate       duplicates           is duplicated by
```

### Updated issue view

Add linked issues section to `ajira issue view` output:

```
Links:
  blocks GCP-456 (In Progress) - Fix the login page
  is blocked by GCP-789 (Done) - Database migration
```

Data comes from the existing issue fetch by adding `issuelinks` to requested fields. No additional API calls required.

## Command Aliases

New aliases for existing commands:

- `ajira issue priority` -> alias `priorities`
- `ajira issue type` -> alias `types`
- `ajira issue status` -> alias `statuses`
- `ajira issue link types` -> alias `type`
- `ajira issue link url` -> alias `web`

## Why

- Argument order `KEY1 TYPE KEY2` reads naturally as "KEY1 blocks KEY2", making directionality intuitive
- Pre-validation of link types provides clear error messages before API calls
- Removing all links between issues simplifies the command (no need to specify type)
- Including linked issues in view output requires no extra API calls
- Aliases support both singular and plural forms users might try

## Trade-offs

Accept:

- `link` command has nested subcommands (slightly deeper hierarchy)
- `remove` deletes all link types between issues (cannot selectively remove one type)
- Pre-validation requires an extra API call to fetch link types

Gain:

- Intuitive sentence-like syntax for creating links
- Clear error messages with valid options listed
- Simple remove command without needing to know link types
- Consistent with existing CLI patterns

## Alternatives

Argument order `KEY1 KEY2 TYPE`:

- Pro: Type at end like other commands
- Con: Unclear which issue is the "source" of the relationship
- Rejected: Sentence-like order is more intuitive for directional links

Require TYPE for remove:

- Pro: More precise control
- Con: Extra complexity for common case
- Con: Requires extra API call to find link type if unknown
- Rejected: Removing all links is simpler and covers most use cases

Separate `unlink` command instead of `remove`:

- Pro: Matches project document naming
- Con: Inconsistent with `add` (would expect `link`/`unlink` not `add`/`unlink`)
- Rejected: `add`/`remove` pair is more consistent
