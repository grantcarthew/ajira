# P-007: Issue Linking

- Status: Proposed
- Started:
- Completed:

## Overview

Add issue linking capabilities to ajira. Issue linking is essential for tracking dependencies, duplicates, and relationships between issues. This includes linking issues to each other and adding remote web links.

## Goals

1. Implement `issue link` command to link two issues together
2. Implement `issue unlink` command to remove links between issues
3. Implement `issue link remote` command to add web links to issues
4. Support all standard Jira link types (Blocks, Duplicates, Relates to, etc.)
5. Display linked issues in `issue view` output

## Scope

In Scope:

- `ajira issue link <key1> <key2> <link-type>` - Create link between issues
- `ajira issue unlink <key1> <key2>` - Remove link between issues
- `ajira issue link remote <key> <url> [title]` - Add remote web link
- `ajira issue link list` - List available link types
- Update `issue view` to show linked issues
- JSON output support for all commands

Out of Scope:

- Bulk linking operations (covered in P-013)
- Interactive link type selection (non-interactive CLI)

## Success Criteria

- [ ] `issue link` creates links between issues with specified type
- [ ] `issue unlink` removes links between issues
- [ ] `issue link remote` adds web links to issues
- [ ] `issue link list` displays available link types
- [ ] `issue view` shows linked issues section
- [ ] All commands support `--json` output
- [ ] Error messages clearly indicate invalid link types or missing issues
- [ ] Tests cover link creation, removal, and display

## Deliverables

- `internal/cli/issue_link.go` - Link command implementation
- `internal/cli/issue_unlink.go` - Unlink command implementation
- `internal/cli/issue_link_remote.go` - Remote link implementation
- Updated `internal/cli/issue_view.go` - Show linked issues
- DR-009: Issue Linking Design (if significant decisions needed)
- Integration tests for linking functionality

## Research Areas

- Jira REST API endpoints for issue links
- Available link types and their directionality (inward/outward)
- Remote link API structure
- How to display link direction in CLI output

## Questions and Uncertainties

- How to handle bidirectional vs unidirectional link types?
- Should we validate that both issues exist before attempting to link?
- How to display link direction clearly in text output?

## Dependencies

None - builds on existing issue infrastructure.
