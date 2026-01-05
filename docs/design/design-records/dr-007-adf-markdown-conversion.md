# DR-007: ADF/Markdown Conversion

- Date: 2025-12-24
- Status: Accepted
- Category: Converter

## Problem

The Jira API requires Atlassian Document Format (ADF) for rich text fields (descriptions, comments). ADF is a JSON-based document format that is verbose and unfriendly for humans and AI agents.

Users and agents should interact with ajira using Markdown only. The ADF conversion must be invisible.

## Decision

Implement bidirectional Markdown/ADF conversion as an internal layer:

- All user input accepts Markdown
- All output displays Markdown
- ADF conversion happens transparently at the API boundary

Supported Markdown elements with 1:1 ADF mappings:

| Markdown | ADF Node |
|----------|----------|
| Paragraphs | `paragraph` |
| Headings (h1-h6) | `heading` with `level` attr |
| Bold | `strong` mark |
| Italic | `em` mark |
| Strikethrough | `strike` mark |
| Inline code | `code` mark |
| Code blocks | `codeBlock` with `language` attr |
| Links | `link` mark |
| Ordered lists | `orderedList` |
| Unordered lists | `bulletList` |
| Task lists | `taskList`/`taskItem` with `state` attr (`TODO`/`DONE`) |
| Nested lists | Supported (arbitrary depth) |
| Tables | `table`/`tableRow`/`tableCell`/`tableHeader` |
| Blockquotes | `blockquote` |
| Horizontal rules | `rule` |
| Line breaks | `hardBreak` |

Unsupported (out of scope):

- Images/media (requires separate upload API)
- Mentions (@user)
- Emojis
- Panels, expands, layouts
- Issue links, Jira macros
- Nested tables

## Why

- Markdown is universally understood by humans and AI agents
- ADF is Jira-specific and verbose
- Clean 1:1 mappings exist for all standard Markdown elements
- No "fancy footwork" required (unlike Confluence Storage Format)
- Keeps the CLI simple and predictable

## Trade-offs

Accept:

- Cannot use advanced Jira formatting features (panels, expands)
- Images require separate handling outside this converter
- Some ADF content from Jira may render as best-effort Markdown

Gain:

- Simple, predictable Markdown interface
- No learning curve for users
- AI agents work naturally with Markdown
- Testable with standard Markdown samples

## Alternatives

Pass-through ADF:

- Pro: Full Jira feature support
- Con: Users must learn ADF JSON structure
- Con: AI agents struggle with verbose JSON
- Rejected: Defeats purpose of human-friendly CLI

Limited Markdown subset:

- Pro: Even simpler implementation
- Con: Missing common features (tables, task lists)
- Rejected: Standard Markdown is expected

## Implementation Notes

Location: `internal/converter/`

Package API:

- Two simple functions: `MarkdownToADF` returns ADF and error, `ADFToMarkdown` returns string
- Error on parse failure for Markdown to ADF; best-effort for ADF to Markdown

Markdown to ADF:

- Use goldmark library with GFM extension for parsing
- Walk AST nodes and emit ADF JSON structures
- Pattern from acon project (Confluence converter) applies
- Hardcode ADF version 1

ADF to Markdown:

- Parse ADF JSON into Go structs
- Recursively walk nodes and emit Markdown text
- Skip unsupported node types gracefully (no errors)
- Escape Markdown special characters in text content to prevent accidental formatting

Line breaks:

- Only explicit hard breaks (two trailing spaces or `<br>`) map to ADF `hardBreak`
- Soft breaks become spaces per standard Markdown behaviour

Task lists:

- Checkbox states map to `state` attribute: unchecked is `TODO`, checked is `DONE`
- Generate UUID for required `localId` attribute on `taskList` and `taskItem`
- Limitation: Updating content regenerates `localId` values; document this behaviour

Tables:

- GFM column alignment syntax silently ignored (ADF tables lack alignment support)

Inline HTML:

- Extract text content from HTML tags, discard the tags
- Preserves content, loses HTML formatting

Round-trip fidelity:

- Semantic equivalence, not character-for-character
- Alternative Markdown syntaxes may normalise (e.g., `__bold__` becomes `**bold**`)
- Whitespace may normalise
- Edge cases expected and accepted
