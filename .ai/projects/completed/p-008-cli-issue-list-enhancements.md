# p-008: Issue List Enhancements

- Status: Completed
- Started: 2026-01-09
- Completed: 2026-01-09

## Overview

Enhance the `issue list` command with additional filter flags and sorting options. The current implementation has basic JQL support but lacks convenience filters for common query patterns.

## Goals

1. Add convenience filter flags for common query patterns
2. Implement sorting and ordering options
3. Maintain backwards compatibility with existing flags

## Scope

In Scope:

New filter flags:

- `--reporter`, `-r` - Filter by reporter
- `--priority`, `-P` - Filter by priority
- `--labels`, `-L` - Filter by labels (comma-separated)
- `--watching`, `-w` - Issues current user is watching

Sorting options:

- `--order-by` - Sort field (created, updated, priority, key, rank)
- `--reverse` - Reverse sort order (ASC instead of DESC)

Out of Scope:

- Date filters (--created, --updated) - agents can use JQL
- Output format options (--plain, --csv, --columns, --no-headers) - agents use --json
- Interactive TUI mode (ajira is non-interactive)

## Success Criteria

- [x] `--reporter` filter correctly modifies JQL query
- [x] `--priority` filter correctly modifies JQL query
- [x] `--labels` filter correctly modifies JQL query (supports multiple labels)
- [x] `--watching` filter correctly modifies JQL query
- [x] `--order-by` controls sort field
- [x] `--reverse` changes sort direction to ASC
- [x] Filters combine correctly with existing filters (AND logic)
- [x] Tests cover all new flags and combinations

## Deliverables

- [x] Updated `internal/cli/issue_list.go` - Enhanced filtering with validation
- [x] Updated `internal/cli/issue_test.go` - Unit tests for JQL building
- [x] Created `docs/flags-and-arguments.md` - Document new flags
- [x] dr-010: Issue List Filtering and Sorting

## Dependencies

None - enhances existing command.
