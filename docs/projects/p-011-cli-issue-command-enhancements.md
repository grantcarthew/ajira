# P-011: Issue Command Enhancements

- Status: Proposed
- Started:
- Completed:

## Overview

Enhance existing issue commands (create, edit, move, delete) with additional field support and options. This includes parent/epic assignment, components, fix versions, custom fields, and transition enhancements.

## Goals

1. Add parent/epic field support to create and edit
2. Add component and fix version support
3. Add custom field support
4. Enhance move command with comment and resolution
5. Add cascade delete for subtasks
6. Support label removal syntax

## Scope

In Scope:

Create enhancements (`issue create`):

- `--parent`, `-P` - Attach to epic/parent issue
- `--component`, `-C` - Set component(s)
- `--fix-version` - Set fix version(s)
- `--custom "field:value"` - Set custom field

Edit enhancements (`issue edit`):

- `--parent`, `-P` - Change parent/epic
- `--component`, `-C` - Set/add/remove components
- `--fix-version` - Set/add/remove fix versions
- `--custom "field:value"` - Set custom field
- Minus notation for removal (`--label -old-label`)

Move enhancements (`issue move`):

- `--comment`, `-m` - Add comment during transition
- `--resolution`, `-R` - Set resolution (for done transitions)
- `--assignee`, `-a` - Set assignee during transition

Delete enhancements (`issue delete`):

- `--cascade` - Delete issue with all subtasks

Out of Scope:

- Attachment management (separate project)
- Bulk operations (covered in P-013)

## Success Criteria

- [ ] `issue create --parent` attaches issue to epic
- [ ] `issue create --component` sets components
- [ ] `issue create --fix-version` sets fix version
- [ ] `issue create --custom` sets custom fields
- [ ] `issue edit` supports all new fields
- [ ] `issue edit --label -name` removes label
- [ ] `issue move --comment` adds comment during transition
- [ ] `issue move --resolution` sets resolution
- [ ] `issue delete --cascade` deletes subtasks
- [ ] Field validation provides clear error messages
- [ ] Tests cover all new options

## Deliverables

- Updated `internal/cli/issue_create.go` - New field support
- Updated `internal/cli/issue_edit.go` - New field support, minus notation
- Updated `internal/cli/issue_move.go` - Comment and resolution
- Updated `internal/cli/issue_delete.go` - Cascade option
- DR-012: Field Management Patterns (custom fields, minus notation)
- Unit tests for field handling
- Integration tests for enhanced commands

## Research Areas

- Jira API for setting parent/epic (differs between project types)
- Custom field API and field ID discovery
- Transition screens and required fields
- Subtask deletion API

## Questions and Uncertainties

- How to discover custom field IDs programmatically?
- Which transitions allow comments/resolution?
- How does parent field work in classic vs next-gen projects?
- Should we validate components/versions exist before setting?

## Dependencies

- P-010 (Agile Features) - May share epic-related code
