# Test Data and Feature Gap Documentation

This directory contains test fixtures for validating ajira's bidirectional Markdown to ADF (Atlassian Document Format) conversion for Jira issues and comments.

## Quick Start

```bash
# Run the automated round-trip test (recommended)
./testdata/roundtrip-test.sh

# Or manually create a test issue
cat testdata/comprehensive-markdown.md | ./ajira issue create -s "Markdown Test"

# View it back to check round-trip conversion
./ajira issue view PROJ-123
```

## Automated Testing

The `roundtrip-test.sh` script provides comprehensive round-trip testing:

```bash
# Run the test
./testdata/roundtrip-test.sh

# Features:
# - Creates a test issue with comprehensive-markdown.md as description
# - Retrieves and validates 68 feature checks
# - Reports pass/warn/fail status for each feature
# - Opens the issue in your browser for visual review
# - Prompts to delete the test issue (use -y to auto-delete)
```

## Feature Support Matrix

### CLI Commands

| Command | Status | Notes |
| ------- | ------ | ----- |
| `ajira me` | Working | Display current user |
| `ajira project list` | Working | List accessible projects |
| `ajira issue list` | Working | List issues with filters |
| `ajira issue view` | Working | View issue details with ADF to Markdown |
| `ajira issue create` | Working | Create issue with Markdown to ADF |
| `ajira issue edit` | Working | Edit summary/description |
| `ajira issue delete` | Working | Delete an issue |
| `ajira issue assign` | Working | Assign to user/me/unassigned |
| `ajira issue move` | Working | Transition issue status |
| `ajira issue comment add` | Working | Add comment with Markdown |
| `ajira issue type` | Working | List available issue types |
| `ajira issue status` | Working | List available statuses |
| `ajira issue priority` | Working | List available priorities |

### Markdown to ADF Conversion

| Feature | MD to ADF | ADF to MD | Status | Notes |
| ------- | :-------: | :-------: | ------ | ----- |
| **Text Formatting** | | | | |
| Bold `**text**` | Yes | Yes | Working | |
| Italic `*text*` | Yes | Yes | Working | |
| Bold+Italic `***text***` | Yes | Yes | Working | |
| Strikethrough `~~text~~` | Yes | Yes | Working | |
| Inline code `` `code` `` | Yes | Yes | Working | Cannot combine with bold/italic (ADF limit) |
| **Headings** | | | | |
| H1-H6 | Yes | Yes | Working | All heading levels supported |
| **Code Blocks** | | | | |
| Fenced with language | Yes | Yes | Working | Language preserved |
| Fenced without language | Yes | Yes | Working | Renders as plain code |
| Special chars (`<`, `>`, `&`) | Yes | Yes | Working | Properly escaped in code |
| Backslashes and regex | Yes | Yes | Working | Preserved correctly |
| Quotes (escaped) | Yes | Yes | Working | Single, double, escaped |
| Indented (4-space) | Yes | Yes | Working | Converted to fenced block |
| Empty code blocks | Yes | Yes | Working | Uses space placeholder (workaround) |
| **Lists** | | | | |
| Unordered | Yes | Yes | Working | |
| Ordered | Yes | Yes | Working | |
| Nested (3+ levels) | Yes | Yes | Working | |
| Mixed nested | Yes | Yes | Working | |
| Deeply nested (5 levels) | Yes | Yes | Working | |
| Task lists `- [ ]` | Yes | Yes | Working | Checkbox state preserved |
| **Tables** | | | | |
| Basic tables | Yes | Yes | Working | |
| Column alignment | Yes | No | Partial | ADF does not preserve alignment |
| Empty cells | Yes | Yes | Working | |
| Escaped pipes `\|` | Yes | Yes | Working | Pipe content preserved |
| Formatted headers | Yes | Yes | Working | Bold, italic, code in headers |
| Code in cells | Yes | Yes | Working | |
| **Links** | | | | |
| Basic links | Yes | Yes | Working | |
| Multiple links | Yes | Yes | Working | |
| Link titles | Yes | No | Partial | ADF does not preserve titles |
| Links with special chars | Yes | Yes | Working | Query params, anchors |
| AutoLinks `<url>` | Yes | Yes | Working | Converted to regular links |
| Reference-style links | Yes | Yes | Working | Resolved during parse |
| **Blockquotes** | | | | |
| Simple | Yes | Yes | Working | |
| With formatting | Yes | Yes | Working | |
| With lists | Yes | Yes | Working | |
| With code blocks | Yes | Yes | Working | |
| Nested blockquotes | Yes | Yes | Working | Flattened (workaround) |
| **Horizontal Rules** | | | | |
| `---`, `***`, `___` | Yes | Yes | Working | All render as rule |
| **Unicode & Special Chars** | | | | |
| Unicode text (CJK, etc.) | Yes | Yes | Working | Japanese, Chinese, Korean, etc. |
| HTML entities | Yes | Yes | Working | Properly escaped |
| Emoji | Yes | Yes | Working | Unicode emoji supported |
| Mathematical symbols | Yes | Yes | Working | ×, π, ∞, etc. |
| Arrows | Yes | Yes | Working | →, ←, ↑, ↓, etc. |
| **Edge Cases** | | | | |
| Escaped chars `\*` | Yes | Yes | Working | Escapes preserved |
| Hard line breaks | Yes | Yes | Working | Two-space and backslash |
| Double-backtick code | Yes | Yes | Working | Backtick in code spans |
| Consecutive code blocks | Yes | Yes | Working | |
| Long lines | Yes | Yes | Working | Wrapped correctly |
| Paragraph breaks | Yes | Yes | Working | Single vs double breaks |

