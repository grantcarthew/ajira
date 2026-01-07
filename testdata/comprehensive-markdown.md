# Ajira Markdown Test Document

This document tests all Markdown features supported by ajira for conversion to Jira's Atlassian Document Format (ADF). Use this to verify Markdown renders correctly in Jira issues and comments.

## Text Formatting

### Basic Formatting

This paragraph contains **bold text**, _italic text_, and _**bold italic text**_ combined. Here is some `inline code` within a sentence.

### Strikethrough

This feature uses ~~strikethrough~~ text for deleted content.

### Mixed Formatting

You can combine **bold** and _italic_ with `inline code` in the same line.

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

### Shell Commands

```bash
#!/bin/bash

# Install dependencies
npm install

# Build the project
go build -o ajira ./cmd/ajira

# Run tests
go test -v ./...
```

### Code Block Without Language

```
This is a plain code block
without any language specified.
It should still be formatted as code.
```

### Terraform (Common for Jira GCP Project)

```hcl
resource "google_compute_instance" "default" {
  name         = "my-instance"
  machine_type = "e2-medium"
  zone         = "australia-southeast1-b"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
    }
  }

  network_interface {
    network = "default"
    access_config {}
  }
}
```

### JSON Example

```json
{
  "name": "ajira",
  "version": "1.0.0",
  "config": {
    "timeout": 30,
    "retries": 3
  }
}
```

### YAML Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
```

## Lists

### Unordered List

- First item
- Second item
- Third item with **bold** and _italic_
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
| Italic | Working | _text_ |
| Code | Working | `code` |

### Table with Alignment

| Left Aligned | Center Aligned | Right Aligned |
|:-------------|:--------------:|--------------:|
| Left | Center | Right |
| Data | Data | Data |
| More | More | More |

### Table with Code

| Command | Description |
|---------|-------------|
| `ajira issue list` | List issues |
| `ajira issue view KEY` | View issue details |
| `ajira issue create -s "Summary"` | Create new issue |

## Links

### External Links

Visit the [Atlassian Documentation](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/) for API details.

Check out the [Go Programming Language](https://go.dev/) official website.

### Multiple Links in Paragraph

Here are useful resources: [GitHub](https://github.com), [Stack Overflow](https://stackoverflow.com), and [Go Docs](https://pkg.go.dev).

### Jira-Style Links (Common in Issues)

Related ticket: [GCP-123](https://autogeneral-au.atlassian.net/browse/GCP-123)

Confluence page: [Project Documentation](https://autogeneral-au.atlassian.net/wiki/spaces/CSA1/overview)

## Blockquotes

### Simple Blockquote

> This is a simple blockquote. It can span multiple lines and is useful for quoting requirements or comments.

### Blockquote with Formatting

> **Note:** Blockquotes can contain **formatting** and `inline code`.

### Nested Blockquotes

> **Note:** ADF does not support nested blockquotes. Blockquote content can only contain paragraphs, lists, code blocks, and media - not other blockquotes. See the [ADF blockquote spec](https://developer.atlassian.com/cloud/jira/platform/apis/document/nodes/blockquote/).

### Blockquote with List

> This blockquote contains a list:
>
> - First item in quote
> - Second item in quote
> - Third item in quote

## Horizontal Rules

Section separator:

---

Another separator:

***

## Special Characters

### Common Characters in Technical Docs

- Ampersands & angle brackets < > work correctly
- Quotes: "double" and 'single'
- Pipes: command1 | command2 | command3
- Backslashes: C:\Users\path\file.txt

### Mathematical and Arrows

- Mathematical: 2 x 3 = 6
- Arrows: -> <- => <= <->
- Comparison: >= <= != ==

### Emoji (if supported)

Common emoji: :rocket: :white_check_mark: :x: :warning:

Unicode emoji: 🚀 ✅ ❌ ⚠️ 💻

## Edge Cases

### Escaped Characters

These should render literally:

\*not italic\* and \`not code\` and \[not a link\](url)

### Inline Code with Special Characters

Command with flags: `go test -v -race ./...`

Path with spaces: `"/path/with spaces/file.txt"`

### Long Lines

This is an extremely long line that should wrap correctly in Jira without breaking the formatting or causing any display issues in the rendered output regardless of the screen width or display settings.

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

---

_Generated for ajira CLI testing - Markdown to ADF conversion verification._
