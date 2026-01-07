# P-010: Agile Features

- Status: Proposed
- Started:
- Completed:

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

Epic commands:
- `ajira epic list` - List epics in project
- `ajira epic list <key>` - List issues in an epic
- `ajira epic create` - Create a new epic
- `ajira epic add <epic-key> <issue-keys...>` - Add issues to epic
- `ajira epic remove <issue-keys...>` - Remove issues from epic

Sprint commands:
- `ajira sprint list` - List sprints
- `ajira sprint list <id>` - List issues in sprint
- `ajira sprint list --current` - Current active sprint
- `ajira sprint list --state <state>` - Filter by state (active, future, closed)
- `ajira sprint add <sprint-id> <issue-keys...>` - Add issues to sprint

Board commands:
- `ajira board list` - List boards in project

Out of Scope:

- Sprint creation/management (typically done in UI)
- Board creation (admin function)
- Backlog management
- Velocity/burndown charts

## Success Criteria

- [ ] `epic list` displays epics with key, name, status
- [ ] `epic list <key>` shows issues in epic with standard filters
- [ ] `epic create` creates epic with name and optional fields
- [ ] `epic add` adds multiple issues to an epic
- [ ] `epic remove` removes issues from their epic
- [ ] `sprint list` shows sprints with id, name, state, dates
- [ ] `sprint list <id>` shows issues in sprint
- [ ] `sprint list --current` shows current sprint issues
- [ ] `sprint add` adds issues to sprint
- [ ] `board list` shows boards with id and name
- [ ] All commands support `--json` output
- [ ] Tests cover epic and sprint operations

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
- DR-011: Agile Command Structure
- Integration tests for agile operations

## Research Areas

- Jira Agile REST API (different from standard Jira API)
- Epic field handling (epic name vs epic link in classic vs next-gen)
- Sprint API endpoints and authentication
- Board types (scrum vs kanban) and their differences

## Questions and Uncertainties

- How do epic fields differ between classic and next-gen projects?
- Is the Agile API available on all Jira instances?
- How to handle projects without boards?
- Should sprint commands require board ID or infer from project?

## Dependencies

- P-008 (Issue List Enhancements) - for consistent filtering in epic/sprint issue lists
