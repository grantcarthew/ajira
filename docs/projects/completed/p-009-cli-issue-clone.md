# P-009: Issue Clone

- Status: Completed
- Started: 2026-01-12
- Completed: 2026-01-12

## Overview

Add the ability to clone issues. Cloning is a common workflow for creating similar issues, copying templates, or duplicating work items across projects. The clone command allows field modifications during the clone operation.

## Goals

1. Implement `issue clone` command to duplicate an issue
2. Support field modifications during clone (summary, priority, assignee, reporter, type, labels)
3. Optionally link cloned issue to original
4. Support cross-project cloning with validation

## Scope

In Scope:

- `ajira issue clone <key>` - Clone an issue
- `--summary`, `-s` - Override summary
- `--assignee`, `-a` - Override assignee
- `--reporter`, `-r` - Override reporter
- `--priority`, `-P` - Override priority
- `--type`, `-t` - Override issue type
- `--labels`, `-L` - Override labels
- `--link` - Create link to original (default "Clones", or specify type)
- `--project`, `-p` - Clone to different project (global flag)

Out of Scope:

- Text substitution (YAGNI - use edit command after clone)
- Cloning attachments (API complexity)
- Cloning comments (YAGNI)
- Cloning subtasks (consider for future)
- Bulk cloning (covered in P-013)

## Success Criteria

- [ ] `issue clone` creates a new issue with same fields as original
- [ ] Field overrides work correctly (summary, priority, assignee, reporter, type, labels)
- [ ] `--project` allows cross-project cloning with issue type validation
- [ ] `--link` creates relationship to original using validated link type
- [ ] Text output returns new issue URL
- [ ] JSON output includes original and cloned keys
- [ ] Error handling for permission issues and invalid fields
- [ ] Tests cover cloning with various field combinations

## Deliverables

- `internal/cli/issue_clone.go` - Clone command implementation
- `internal/jira/validate.go` - Add ValidateLinkType function
- DR-011: Issue Clone Command
- Unit tests for field mapping and validation
- Integration tests for clone operations

## Technical Approach

1. GET original issue with all fields
2. Validate target project issue type if cross-project clone
3. Build create request with original fields
4. Apply field overrides from flags
5. POST to create new issue
6. Optionally create link to original (validated link type)
7. Return new issue URL

## Dependencies

- P-007 (Issue Linking) - Completed, provides link creation pattern
