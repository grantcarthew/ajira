# P-012: Time Tracking

- Status: Deferred
- Started:
- Completed:
- Deferred: 2026-01-13
- Reason: Time tracking not actively used in target Jira instance

## Overview

Add worklog (time tracking) functionality to ajira. Time tracking is essential for teams that log work hours against issues. This project adds commands to add and list worklogs.

## Goals

1. Implement worklog add command
2. Implement worklog list command
3. Support flexible time format parsing
4. Include worklog data in issue view
5. Support worklog comments

## Scope

In Scope:

Worklog commands:

- `ajira issue worklog add <key> <time>` - Add time entry
- `ajira issue worklog list <key>` - List worklogs on issue
- `--comment`, `-m` - Worklog description
- `--started` - When work was performed (defaults to now)

Time format support:

- Duration strings: "2h", "30m", "1d", "2h 30m"
- Jira format: "2d 4h 30m"

Issue view integration:

- Show total time logged
- Option to show worklog entries

Out of Scope:

- Worklog editing (rarely needed via CLI)
- Worklog deletion (rarely needed via CLI)
- Timesheet views (aggregation across issues)
- Tempo or other time tracking plugin integration

## Success Criteria

- [ ] `issue worklog add` creates worklog entry
- [ ] Time parsing handles common formats (h, m, d combinations)
- [ ] `--comment` adds description to worklog
- [ ] `--started` sets work date/time
- [ ] `issue worklog list` shows worklogs with author, time, date
- [ ] `issue view` shows time tracking summary
- [ ] JSON output includes worklog details
- [ ] Error handling for invalid time formats
- [ ] Tests cover time parsing and worklog operations

## Deliverables

- `internal/cli/issue_worklog.go` - Worklog command group
- `internal/cli/issue_worklog_add.go` - Add worklog implementation
- `internal/cli/issue_worklog_list.go` - List worklogs implementation
- `internal/jira/time.go` - Time duration parsing
- Updated `internal/cli/issue_view.go` - Show time tracking
- DR-013: Worklog Command Design
- Unit tests for time parsing
- Integration tests for worklog operations

## Technical Approach

Time parsing strategy:

1. Parse duration string into seconds
2. Support: d (days), h (hours), m (minutes)
3. Handle combinations: "1d 2h 30m"
4. Convert to Jira's expected format

## Research Areas

- Jira worklog API structure
- Time tracking permission requirements
- Original estimate vs remaining estimate fields
- Time format in API (seconds vs formatted string)

## Questions and Uncertainties

- Does ajira's target Jira instance have time tracking enabled?
- How to handle timezone for started time?
- Should we show remaining estimate in issue view?
- What's the maximum worklog duration allowed?

## Dependencies

None - new functionality.
