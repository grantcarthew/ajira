# Ajira Comprehensive Test Document

This document tests all Markdown features supported by ajira for conversion to Jira's Atlassian Document Format (ADF). Use this to verify Markdown renders correctly in Jira issues and comments.

## Text Formatting

### Basic Formatting

This paragraph contains **bold text**, *italic text*, and ***bold italic text*** combined. Here is some `inline code` within a sentence.

### Strikethrough

This feature uses ~~strikethrough~~ text for deleted content.

### Mixed Formatting

You can combine **bold with `inline code`** and *italic with `inline code`* in the same line.

> **Note:** ADF does not support combining `code` marks with other formatting marks like bold or italic. The `code` mark can only combine with `link`. For example, `**`code`**` will render as just `code` without bold formatting. This is an [ADF specification limitation](https://developer.atlassian.com/cloud/jira/platform/apis/document/marks/code/).

## Headings

### This is H3

#### This is H4

##### This is H5

###### This is H6

## Code Blocks

### Go Code

```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    message := "Hello, World!"
    fmt.Println(strings.ToUpper(message))
}
```

### Python Code

```python
def fibonacci(n):
    """Generate Fibonacci sequence up to n terms."""
    a, b = 0, 1
    result = []
    for _ in range(n):
        result.append(a)
        a, b = b, a + b
    return result

if __name__ == "__main__":
    print(fibonacci(10))
```

### JavaScript Code

```javascript
const fetchData = async (url) => {
    try {
        const response = await fetch(url);
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('Error:', error);
        throw error;
    }
};
```

### Code with Special Characters

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Test & Demo</title>
</head>
<body>
    <div class="container">
        <p>if (a < b && c > d) { /* comment */ }</p>
    </div>
</body>
</html>
```

### Code with Backslashes and Regex

```go
func validatePath(path string) bool {
    // Windows paths: C:\Users\test\file.txt
    winRegex := regexp.MustCompile(`^[A-Z]:\\[\w\\]+$`)
    // Unix paths: /home/user/file.txt
    unixRegex := regexp.MustCompile(`^/[\w/]+$`)
    return winRegex.MatchString(path) || unixRegex.MatchString(path)
}
```

### Code with Quotes

```python
def process_text(text):
    single = 'Hello'
    double = "World"
    escaped = "She said \"Hello\" and 'waved'"
    multiline = """
    This is a
    multiline string
    """
    return f"{single} {double}: {escaped}"
```

### Code Block Without Language

```
This is a plain code block
without any language specified.
It should still be formatted as code.
```

### Shell Commands

```bash
#!/bin/bash

# Install dependencies
npm install

# Build the project
go build -o ajira -ldflags "-X main.version=v1.0.0"

# Run tests
go test -v ./... 2>&1 | tee test-output.log

# Check exit status
if [ $? -eq 0 ]; then
    echo "All tests passed!"
else
    echo "Some tests failed."
    exit 1
fi
```

### Indented Code Block (4-space prefix)

    This is an indented code block
    created with 4 spaces instead of fences.
    It should render as code without a language.

### Empty Code Block

```
```

### Code Block with Only Whitespace

```

```

## Lists

### Unordered List

- First item
- Second item
- Third item with **bold** and *italic*
- Fourth item with `inline code`

### Ordered List

1. Step one
2. Step two
3. Step three
4. Step four

### Nested Lists

- Parent item one
  - Child item 1.1
  - Child item 1.2
- Parent item two
  - Child item 2.1
    - Grandchild 2.1.1
    - Grandchild 2.1.2
  - Child item 2.2

### Mixed Nested Lists

1. Ordered parent one
   - Unordered child A
   - Unordered child B
2. Ordered parent two
   1. Nested ordered 2.1
   2. Nested ordered 2.2
      - Mixed deep nesting
      - Another mixed item

### Deeply Nested Lists (5 levels)

- Level 1
  - Level 2
    - Level 3
      - Level 4
        - Level 5

### Task Lists (GFM Checkboxes)

- [ ] Unchecked task item
- [x] Checked/completed task
- [ ] Another pending task
- [x] Another completed task

## Tables

### Simple Table

| Feature | Status | Notes |
|---------|--------|-------|
| Headings | Working | H1-H6 |
| Bold | Working | **text** |
| Italic | Working | *text* |
| Code | Working | `code` |

### Table with Alignment

| Left Aligned | Center Aligned | Right Aligned |
|:-------------|:--------------:|--------------:|
| Left | Center | Right |
| Data | Data | Data |
| More | More | More |

### Table with Complex Content

| Language | Example | Description |
|----------|---------|-------------|
| Go | `fmt.Println()` | Print with newline |
| Python | `print()` | Standard output |
| JavaScript | `console.log()` | Browser console |
| Rust | `println!()` | Macro-based print |

### Table with Empty Cells

| Column A | Column B | Column C |
|----------|----------|----------|
| Has data | | Empty middle |
| | Empty start | Has data |
| Data | Data | |

### Table with Escaped Pipes

| Expression | Result | Notes |
|------------|--------|-------|
| a \| b | OR operation | Logical or |
| x \| y \| z | Multiple | Chained |

### Table with Formatted Headers

| **Bold Header** | *Italic Header* | `Code Header` |
|-----------------|-----------------|---------------|
| normal data | normal data | normal data |

## Links

### External Links

Visit the [Atlassian Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/) for API details.

Check out the [Go Programming Language](https://go.dev/) official website.

### Multiple Links in Paragraph

Here are some useful resources: [GitHub](https://github.com), [Stack Overflow](https://stackoverflow.com), and [Go Docs](https://pkg.go.dev).

### Link with Title (may be lost in ADF)

[Go Documentation](https://go.dev/doc/ "Official Go documentation and tutorials")

### Links with Special Characters

[Search with params](https://example.com/search?q=foo&bar=baz#anchor)

[URL with spaces](https://example.com/path%20with%20spaces)

### AutoLinks

<https://example.com>

<https://go.dev/doc/>

<user@example.com>

### Reference-Style Links

Here is a [reference link][reflink] and another [link][golang].

[reflink]: https://example.com/reference "Reference Example"
[golang]: https://go.dev

## Blockquotes

### Simple Blockquote

> This is a simple blockquote. It can span multiple lines and is useful for quoting requirements or comments.

### Blockquote with Formatting

> **Note:** Blockquotes can contain **formatting** and `inline code`.

### Nested Blockquotes

> Outer blockquote level one
> > Inner blockquote level two
> > > Deeply nested level three

> **Note:** ADF does not support nested blockquotes. They will be flattened. See the [ADF blockquote spec](https://developer.atlassian.com/cloud/jira/platform/apis/document/nodes/blockquote/).

### Blockquote with List

> This blockquote contains a list:
>
> - First item in quote
> - Second item in quote
> - Third item in quote

### Blockquote with Code

> Here is some code in a blockquote:
>
> ```go
> fmt.Println("Hello from blockquote")
> ```

## Horizontal Rules

Various horizontal rule syntaxes:

---

***

___

## Unicode and Special Characters

### Unicode Text

- Japanese: こんにちは世界
- Chinese: 你好世界
- Korean: 안녕하세요
- Russian: Привет мир
- Arabic: مرحبا بالعالم
- Emoji: 🚀 🎉 ✨ 💻 🔧

### Special Characters in Text

Ampersands & angle brackets < > and quotes "double" 'single' work correctly.

Mathematical: 2 × 3 = 6, π ≈ 3.14159, ∞ infinity

Arrows: → ← ↑ ↓ ↔ ⇒ ⇐

## Edge Cases

### Escaped Characters

These should render literally, not as formatting:

\*not italic\* and \`not code\` and \[not a link\](url)

\# Not a heading

\- Not a list item

### Hard Line Breaks

Line with two trailing spaces
forces a line break here.

Line with backslash\
also forces a break.

### Double-Backtick Code Spans

Code with backtick inside: ``code with ` backtick``

Multiple backticks: ``` `multiple` backticks ```

### Empty Sections

This section has no special formatting.

### Long Lines

This is an extremely long line that should wrap correctly in both Jira and when converted back to Markdown without breaking the formatting or causing any display issues in the rendered output regardless of the screen width or display settings being used by the reader of this document.

### Consecutive Code Blocks

```go
// First code block
func first() {}
```

```python
# Second code block immediately after
def second():
    pass
```

```javascript
// Third code block
const third = () => {};
```

### Paragraphs with Single vs Double Line Breaks

This is paragraph one.
This is still paragraph one (single line break).

This is paragraph two (double line break above).

---

*Generated for ajira CLI testing - Markdown to ADF conversion verification.*
