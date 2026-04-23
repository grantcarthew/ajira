# Markdown Formatting Reference

ajira input is CommonMark + GFM, converted to Atlassian Document Format (ADF). Jira wiki syntax does NOT render.

## Cheatsheet

| Format | Markdown | Jira Wiki (NOT supported) |
|--------|----------|---------------------------|
| Bold | `**text**` | `*text*` |
| Italic | `*text*` or `_text_` | `_text_` |
| Strikethrough | `~~text~~` | `-text-` |
| Inline code | `` `code` `` | `{{code}}` |
| Heading | `# Text` … `###### Text` | `h1.` … `h6.` |
| Link | `[text](url)` | `[text\|url]` |
| Bullet | `- item` | `* item` |
| Numbered | `1. item` | `# item` |
| Task | `- [ ] todo` / `- [x] done` | — |
| Quote | `> text` | `{quote}...{quote}` |
| Rule | `---` or `***` | `----` |

## Tables

```
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
```

## Code Blocks

Fenced with triple backticks; language tag preserved:

````
```python
def hello():
    print("Hi")
```
````

## Gotchas

- Images (`![alt](url)`) drop the image; only alt text kept
- HTML tags stripped; text content kept
- Nested blockquotes flattened to single level
- Table alignment syntax (`|:---:|`) ignored
- Soft line breaks become hard breaks
- Task items: inline content only, no nested blocks

## Example

```
ajira issue create -s "New feature" -d "## Overview

Adds **user auth**.

- [ ] Login form
- [x] Session management

| Component | Status |
|-----------|--------|
| Backend   | Done   |
| Frontend  | WIP    |
"
```

Or from file: `ajira issue create -s "Feature" -f description.md`
