# DR-007: ADF/Markdown Conversion

- Date: 2025-12-24
- Status: Accepted
- Category: Converter
- Updated: 2026-01-06

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
- Emojis (shortcodes preserved as text)
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

## ADF Specification Constraints

The ADF format has specific structural requirements that affect conversion. These are enforced by Jira's API validation.

### Mark Compatibility

Per the ADF specification, certain marks cannot be combined:

| Mark | Can Combine With |
|------|------------------|
| `code` | `link` only |
| `strong` | `em`, `strike`, `link`, `underline`, `textColor`, `backgroundColor` |
| `em` | `strong`, `strike`, `link`, `underline`, `textColor`, `backgroundColor` |
| `strike` | `strong`, `em`, `link`, `underline`, `textColor`, `backgroundColor` |
| `link` | All marks |

Critical constraint: The `code` mark can ONLY combine with `link`. Combinations like bold+code or italic+code are rejected by Jira with `INVALID_INPUT`.

Converter behavior: When converting Markdown like `**\`code\`**` (bold code), the converter applies only the `code` mark, silently dropping incompatible marks. Code formatting takes precedence.

Reference: https://developer.atlassian.com/cloud/jira/platform/apis/document/marks/code/

### Task Item Structure

Per the ADF specification, `taskItem` nodes must contain inline content directly, not wrapped in block nodes:

Valid structure:
```json
{
  "type": "taskItem",
  "attrs": {"localId": "uuid", "state": "TODO"},
  "content": [
    {"type": "text", "text": "Task description"}
  ]
}
```

Invalid structure (rejected by Jira):
```json
{
  "type": "taskItem",
  "attrs": {"localId": "uuid", "state": "TODO"},
  "content": [
    {
      "type": "paragraph",
      "content": [{"type": "text", "text": "Task description"}]
    }
  ]
}
```

Reference: ADF JSON schema at http://go.atlassian.com/adf-json-schema

### Blockquote Nesting

Blockquotes cannot contain other blockquotes. The allowed content types for `blockquote` are:

- `paragraph`
- `orderedList`
- `bulletList`
- `codeBlock`
- `mediaSingle`
- `mediaGroup`
- `extension`

Markdown nested blockquotes (`> > text`) cannot be represented in ADF and will be rejected.

### Table Alignment

ADF tables do not support column alignment. GFM alignment syntax (`:---`, `:---:`, `---:`) is parsed but the alignment information is lost in conversion.

## Implementation Notes

Location: `internal/converter/`

Package API:

- `MarkdownToADF(markdown string) (*ADF, error)` - Convert Markdown to ADF struct
- `ADFToMarkdown(adfJSON []byte) string` - Convert ADF JSON to Markdown
- `ADFToMarkdownFromStruct(doc *ADF) string` - Convert ADF struct to Markdown

### Markdown to ADF

Parser: goldmark library with GFM extension

Key behaviors:

1. Goldmark splits text at potential emphasis boundaries (underscores), creating multiple text nodes
2. Walk AST nodes and emit ADF JSON structures
3. Hardcode ADF version 1
4. Generate UUIDs for `localId` attributes on task lists/items
5. Check mark compatibility when applying marks to nodes with existing `code` mark

Mark application order:

When nested formatting is encountered (e.g., `**bold with \`code\`**`):

1. Process innermost nodes first
2. When adding outer marks, check if inner node has `code` mark
3. If `code` mark present and new mark is incompatible, skip adding the new mark
4. If `code` mark present and new mark is `link`, add it (allowed combination)

### ADF to Markdown

Key behaviors:

1. Parse ADF JSON into Go structs
2. Merge adjacent text nodes with identical marks before escaping (prevents over-escaping from goldmark splits)
3. Recursively walk nodes and emit Markdown text
4. Skip unsupported node types gracefully (no errors)
5. Escape Markdown special characters using minimal escaping strategy

Text node merging:

Goldmark may split text like `:white_check_mark:` into three nodes (`:white_`, `check_`, `mark:`). Before escaping, adjacent text nodes with identical marks are merged to prevent underscores at node boundaries from being over-escaped.

### Escaping Strategy

The ADF to Markdown converter uses minimal escaping to preserve round-trip fidelity:

Characters escaped:

| Character | When Escaped |
|-----------|--------------|
| `` ` `` | Always (prevents inline code) |
| `*` | Always (prevents emphasis) |
| `[` | Always (prevents links) |
| `_` | Only at word boundaries (not between letters/digits) |

Characters NOT escaped:

| Character | Reason |
|-----------|--------|
| `\` | Escaping causes double-escaping on round-trip |
| `|` | Only meaningful in tables, handled separately |
| `]` | Only meaningful after `[` |
| `#`, `+`, `-`, `!`, `.` | Only special at line start |

Already-escaped detection:

If a character is preceded by a backslash in the source text, it is already escaped and no additional backslash is added. This prevents `\*` from becoming `\\*` on round-trip.

Underscore handling:

Underscores are only escaped when they could trigger emphasis (at word boundaries). An underscore between word characters or other underscores is not escaped:

- `white_check_mark` - underscores NOT escaped (between letters)
- `_italic_` - underscores escaped (at word boundaries)
- `a _ b` - underscore escaped (surrounded by spaces)

### Round-Trip Fidelity

The conversion aims for semantic equivalence, not character-for-character matching.

Expected differences on round-trip:

| Original | After Round-Trip | Reason |
|----------|------------------|--------|
| `__bold__` | `**bold**` | Syntax normalization |
| `***` | `---` | Both valid horizontal rules |
| `\|------\|` | `\| --- \|` | Table separator normalization |
| `:---:` | `---` | Column alignment lost (ADF limitation) |
| 3-space indent | 2-space indent | List indentation normalization |
| `> ` (blank quote line) | `>` + space | Trailing space variation |

Content that round-trips exactly:

- All text content
- Formatting (bold, italic, code, strikethrough)
- Links with URLs
- Code blocks with language
- Lists (ordered, unordered, task)
- Tables (content, not alignment)
- Headings (all levels)
- Blockquotes (non-nested)
- Escaped characters (`\*`, `\``, `\[`)

## Updates

- 2025-12-24: Initial design record
- 2026-01-06: Added ADF specification constraints (mark compatibility, taskItem structure, blockquote nesting), escaping strategy, text node merging, and round-trip fidelity details based on integration testing
