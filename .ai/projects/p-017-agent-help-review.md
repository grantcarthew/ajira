# p-017: Agent Help Review

- Status: Proposed
- Started:
- Completed:

## Overview

Review and optimize the `ajira help agents` documentation for token efficiency and completeness. Ensure all commands are documented with useful examples while minimizing token usage.

## Goals

1. Review all command examples for usefulness and redundancy
2. Ensure new p-014 commands are properly represented
3. Verify schemas.md matches current JSON output
4. Check for missing commands or flags
5. Optimize for token efficiency

## Review Checklist

- [ ] Review `internal/cli/help/agents.md` for redundant examples
- [ ] Verify all commands have at least one example
- [ ] Check flag coverage (are important flags shown?)
- [ ] Review `internal/cli/help/schemas.md` for accuracy
- [ ] Test JSON output matches documented schemas
- [ ] Consider grouping/ordering of commands
- [ ] Check markdown.md is still accurate
- [ ] Run token count comparison before/after

## Files to Review

- `internal/cli/help/agents.md` - Main agent reference
- `internal/cli/help/schemas.md` - JSON output schemas
- `internal/cli/help/markdown.md` - Markdown syntax guide

## Notes

Recent changes from p-014 added:
- `ajira issue watch/unwatch`
- `ajira open`
- `ajira release list`
- `ajira user search`
- `ajira field list`

These were consolidated into the Commands section and schemas.md was updated.
