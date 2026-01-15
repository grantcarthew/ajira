# Markdown Formatting Reference

ajira uses Markdown for all rich text input (descriptions, comments). This is
standard CommonMark/GitHub-Flavored Markdown, NOT Jira wiki markup.

IMPORTANT: Jira wiki syntax (||, h2., {{}}, *bold*) does NOT work.
Use the Markdown syntax shown below.

## Quick Reference

| Format | Markdown | Result |
|--------|----------|--------|
| Bold | `**text**` | **text** |
| Italic | `*text*` or `_text_` | *text* |
| Bold+Italic | `***text***` | ***text*** |
| Strikethrough | `~~text~~` | ~~text~~ |
| Inline code | `` `code` `` | `code` |
| Link | `[text](url)` | [text](url) |

## Headings

```
# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6
```

## Text Formatting

```
**bold text**
*italic text*
***bold and italic***
~~strikethrough~~
`inline code`
```

## Links

```
[Link text](https://example.com)
[Link with title](https://example.com "Title")
<https://example.com>
```

## Lists

Bullet list:
```
- Item one
- Item two
  - Nested item
  - Another nested
- Item three
```

Numbered list:
```
1. First item
2. Second item
3. Third item
```

Task list:
```
- [ ] Unchecked task
- [x] Completed task
- [ ] Another task
```

## Tables

Tables require a header row, separator row, and data rows:

```
| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |
```

Alignment (optional):
```
| Left | Centre | Right |
|:-----|:------:|------:|
| L    | C      | R     |
```

## Code Blocks

Fenced code block with language:
```
` ` `python
def hello():
    print("Hello, world!")
` ` `
```

(Remove spaces between backticks above)

Indented code block (4 spaces):
```
    function example() {
        return true;
    }
```

## Blockquotes

```
> This is a blockquote.
> It can span multiple lines.
>
> And have multiple paragraphs.
```

## Horizontal Rule

```
---
```

or

```
***
```

## Line Breaks

Single newlines within a paragraph are preserved.

For explicit line breaks, end a line with two spaces or use `\`:
```
Line one
Line two
```

## Escaping Special Characters

Use backslash to escape Markdown characters:
```
\*not italic\*
\# not a heading
\[not a link\]
```

## Jira Wiki vs Markdown Comparison

If you are familiar with Jira wiki markup, use these Markdown equivalents:

| Feature | Jira Wiki (NOT supported) | Markdown (USE THIS) |
|---------|---------------------------|---------------------|
| Bold | `*text*` | `**text**` |
| Italic | `_text_` | `*text*` or `_text_` |
| Heading 1 | `h1. Text` | `# Text` |
| Heading 2 | `h2. Text` | `## Text` |
| Heading 3 | `h3. Text` | `### Text` |
| Bullet list | `* item` | `- item` |
| Numbered list | `# item` | `1. item` |
| Link | `[text\|url]` | `[text](url)` |
| Inline code | `{{code}}` | `` `code` `` |
| Code block | `{code}...{code}` | ``` ` ` ` ... ` ` ` ``` |
| Table header | `\|\|H1\|\|H2\|\|` | `\| H1 \| H2 \|` |
| Table row | `\|C1\|C2\|` | `\| C1 \| C2 \|` |
| Strikethrough | `-text-` | `~~text~~` |
| Quote | `{quote}...{quote}` | `> text` |

## Examples

Create issue with formatted description:
```
ajira issue create -s "New feature" -d "## Overview

This feature adds **user authentication**.

### Requirements

- [ ] Login form
- [ ] Password reset
- [x] Session management

| Component | Status |
|-----------|--------|
| Backend   | Done   |
| Frontend  | WIP    |
"
```

Add formatted comment:
```
ajira issue comment add PROJ-123 "## Update

Fixed the **authentication bug**.

\`\`\`python
def login(user, password):
    return authenticate(user, password)
\`\`\`
"
```

From file:
```
ajira issue create -s "Feature" -f description.md
ajira issue comment add PROJ-123 -f comment.md
```
