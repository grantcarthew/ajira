# P-008: Issue List Enhancements

- Status: Proposed
- Started:
- Completed:

## Overview

Enhance the `issue list` command with additional filters, sorting options, and output modes. The current implementation has basic JQL support but lacks convenience filters and scripting-friendly output formats that are essential for automation.

## Goals

1. Add convenience filter flags for common query patterns
2. Implement sorting and ordering options
3. Add plain/CSV output modes for scripting
4. Add column selection and header control
5. Maintain backwards compatibility with existing flags

## Scope

In Scope:

New filter flags:

- `--reporter`, `-r` - Filter by reporter
- `--priority`, `-y` - Filter by priority
- `--labels`, `-l` - Filter by labels (repeatable)
- `--created` - Filter by creation date (e.g., `-7d`, `month`, `week`)
- `--updated` - Filter by update date
- `--watching`, `-w` - Issues current user is watching

Sorting options:

- `--order-by` - Sort field (created, updated, priority, key, rank)
- `--reverse` - Reverse sort order

Output options:

- `--plain` - Disable colours, simple format
- `--csv` - CSV output format
- `--columns` - Select output columns
- `--no-headers` - Omit header row

Out of Scope:

- Interactive TUI mode (ajira is non-interactive)
- Raw JSON mode (already have `--json`)
- History/recently viewed (requires additional API)

## Success Criteria

- [ ] All new filter flags correctly modify JQL query
- [ ] `--order-by` and `--reverse` control sort order
- [ ] `--plain` outputs without ANSI colours
- [ ] `--csv` produces valid CSV output
- [ ] `--columns` allows column selection
- [ ] `--no-headers` omits header row
- [ ] Filters combine correctly (AND logic)
- [ ] Date filters support relative formats (-7d, month, week)
- [ ] Tests cover all new flags and combinations

## Deliverables

- Updated `internal/cli/issue_list.go` - Enhanced filtering and output
- DR-009: Output Formats and Filtering Strategy
- Unit tests for JQL building
- Integration tests for output formats

## Research Areas

- JQL syntax for date functions (startOfDay, startOfWeek, etc.)
- JQL syntax for watcher queries
- CSV escaping requirements
- Column width handling for plain mode

## Questions and Uncertainties

- Should `--columns` use field names or display names?
- How to handle columns that don't exist for some issue types?
- Should date filters be absolute or relative only?
- What columns should be available for selection?

## Dependencies

None - enhances existing command.
