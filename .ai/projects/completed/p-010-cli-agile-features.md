# p-010: Agile Features

- Status: Completed
- Started: 2026-01-13
- Completed: 2026-01-13

## Overview

Add support for Jira Agile features: epics, sprints, and boards. These are core to agile workflows and frequently needed for automation. This project adds commands for listing and managing epics and sprints.

## Goals

1. Implement epic commands (list, create, add issues, remove issues)
2. Implement sprint commands (list, add issues)
3. Implement board list command
4. Support filtering issues within epics and sprints
5. Integrate with existing issue list filters

## Scope

In Scope:

Board commands:

- `ajira board list` - List boards (requires project context)

Sprint commands (require JIRA_BOARD or --board):

- `ajira sprint list` - List sprints with state/current filters
- `ajira sprint add <sprint-id> <issue-keys...>` - Add issues to sprint

Epic commands (use project context via JQL):

- `ajira epic list` - List epics (wrapper around issue list -t Epic)
- `ajira epic create` - Create epic (wrapper around issue create -t Epic)
- `ajira epic add <epic-key> <issue-keys...>` - Add issues to epic
- `ajira epic remove <issue-keys...>` - Remove issues from epic

Issue list extensions:

- `ajira issue list --sprint <id>` - Filter issues by sprint
- `ajira issue list --epic <key>` - Filter issues by epic

Out of Scope:

- Sprint creation/management (typically done in UI)
- Board creation (admin function)
- Backlog management
- Velocity/burndown charts

## Success Criteria

- [x] `board list` shows boards with id, name, type, project
- [x] `sprint list` shows sprints with id, name, state, dates, goal
- [x] `sprint list --state` filters by active/future/closed
- [x] `sprint list --current` shows active sprints
- [x] `sprint add` adds issues to sprint
- [x] `epic list` displays epics with key, status, priority, summary
- [x] `epic create` creates epic with summary and optional fields
- [x] `epic add` adds multiple issues to an epic
- [x] `epic remove` removes issues from their epic
- [x] `issue list --sprint` filters issues by sprint ID
- [x] `issue list --epic` filters issues by epic key
- [x] All commands support `--json` output
- [x] Tests cover board, sprint, and epic operations

## Deliverables

- `internal/cli/epic.go` - Epic command group
- `internal/cli/epic_list.go` - Epic list implementation
- `internal/cli/epic_create.go` - Epic create implementation
- `internal/cli/epic_add.go` - Add issues to epic
- `internal/cli/epic_remove.go` - Remove issues from epic
- `internal/cli/sprint.go` - Sprint command group
- `internal/cli/sprint_list.go` - Sprint list implementation
- `internal/cli/sprint_add.go` - Add issues to sprint
- `internal/cli/board.go` - Board list command
- dr-012: CLI Agile Commands
- Integration tests for agile operations

## Research Completed

Jira Agile REST API:

- Uses `/rest/agile/1.0/` base path (separate from standard API)
- Sprint operations require board context
- Epic Agile API has limitations with next-gen projects
- Using JQL for epic operations provides better compatibility

Design decisions:

- Sprint commands require JIRA_BOARD or --board flag
- Epic commands use JQL via project context (no board required)
- Epic add/remove use Agile API (limited next-gen support)
- Issue list extended with --sprint and --epic flags

## Dependencies

- p-008 (Issue List Enhancements) - for consistent filtering in epic/sprint issue lists
