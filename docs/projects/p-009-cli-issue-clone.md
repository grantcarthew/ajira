# P-009: Issue Clone

- Status: Proposed
- Started:
- Completed:

## Overview

Add the ability to clone issues. Cloning is a common workflow for creating similar issues, copying templates, or duplicating work items across projects. The clone command should allow field modifications during the clone operation.

## Goals

1. Implement `issue clone` command to duplicate an issue
2. Support field modifications during clone (summary, priority, assignee, labels)
3. Support text replacement in summary and description
4. Optionally link cloned issue to original
5. Support cross-project cloning

## Scope

In Scope:

- `ajira issue clone <key>` - Clone an issue
- `--summary`, `-s` - Override summary
- `--priority` - Override priority
- `--assignee`, `-a` - Override assignee
- `--labels` - Override labels
- `--project`, `-p` - Clone to different project
- `--replace "find:replace"` - Text substitution in summary/description
- `--link` - Create "cloned from" link to original

Out of Scope:

- Cloning attachments (API complexity)
- Cloning comments (rarely needed)
- Cloning subtasks (consider for future)
- Bulk cloning (covered in P-013)

## Success Criteria

- [ ] `issue clone` creates a new issue with same fields as original
- [ ] Field overrides work correctly (summary, priority, assignee, labels)
- [ ] `--replace` performs text substitution
- [ ] `--project` allows cross-project cloning
- [ ] `--link` creates relationship to original
- [ ] Cloned issue key returned in output
- [ ] JSON output includes original and cloned keys
- [ ] Error handling for permission issues and invalid fields
- [ ] Tests cover cloning with various field combinations

## Deliverables

- `internal/cli/issue_clone.go` - Clone command implementation
- DR-010: Issue Clone Field Handling (if needed)
- Unit tests for field mapping
- Integration tests for clone operations

## Technical Approach

1. GET original issue with all fields
2. Build create request with original fields
3. Apply field overrides from flags
4. Apply text replacements if specified
5. POST to create new issue
6. Optionally create link to original
7. Return new issue key

## Research Areas

- Which fields are cloneable vs system-generated?
- How to handle custom fields during clone?
- Cross-project field compatibility (different issue types, workflows)
- Link type for "cloned from" relationship

## Questions and Uncertainties

- Should we clone the description by default or require explicit flag?
- How to handle fields that don't exist in target project?
- Should we validate target project issue types before cloning?
- What happens if original issue type doesn't exist in target project?

## Dependencies

- P-007 (Issue Linking) - for `--link` functionality (optional, can implement without)
