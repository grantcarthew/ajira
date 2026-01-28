# p-017: Comprehensive Help Review

- Status: Complete
- Started: 2026-01-16
- Completed: 2026-01-16

## Overview

Complete audit of all help content in ajira - both help topics (`ajira help <topic>`) and Cobra command help (`--help`). Verify code-to-help mapping, restructure for appropriate audiences, and review UX.

## Goals

1. Verify all commands/flags in code have accurate help
2. Verify all documented features exist in code
3. Restructure agents.md for token efficiency (AI agents)
4. Compress schemas.md format
5. Keep markdown.md as full human reference
6. Review all --help for accuracy, consistency, and UX

## Results

### Token Comparison

| File | Before | After |
|------|--------|-------|
| agents.md | 87 lines, 2972 bytes | 74 lines, 2640 bytes |
| schemas.md | 69 lines, 1013 bytes | 35 lines, 1387 bytes |
| markdown.md | 210 lines, 3735 bytes | 210 lines, 3735 bytes (unchanged) |
| **Total** | 366 lines, 7720 bytes | 319 lines, 7762 bytes |

**Notes:**
- agents.md reduced by 13 lines and 332 bytes through restructuring
- schemas.md increased by 374 bytes due to adding all missing schemas (epic/sprint/board commands)
- markdown.md unchanged as expected (human reference)
- Overall line count reduced by 47 lines (13% reduction)
- Structure is now more token-efficient with clear Core/Other Commands split

### Changes Made

1. **agents.md** - Restructured with:
   - Condensed key behaviours section
   - Inline markdown conversion table (replacing verbose warnings)
   - Core Commands section with examples
   - Other Commands one-liner referencing --help for less common commands
   - Consolidated chaining section

2. **schemas.md** - Updated with:
   - Compressed `command: fields` format
   - Added all missing schemas: issue clone, issue edit, issue comment edit, epic list/create/add/remove, sprint list/add, board list

3. **markdown.md** - Verified accurate, no changes needed

4. **Cobra --help** - All commands reviewed:
   - All Short/Long/Example text is accurate and consistent
   - All flags are documented
   - No stale or non-existent features found

5. **~/context/environment.md** - Updated ajira section to match new agents.md structure

## Work Items

### 1. agents.md Restructure
- [x] Create Core Commands section with examples
- [x] Create Other Commands one-liner
- [x] Add inline markdown conversion table
- [x] Remove redundant wiki markup warnings
- [x] Verify all documented commands exist in code

### 2. schemas.md Update
- [x] Compress to `command: fields` format
- [x] Add missing schemas (clone, edit, epic/*, sprint/*, board)
- [x] Verify all schemas match actual --json output

### 3. markdown.md Verification
- [x] Confirm all syntax examples are accurate
- [x] No changes expected

### 4. Cobra --help Audit
- [x] Review all command Short text for clarity/consistency
- [x] Review all command Long text for accuracy
- [x] Review all command Example text
- [x] Verify all flags are documented
- [x] Check for stale/non-existent features
- [x] UX review - is help useful and well-organised?

### 5. Code-to-Help Verification
- [x] Every command in code has help
- [x] Every flag in code is documented
- [x] No help references non-existent features

### 6. Post-Review
- [x] Update ~/context/environment.md ajira section
- [x] Token count comparison before/after

## Files Modified

- `internal/cli/help/agents.md` - Restructured
- `internal/cli/help/schemas.md` - Compressed and completed
- `~/context/environment.md` - Updated ajira section
