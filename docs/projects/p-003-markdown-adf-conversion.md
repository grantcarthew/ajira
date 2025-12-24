# P-003: Markdown/ADF Conversion

- Status: In Progress
- Started: 2025-12-24
- Completed:

## Overview

Implement bidirectional conversion between Markdown and Atlassian Document Format (ADF). Users interact with ajira using Markdown exclusively. ADF conversion is invisible at the API boundary.

## Goals

1. Convert Markdown to ADF for issue/comment creation and editing
2. Convert ADF to Markdown for displaying issue descriptions and comments
3. Support all standard Markdown elements with 1:1 ADF mappings
4. Handle unsupported ADF elements gracefully (skip, don't error)

## Scope

In Scope:

- Markdown to ADF conversion (input)
- ADF to Markdown conversion (output)
- Paragraphs and line breaks
- Headings (h1-h6)
- Text formatting (bold, italic, strikethrough, inline code)
- Lists (ordered, unordered, task lists with checkboxes)
- Code blocks with language hints
- Links
- Tables
- Blockquotes
- Horizontal rules

Out of Scope:

- Images and attachments (requires separate upload API)
- Mentions (@user)
- Emojis
- Jira-specific elements (issue links, panels, expands)
- Nested tables

## Success Criteria

- [ ] Markdown to ADF converts all supported elements correctly
- [ ] ADF to Markdown produces readable, valid Markdown
- [ ] Round-trip conversion preserves content semantics
- [ ] Unsupported ADF elements are skipped gracefully
- [ ] Unit tests cover all supported element types
- [ ] Integration test with real Jira API validates ADF acceptance

## Deliverables

- `internal/converter/markdown_to_adf.go`
- `internal/converter/adf_to_markdown.go`
- `internal/converter/adf.go` (ADF type definitions)
- `internal/converter/converter_test.go`
- DR-007: ADF/Markdown Conversion (completed)

## Technical Approach

Use goldmark with GFM extension to parse Markdown into AST, then walk nodes to emit ADF JSON. Pattern adapted from acon project (Confluence converter) but outputs JSON instead of XHTML.

For ADF to Markdown, parse ADF JSON into Go structs, recursively walk nodes, emit Markdown text directly.

## Dependencies

- goldmark library (github.com/yuin/goldmark)
- Reference: acon project `internal/converter/` for pattern guidance
