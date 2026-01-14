# p-007: Issue Linking

- Status: Complete
- Started: 2026-01-08
- Completed: 2026-01-08

## Overview

Add issue linking capabilities to ajira. Issue linking is essential for tracking dependencies, duplicates, and relationships between issues. This includes linking issues to each other and adding remote web links.

## Goals

1. Implement `issue link add` command to link two issues together
2. Implement `issue link remove` command to remove links between issues
3. Implement `issue link url` command to add web links to issues
4. Implement `issue link types` command to list available link types
5. Display linked issues in `issue view` output
6. Add singular/plural aliases to metadata commands

## Scope

In Scope:

- `ajira issue link add <key1> <type> <key2>` - Create link between issues (reads as sentence)
- `ajira issue link remove <key1> <key2>` - Remove all links between issues
- `ajira issue link url <key> <url> [title]` - Add remote web link (alias: web)
- `ajira issue link types` - List available link types (alias: type)
- Update `issue view` to show linked issues with status and summary
- JSON output support for all commands
- Aliases: priority/priorities, type/types, status/statuses

Out of Scope:

- Bulk linking operations (covered in p-013)
- Interactive link type selection (non-interactive CLI)

## Success Criteria

- [x] `issue link add` creates links between issues with specified type
- [x] `issue link remove` removes all links between issues
- [x] `issue link url` adds web links to issues
- [x] `issue link types` displays available link types
- [x] `issue view` shows linked issues section
- [x] All commands support `--json` output
- [x] Error messages clearly indicate invalid link types (pre-validation)
- [x] Tests cover link creation, removal, and display
- [x] Added aliases to existing metadata commands

## Deliverables

- [x] `internal/cli/issue_link.go` - Parent link command
- [x] `internal/cli/issue_link_add.go` - Add link command
- [x] `internal/cli/issue_link_remove.go` - Remove link command
- [x] `internal/cli/issue_link_url.go` - Remote link command
- [x] `internal/cli/issue_link_types.go` - List link types command
- [x] Updated `internal/cli/issue_view.go` - Show linked issues
- [x] Updated `internal/jira/metadata.go` - GetLinkTypes function
- [x] dr-009: Issue Linking Design
- [x] Tests in `internal/cli/issue_link_test.go`
- [x] Aliases added to priority, type, status commands

## Design Decisions

See dr-009 for details:

- Command syntax `KEY1 TYPE KEY2` reads as sentence (e.g., "GCp-123 Blocks GCp-456")
- `remove` deletes all links between two issues (no type required)
- Pre-validation of link types with helpful error messages
- Links displayed in issue view as: `direction KEY (Status) - Summary`

## Notes

- Pre-existing test failure in converter package (TestMarkdownToADF_NestedBlockquotes) - not related to this work
- Fixed during end-to-end testing: Jira API `outwardIssue`/`inwardIssue` naming is counterintuitive - had to swap them to match command syntax
- Comprehensive end-to-end testing completed with real Jira issues