### Legend

- Yes - Working correctly
- Partial - Works with known limitations
- No - Not supported by ADF

## ADF Limitations

These are limitations of Jira's Atlassian Document Format that cannot be fixed in ajira.

### Inline Code Cannot Combine with Other Marks

ADF's `code` mark can only combine with `link`. Attempting to use bold code like `**`code`**` will render as just `code` without bold. This is an [ADF specification limitation](https://developer.atlassian.com/cloud/jira/platform/apis/document/marks/code/).

### Nested Blockquotes

ADF blockquotes can only contain paragraphs, lists, code blocks, and media - not other blockquotes. See the [ADF blockquote spec](https://developer.atlassian.com/cloud/jira/platform/apis/document/nodes/blockquote/). Ajira works around this by flattening nested blockquotes (see Workarounds Applied below).

### Table Column Alignment

ADF tables do not preserve column alignment markers (`:---`, `:---:`, `---:`). Tables render correctly but alignment is lost on round-trip.

### Link Title Attributes

Markdown link titles `[text](url "title")` are not preserved in ADF. The link works but the title attribute is lost.

### Images

ADF supports images via `mediaSingle` nodes, but these require Jira attachment IDs. External image URLs in Markdown are not supported in issue descriptions.

## Workarounds Applied

These issues were discovered during testing and workarounds have been implemented in the ajira converter.

### Empty Code Blocks

**Issue:** Jira ADF rejects code blocks with empty or whitespace-only text nodes, returning `400 Bad Request - INVALID_INPUT`.

**Workaround:** Empty/whitespace-only code blocks are converted with a single space placeholder. The code block renders but appears empty.

**Location:** `internal/converter/markdown_to_adf.go` - `convertFencedCodeBlock()`

### Nested Blockquotes

**Issue:** Markdown supports nested blockquotes (`> > text`), but ADF does not allow blockquotes inside blockquotes. The API rejects nested blockquote structures with `INVALID_INPUT`.

**Workaround:** Nested blockquotes are flattened - inner blockquote content is extracted and included directly in the parent blockquote. Content is preserved but nesting structure is lost.

**Location:** `internal/converter/markdown_to_adf.go` - `convertBlockquote()`

## Testing Instructions

### Using the Automated Test Script

```bash
# Build ajira first
go build -o ajira ./cmd/ajira

# Run the round-trip test
./testdata/roundtrip-test.sh

# Run with auto-delete (no prompt)
./testdata/roundtrip-test.sh -y

# The script will:
# 1. Create a test issue with comprehensive Markdown
# 2. Retrieve the issue and validate features
# 3. Report pass/warn/fail for 68 feature checks
# 4. Open the issue in your browser
# 5. Prompt to delete the test issue (or auto-delete with -y)
```

### Manual Testing

```bash
# Create test issue
cat testdata/comprehensive-markdown.md | ./ajira issue create -s "Markdown Test"

# View the issue content as Markdown
./ajira issue view PROJ-123

# View raw JSON to inspect ADF
./ajira issue view PROJ-123 --json | jq '.description'

# Test comments
echo "**Bold** and _italic_ test" | ./ajira issue comment add PROJ-123 -f -
./ajira issue view PROJ-123 -c 1
```

### Testing Specific Features

Create focused test issues for isolation:

```bash
# Test just tables
echo '| A | B |
|---|---|
| 1 | 2 |' | ./ajira issue create -s "Table Test"

# Test just code blocks
printf '```go\nfmt.Println("test")\n```' | ./ajira issue create -s "Code Test"

# Test task lists
printf '- [ ] Todo\n- [x] Done' | ./ajira issue create -s "Task Test"

# Test special characters in code
printf '```go\nregexp.MustCompile(`^[A-Z]:\\\\[\\w\\\\]+$`)\n```' | ./ajira issue create -s "Regex Test"
```

## Files

| File | Purpose |
| ---- | ------- |
| `README.md` | This documentation |
| `comprehensive-markdown.md` | Full Markdown feature test document |
| `roundtrip-test.sh` | Automated round-trip testing script |

## References

### Atlassian Documentation

- [Atlassian Document Format](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
- [ADF Nodes Reference](https://developer.atlassian.com/cloud/jira/platform/apis/document/nodes/)
- [ADF Marks Reference](https://developer.atlassian.com/cloud/jira/platform/apis/document/marks/)
- [Jira REST API v3](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/)

### Libraries Used

- [Goldmark](https://github.com/yuin/goldmark) - Markdown parser (with GFM extension)
