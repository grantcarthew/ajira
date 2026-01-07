# P-015: CLI Help System

- Status: Completed
- Started: 2026-01-07
- Completed: 2026-01-07

## Overview

Improve ajira's help system to be comprehensive and AI-agent friendly. AI agents need token-efficient documentation with examples to use ajira effectively. This project adds custom help topics, improves per-command help with examples, and documents JSON output schemas.

## Goals

1. Create `ajira help agents` command with comprehensive AI-friendly reference
2. Create `ajira help schemas` command documenting JSON output structures
3. Add examples to all command help text
4. Establish help text template for consistency across commands
5. Use Go embed for maintainable help content

## Scope

In Scope:

Custom help topics:

- `ajira help agents` - Token-efficient AI agent reference (env vars, commands, examples, workflows)
- `ajira help schemas` - JSON output structures for all commands

Per-command improvements:

- Add 2-3 examples per command
- Consistent help template across all commands
- Clear output descriptions

Technical:

- Use `//go:embed` for help topic content
- Help content in `internal/cli/help/*.md`
- Markdown rendering for terminal output

Out of Scope:

- Interactive help/tutorials
- Man page generation
- External documentation website

## Success Criteria

- [x] `ajira help agents` outputs comprehensive AI reference
- [x] `ajira help schemas` documents all JSON output structures
- [x] Every command has at least 2 examples in --help
- [x] Help text follows consistent template
- [x] Embedded Markdown files render correctly in terminal
- [x] `ajira help agents` is under 2000 tokens (~400 tokens)
- [x] Tests verify help commands work

## Deliverables

- `internal/cli/help/` directory with embedded Markdown files
- `internal/cli/help.go` - Help command implementations
- `internal/cli/help/agents.md` - AI agent reference
- `internal/cli/help/schemas.md` - JSON schema documentation
- Updated help text for all existing commands
- DR-016: Help System Design (if significant decisions needed)

## Technical Approach

File structure:

```
internal/cli/
├── help/
│   ├── agents.md
│   └── schemas.md
├── help.go          # Embeds files, registers commands
└── ...
```

Help template for commands:

```
Short: One-line description

Long:
  What it does
  Key behaviors

Examples:
  ajira cmd arg              # Basic usage
  ajira cmd arg --flag       # With flag
  ajira cmd arg --json       # JSON output
```

## Questions and Uncertainties

Resolved:

- Token count for `help agents`: ~400 tokens achieved (under 2000 target)
- `help schemas` format: Field lists only (most token-efficient)
- Additional help topics: Only agents and schemas needed; workflow examples included in agents.md

## Dependencies

None - improves existing CLI infrastructure.

## Notes

Discussion points from design session:

- Use Go 1.16+ embed for help content in Markdown files
- Keep short command help inline, longer topics in embedded files
- Render Markdown with glamour for terminal styling
- Goal: agent runs `ajira help agents` and knows almost everything needed

Implementation decisions (2026-01-07):

- Custom help command replaces Cobra default to support subcommands
- Help topics accessible via `ajira help agents` and `ajira help schemas`
- agents.md focuses on examples, minimal prose, text output preferred over JSON
- schemas.md uses simple field lists for token efficiency
- Root help includes reference to `ajira help agents` for agent discoverability
- No DR required; decisions documented here
