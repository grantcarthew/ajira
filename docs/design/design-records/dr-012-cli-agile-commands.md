# DR-012: CLI Agile Commands

- Date: 2026-01-12
- Status: Proposed
- Category: CLI

## Problem

Users need to work with Jira Agile features (boards, sprints, epics) from the command line. The existing ajira CLI handles issues but lacks support for agile-specific workflows like viewing sprints, adding issues to sprints, and managing epics.

Key constraints:

- Jira has two APIs: standard REST API and Agile REST API (`/rest/agile/1.0/`)
- Sprint operations require a board context
- Epic operations can work via JQL (project-based) or Agile API (board-based)
- The Agile API has limitations with next-gen/team-managed projects

## Decision

Add three command groups for agile operations:

Board commands:

- `ajira board list` - List boards in project

Sprint commands:

- `ajira sprint list` - List sprints for a board
- `ajira sprint add <sprint-id> <issue-keys...>` - Add issues to sprint

Epic commands:

- `ajira epic list` - List epics (wrapper around issue list)
- `ajira epic create` - Create epic (wrapper around issue create)
- `ajira epic add <epic-key> <issue-keys...>` - Add issues to epic
- `ajira epic remove <issue-keys...>` - Remove issues from epic

Issue list extensions:

- `ajira issue list --sprint <id>` - Filter issues by sprint
- `ajira issue list --epic <key>` - Filter issues by epic

Board context:

- Environment variable: `JIRA_BOARD`
- Flag: `--board` (no short flag)
- At least one required for sprint commands
- Flag overrides environment variable
- Not required for epic commands (use project context)

## Why

Board commands use Agile API:

- Board list requires Agile API, no JQL equivalent
- Returns board ID needed for sprint operations
- Filters by project when JIRA_PROJECT is set

Sprint commands use Agile API:

- Sprint list requires board context (API design)
- Sprint add uses dedicated endpoint for moving issues
- No JQL alternative for sprint management operations

Epic commands use JQL approach:

- Epics are issues with type "Epic", searchable via JQL
- JQL provides richer data (key, status, priority) than Agile API
- Works with both classic and next-gen projects
- Avoids requiring board context for epic operations
- epic list and epic create are thin wrappers around existing issue commands

Issue list extensions use JQL:

- Adds `parent = <key>` for epic filtering
- Adds `sprint = <id>` for sprint filtering
- Inherits all existing issue list filters

No short flags for new agile flags:

- `--board`, `--sprint`, `--epic` have no short flags
- Avoids conflicts with existing flags (`-s` is summary, `-b` is body)
- These flags are used less frequently than core flags

## Structure

### Board List

```
ajira board list [flags]
```

Output:

```
ID      NAME                    TYPE      PROJECT
1342    GCP Board               scrum     GCP
1455    Support Board           kanban    SUP
```

Flags:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| --project | -p | string | $JIRA_PROJECT | Filter by project |
| --limit | -l | int | 0 | Maximum boards (0 = all) |
| --json | -j | bool | false | JSON output |

Behaviour:

- If project set (env or flag): filter to that project
- If no project: show all accessible boards

### Sprint List

```
ajira sprint list [flags]
```

Output:

```
ID      NAME                    STATE     START         END           GOAL
42      Sprint 23               active    2026-01-06    2026-01-20    Finish auth
43      Sprint 24               future    2026-01-20    2026-02-03
```

Flags:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| --board | | int | $JIRA_BOARD | Board ID (required) |
| --state | | string | | Filter: active, future, closed |
| --current | | bool | false | Shorthand for --state active |
| --limit | -l | int | 0 | Maximum sprints (0 = all) |
| --json | -j | bool | false | JSON output |

### Sprint Add

```
ajira sprint add <sprint-id> <issue-keys...>
```

Output:

```
Added 3 issues to sprint 42
```

Constraints:

- API limits 50 issues per request (let API enforce)
- Can only add to open (future) or active sprints

### Epic List

```
ajira epic list [flags]
```

Wrapper around `issue list -t Epic`. Inherits all issue list flags.

Output: Same as issue list (key, type, status, priority, summary)

### Epic Create

```
ajira epic create -s "Epic name" [flags]
```

Wrapper around `issue create -t Epic`. Inherits all issue create flags.

### Epic Add

```
ajira epic add <epic-key> <issue-keys...>
```

Output:

```
Added 3 issues to epic GCP-50
```

Constraints:

- API limits 50 issues per request (let API enforce)
- Issues can only belong to one epic (moves from previous)
- Does not work for next-gen projects (API limitation)

### Epic Remove

```
ajira epic remove <issue-keys...>
```

Output:

```
Removed 2 issues from epic
```

No epic key argument needed - API clears epic link from specified issues.

### Issue List Extensions

New flags for issue list:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| --sprint | | string | Filter by sprint ID |
| --epic | | string | Filter by epic key |

These combine with existing filters:

```
ajira issue list --sprint 42 --status "In Progress"
ajira issue list --epic GCP-50 --assignee me
```

## Trade-offs

Accept:

- Sprint commands require board ID (API design limitation)
- Epic add/remove do not work with next-gen projects (API limitation)
- No short flags for --board, --sprint, --epic (to avoid conflicts)
- Two ways to list epics: `epic list` and `issue list -t Epic`

Gain:

- Consistent UX: epic commands use project context like issue commands
- JQL approach for epics works with all project types
- Epic list/create inherit all existing issue flags for free
- Sprint and epic issue filtering integrates with existing issue list

## Alternatives

Agile API for epic list:

- Pro: Uses dedicated epic endpoint
- Con: Requires board context (extra configuration)
- Con: Returns limited fields (no key, status, priority)
- Con: Limited support for next-gen projects
- Rejected: JQL approach provides better data and UX

Separate epic issues command:

- Pro: Explicit command for listing epic issues
- Con: Duplicates issue list functionality
- Con: Cannot combine with other issue filters
- Rejected: --epic flag on issue list is more flexible

Short flags for agile options:

- Pro: Faster typing
- Con: -b conflicts with --body on comment add
- Con: -s conflicts with --summary on create/edit
- Con: -e not obviously "epic" to users
- Rejected: Avoid confusion, these flags are less frequently used

Client-side 50 issue limit validation:

- Pro: Faster feedback
- Con: Couples client to API limits that may change
- Rejected: Let API enforce its own constraints
