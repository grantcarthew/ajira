package converter

import (
	"encoding/json"
	"strings"
	"testing"
)

// Test helpers

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return data
}

// MarkdownToADF tests

func TestMarkdownToADF_Paragraph(t *testing.T) {
	md := "Hello world"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if adf.Version != 1 {
		t.Errorf("expected version 1, got %d", adf.Version)
	}
	if adf.Type != NodeTypeDoc {
		t.Errorf("expected type 'doc', got %q", adf.Type)
	}
	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	para := adf.Content[0]
	if para.Type != NodeTypeParagraph {
		t.Errorf("expected paragraph, got %q", para.Type)
	}
	if len(para.Content) == 0 {
		t.Fatal("expected at least 1 text node")
	}
	// Combine all text content
	var text string
	for _, node := range para.Content {
		text += node.Text
	}
	if text != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", text)
	}
}

func TestMarkdownToADF_Headings(t *testing.T) {
	tests := []struct {
		md    string
		level int
	}{
		{"# Heading 1", 1},
		{"## Heading 2", 2},
		{"### Heading 3", 3},
		{"#### Heading 4", 4},
		{"##### Heading 5", 5},
		{"###### Heading 6", 6},
	}

	for _, tc := range tests {
		t.Run(tc.md, func(t *testing.T) {
			adf, err := MarkdownToADF(tc.md)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(adf.Content) != 1 {
				t.Fatalf("expected 1 content node, got %d", len(adf.Content))
			}

			heading := adf.Content[0]
			if heading.Type != NodeTypeHeading {
				t.Errorf("expected heading, got %q", heading.Type)
			}

			level, ok := heading.Attrs["level"].(int)
			if !ok {
				t.Fatalf("level attribute not found or wrong type")
			}
			if level != tc.level {
				t.Errorf("expected level %d, got %d", tc.level, level)
			}
		})
	}
}

func TestMarkdownToADF_Bold(t *testing.T) {
	md := "This is **bold** text"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	foundBold := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrong {
				foundBold = true
				if node.Text != "bold" {
					t.Errorf("expected bold text 'bold', got %q", node.Text)
				}
			}
		}
	}
	if !foundBold {
		t.Error("expected to find bold mark")
	}
}

func TestMarkdownToADF_Italic(t *testing.T) {
	md := "This is *italic* text"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	foundItalic := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeEm {
				foundItalic = true
				if node.Text != "italic" {
					t.Errorf("expected italic text 'italic', got %q", node.Text)
				}
			}
		}
	}
	if !foundItalic {
		t.Error("expected to find italic mark")
	}
}

func TestMarkdownToADF_Strikethrough(t *testing.T) {
	md := "This is ~~deleted~~ text"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	foundStrike := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrike {
				foundStrike = true
				if node.Text != "deleted" {
					t.Errorf("expected strikethrough text 'deleted', got %q", node.Text)
				}
			}
		}
	}
	if !foundStrike {
		t.Error("expected to find strikethrough mark")
	}
}

func TestMarkdownToADF_InlineCode(t *testing.T) {
	md := "Use the `fmt.Println` function"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	foundCode := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeCode {
				foundCode = true
				if node.Text != "fmt.Println" {
					t.Errorf("expected code text 'fmt.Println', got %q", node.Text)
				}
			}
		}
	}
	if !foundCode {
		t.Error("expected to find code mark")
	}
}

func TestMarkdownToADF_CodeBlock(t *testing.T) {
	md := "```go\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n```"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	codeBlock := adf.Content[0]
	if codeBlock.Type != NodeTypeCodeBlock {
		t.Errorf("expected codeBlock, got %q", codeBlock.Type)
	}

	lang, ok := codeBlock.Attrs["language"].(string)
	if !ok || lang != "go" {
		t.Errorf("expected language 'go', got %q", lang)
	}
}

func TestMarkdownToADF_Link(t *testing.T) {
	md := "Check out [Google](https://google.com)"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
				href, _ := mark.Attrs["href"].(string)
				if href != "https://google.com" {
					t.Errorf("expected href 'https://google.com', got %q", href)
				}
				if node.Text != "Google" {
					t.Errorf("expected link text 'Google', got %q", node.Text)
				}
			}
		}
	}
	if !foundLink {
		t.Error("expected to find link mark")
	}
}

func TestMarkdownToADF_BulletList(t *testing.T) {
	md := "- Item 1\n- Item 2\n- Item 3"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	list := adf.Content[0]
	if list.Type != NodeTypeBulletList {
		t.Errorf("expected bulletList, got %q", list.Type)
	}
	if len(list.Content) != 3 {
		t.Errorf("expected 3 list items, got %d", len(list.Content))
	}
}

func TestMarkdownToADF_OrderedList(t *testing.T) {
	md := "1. First\n2. Second\n3. Third"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	list := adf.Content[0]
	if list.Type != NodeTypeOrderedList {
		t.Errorf("expected orderedList, got %q", list.Type)
	}
	if len(list.Content) != 3 {
		t.Errorf("expected 3 list items, got %d", len(list.Content))
	}
}

func TestMarkdownToADF_TaskList(t *testing.T) {
	md := "- [ ] Todo item\n- [x] Done item"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	list := adf.Content[0]
	if list.Type != NodeTypeTaskList {
		t.Errorf("expected taskList, got %q", list.Type)
	}
	if len(list.Content) != 2 {
		t.Errorf("expected 2 task items, got %d", len(list.Content))
	}

	// Check first item is TODO
	if state, ok := list.Content[0].Attrs["state"].(string); !ok || state != TaskStateTODO {
		t.Errorf("expected first item state TODO, got %v", list.Content[0].Attrs["state"])
	}

	// Check second item is DONE
	if state, ok := list.Content[1].Attrs["state"].(string); !ok || state != TaskStateDONE {
		t.Errorf("expected second item state DONE, got %v", list.Content[1].Attrs["state"])
	}

	// Check localId is present
	if _, ok := list.Attrs["localId"].(string); !ok {
		t.Error("expected taskList to have localId")
	}
}

func TestMarkdownToADF_NestedList(t *testing.T) {
	md := "- Item 1\n  - Nested 1\n  - Nested 2\n- Item 2"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list := adf.Content[0]
	if list.Type != NodeTypeBulletList {
		t.Errorf("expected bulletList, got %q", list.Type)
	}

	// First item should have nested list
	firstItem := list.Content[0]
	foundNestedList := false
	for _, child := range firstItem.Content {
		if child.Type == NodeTypeBulletList {
			foundNestedList = true
			if len(child.Content) != 2 {
				t.Errorf("expected 2 nested items, got %d", len(child.Content))
			}
		}
	}
	if !foundNestedList {
		t.Error("expected to find nested list")
	}
}

func TestMarkdownToADF_Table(t *testing.T) {
	md := "| Header 1 | Header 2 |\n|----------|----------|\n| Cell 1   | Cell 2   |"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	table := adf.Content[0]
	if table.Type != NodeTypeTable {
		t.Errorf("expected table, got %q", table.Type)
	}
	if len(table.Content) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.Content))
	}

	// Check header row
	headerRow := table.Content[0]
	if headerRow.Type != NodeTypeTableRow {
		t.Errorf("expected tableRow, got %q", headerRow.Type)
	}
	if len(headerRow.Content) != 2 {
		t.Errorf("expected 2 header cells, got %d", len(headerRow.Content))
	}
	if headerRow.Content[0].Type != NodeTypeTableHeader {
		t.Errorf("expected tableHeader, got %q", headerRow.Content[0].Type)
	}
}

func TestMarkdownToADF_Blockquote(t *testing.T) {
	md := "> This is a quote"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	quote := adf.Content[0]
	if quote.Type != NodeTypeBlockquote {
		t.Errorf("expected blockquote, got %q", quote.Type)
	}
}

func TestMarkdownToADF_HorizontalRule(t *testing.T) {
	md := "Above\n\n---\n\nBelow"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundRule := false
	for _, node := range adf.Content {
		if node.Type == NodeTypeRule {
			foundRule = true
		}
	}
	if !foundRule {
		t.Error("expected to find horizontal rule")
	}
}

// ADFToMarkdown tests

func TestADFToMarkdown_Paragraph(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Hello, world!"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "Hello, world!") {
		t.Errorf("expected markdown to contain 'Hello, world!', got %q", md)
	}
}

func TestADFToMarkdown_Heading(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{1, "# Test"},
		{2, "## Test"},
		{3, "### Test"},
		{4, "#### Test"},
		{5, "##### Test"},
		{6, "###### Test"},
	}

	for _, tc := range tests {
		adf := &ADF{
			Version: 1,
			Type:    NodeTypeDoc,
			Content: []ADFNode{
				{
					Type:  NodeTypeHeading,
					Attrs: map[string]any{"level": tc.level},
					Content: []ADFNode{
						{Type: NodeTypeText, Text: "Test"},
					},
				},
			},
		}

		md := ADFToMarkdownFromStruct(adf)
		if !strings.HasPrefix(md, tc.expected) {
			t.Errorf("level %d: expected prefix %q, got %q", tc.level, tc.expected, md)
		}
	}
}

func TestADFToMarkdown_Bold(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "This is "},
					{Type: NodeTypeText, Text: "bold", Marks: []ADFMark{{Type: MarkTypeStrong}}},
					{Type: NodeTypeText, Text: " text"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected markdown to contain '**bold**', got %q", md)
	}
}

func TestADFToMarkdown_Italic(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "This is "},
					{Type: NodeTypeText, Text: "italic", Marks: []ADFMark{{Type: MarkTypeEm}}},
					{Type: NodeTypeText, Text: " text"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "*italic*") {
		t.Errorf("expected markdown to contain '*italic*', got %q", md)
	}
}

func TestADFToMarkdown_CodeBlock(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:  NodeTypeCodeBlock,
				Attrs: map[string]any{"language": "go"},
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "fmt.Println(\"Hello\")"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "```go") {
		t.Errorf("expected markdown to contain '```go', got %q", md)
	}
	if !strings.Contains(md, "fmt.Println") {
		t.Errorf("expected markdown to contain code, got %q", md)
	}
}

func TestADFToMarkdown_Link(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{
						Type: NodeTypeText,
						Text: "Google",
						Marks: []ADFMark{
							{Type: MarkTypeLink, Attrs: map[string]any{"href": "https://google.com"}},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "[Google](https://google.com)") {
		t.Errorf("expected markdown to contain link, got %q", md)
	}
}

func TestADFToMarkdown_BulletList(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeBulletList,
				Content: []ADFNode{
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Item 1"}}},
						},
					},
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Item 2"}}},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "- Item 1") {
		t.Errorf("expected markdown to contain '- Item 1', got %q", md)
	}
	if !strings.Contains(md, "- Item 2") {
		t.Errorf("expected markdown to contain '- Item 2', got %q", md)
	}
}

func TestADFToMarkdown_OrderedList(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeOrderedList,
				Content: []ADFNode{
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "First"}}},
						},
					},
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Second"}}},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "1. First") {
		t.Errorf("expected markdown to contain '1. First', got %q", md)
	}
	if !strings.Contains(md, "2. Second") {
		t.Errorf("expected markdown to contain '2. Second', got %q", md)
	}
}

func TestADFToMarkdown_TaskList(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:  NodeTypeTaskList,
				Attrs: map[string]any{"localId": "test-id"},
				Content: []ADFNode{
					{
						Type:  NodeTypeTaskItem,
						Attrs: map[string]any{"localId": "item-1", "state": TaskStateTODO},
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Todo"}}},
						},
					},
					{
						Type:  NodeTypeTaskItem,
						Attrs: map[string]any{"localId": "item-2", "state": TaskStateDONE},
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Done"}}},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "- [ ] Todo") {
		t.Errorf("expected markdown to contain '- [ ] Todo', got %q", md)
	}
	if !strings.Contains(md, "- [x] Done") {
		t.Errorf("expected markdown to contain '- [x] Done', got %q", md)
	}
}

func TestADFToMarkdown_Table(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeTable,
				Content: []ADFNode{
					{
						Type: NodeTypeTableRow,
						Content: []ADFNode{
							{
								Type: NodeTypeTableHeader,
								Content: []ADFNode{
									{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Header 1"}}},
								},
							},
							{
								Type: NodeTypeTableHeader,
								Content: []ADFNode{
									{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Header 2"}}},
								},
							},
						},
					},
					{
						Type: NodeTypeTableRow,
						Content: []ADFNode{
							{
								Type: NodeTypeTableCell,
								Content: []ADFNode{
									{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Cell 1"}}},
								},
							},
							{
								Type: NodeTypeTableCell,
								Content: []ADFNode{
									{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Cell 2"}}},
								},
							},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "| Header 1") {
		t.Errorf("expected markdown to contain table header, got %q", md)
	}
	if !strings.Contains(md, "| --- |") {
		t.Errorf("expected markdown to contain table separator, got %q", md)
	}
	if !strings.Contains(md, "| Cell 1") {
		t.Errorf("expected markdown to contain table cell, got %q", md)
	}
}

func TestADFToMarkdown_Blockquote(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeBlockquote,
				Content: []ADFNode{
					{
						Type: NodeTypeParagraph,
						Content: []ADFNode{
							{Type: NodeTypeText, Text: "This is a quote"},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "> This is a quote") {
		t.Errorf("expected markdown to contain blockquote, got %q", md)
	}
}

func TestADFToMarkdown_Rule(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: NodeTypeRule},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "---") {
		t.Errorf("expected markdown to contain '---', got %q", md)
	}
}

func TestADFToMarkdown_FromJSON(t *testing.T) {
	adfJSON := `{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{"type": "text", "text": "Hello from JSON"}
				]
			}
		]
	}`

	md := ADFToMarkdown([]byte(adfJSON))
	if !strings.Contains(md, "Hello from JSON") {
		t.Errorf("expected markdown to contain 'Hello from JSON', got %q", md)
	}
}

func TestADFToMarkdown_UnsupportedNodeSkipped(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: "unsupported_type"},
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Valid paragraph"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "Valid paragraph") {
		t.Errorf("expected markdown to contain valid content, got %q", md)
	}
}

// =============================================================================
// Markdown to ADF Edge Cases
// =============================================================================

func TestMarkdownToADF_EmptyInput(t *testing.T) {
	adf, err := MarkdownToADF("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adf.Version != 1 {
		t.Errorf("expected version 1, got %d", adf.Version)
	}
	if adf.Type != NodeTypeDoc {
		t.Errorf("expected type 'doc', got %q", adf.Type)
	}
	if len(adf.Content) != 0 {
		t.Errorf("expected 0 content nodes for empty input, got %d", len(adf.Content))
	}
}

func TestMarkdownToADF_WhitespaceOnly(t *testing.T) {
	tests := []string{
		" ",
		"   ",
		"\t",
		"\n",
		"\n\n\n",
		"  \n  \n  ",
	}
	for _, input := range tests {
		adf, err := MarkdownToADF(input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", input, err)
		}
		// Whitespace-only should produce empty or minimal content
		if adf.Version != 1 || adf.Type != NodeTypeDoc {
			t.Errorf("invalid ADF structure for whitespace input")
		}
	}
}

func TestMarkdownToADF_MultipleParagraphs(t *testing.T) {
	md := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(adf.Content) != 3 {
		t.Errorf("expected 3 paragraphs, got %d", len(adf.Content))
	}
	for i, node := range adf.Content {
		if node.Type != NodeTypeParagraph {
			t.Errorf("node %d: expected paragraph, got %q", i, node.Type)
		}
	}
}

func TestMarkdownToADF_NestedFormatting_BoldInItalic(t *testing.T) {
	md := "*italic with **bold** inside*"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundBold := false
	foundItalic := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrong {
				foundBold = true
			}
			if mark.Type == MarkTypeEm {
				foundItalic = true
			}
		}
	}
	if !foundBold {
		t.Error("expected to find bold mark")
	}
	if !foundItalic {
		t.Error("expected to find italic mark")
	}
}

func TestMarkdownToADF_NestedFormatting_ItalicInBold(t *testing.T) {
	md := "**bold with *italic* inside**"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundBold := false
	foundItalic := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrong {
				foundBold = true
			}
			if mark.Type == MarkTypeEm {
				foundItalic = true
			}
		}
	}
	if !foundBold || !foundItalic {
		t.Error("expected to find both bold and italic marks")
	}
}

func TestMarkdownToADF_TripleEmphasis(t *testing.T) {
	md := "***bold and italic***"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundBold := false
	foundItalic := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrong {
				foundBold = true
			}
			if mark.Type == MarkTypeEm {
				foundItalic = true
			}
		}
	}
	if !foundBold || !foundItalic {
		t.Error("expected *** to produce both bold and italic")
	}
}

func TestMarkdownToADF_AlternativeBoldSyntax(t *testing.T) {
	md := "This is __bold__ text"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundBold := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrong {
				foundBold = true
			}
		}
	}
	if !foundBold {
		t.Error("expected __ syntax to produce bold")
	}
}

func TestMarkdownToADF_AlternativeItalicSyntax(t *testing.T) {
	md := "This is _italic_ text"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundItalic := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeEm {
				foundItalic = true
			}
		}
	}
	if !foundItalic {
		t.Error("expected _ syntax to produce italic")
	}
}

func TestMarkdownToADF_HeadingWithFormatting(t *testing.T) {
	md := "## Heading with **bold** and *italic*"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	heading := adf.Content[0]
	if heading.Type != NodeTypeHeading {
		t.Fatalf("expected heading, got %q", heading.Type)
	}
	foundBold := false
	foundItalic := false
	for _, node := range heading.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrong {
				foundBold = true
			}
			if mark.Type == MarkTypeEm {
				foundItalic = true
			}
		}
	}
	if !foundBold || !foundItalic {
		t.Error("expected heading to contain bold and italic marks")
	}
}

func TestMarkdownToADF_HeadingWithCode(t *testing.T) {
	md := "## Heading with `code`"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	heading := adf.Content[0]
	foundCode := false
	for _, node := range heading.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeCode {
				foundCode = true
			}
		}
	}
	if !foundCode {
		t.Error("expected heading to contain code mark")
	}
}

func TestMarkdownToADF_HeadingWithLink(t *testing.T) {
	md := "## Heading with [link](https://example.com)"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	heading := adf.Content[0]
	foundLink := false
	for _, node := range heading.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
			}
		}
	}
	if !foundLink {
		t.Error("expected heading to contain link mark")
	}
}

func TestMarkdownToADF_CodeBlockWithoutLanguage(t *testing.T) {
	md := "```\nplain code\n```"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	codeBlock := adf.Content[0]
	if codeBlock.Type != NodeTypeCodeBlock {
		t.Fatalf("expected codeBlock, got %q", codeBlock.Type)
	}
	// Should have no language attribute or empty language
	if lang, ok := codeBlock.Attrs["language"]; ok && lang != "" {
		t.Errorf("expected no language, got %q", lang)
	}
}

func TestMarkdownToADF_CodeBlockWithSpecialChars(t *testing.T) {
	md := "```go\nfunc() { return \"<>&\\\"'\" }\n```"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	codeBlock := adf.Content[0]
	if len(codeBlock.Content) == 0 {
		t.Fatal("expected code content")
	}
	text := codeBlock.Content[0].Text
	if !strings.Contains(text, "<>&") {
		t.Errorf("expected special chars to be preserved, got %q", text)
	}
}

func TestMarkdownToADF_CodeBlockMultiline(t *testing.T) {
	md := "```python\ndef hello():\n    print(\"Hello\")\n    print(\"World\")\n```"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	codeBlock := adf.Content[0]
	text := codeBlock.Content[0].Text
	lines := strings.Split(text, "\n")
	if len(lines) < 3 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

func TestMarkdownToADF_IndentedCodeBlock(t *testing.T) {
	md := "    indented code\n    more code"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(adf.Content) == 0 {
		t.Fatal("expected content")
	}
	// Indented code blocks should produce codeBlock
	codeBlock := adf.Content[0]
	if codeBlock.Type != NodeTypeCodeBlock {
		t.Errorf("expected codeBlock for indented code, got %q", codeBlock.Type)
	}
}

func TestMarkdownToADF_InlineCodeWithSpecialChars(t *testing.T) {
	md := "Use `<div class=\"test\">`"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundCode := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeCode {
				foundCode = true
				if !strings.Contains(node.Text, "<div") {
					t.Errorf("expected code to contain HTML, got %q", node.Text)
				}
			}
		}
	}
	if !foundCode {
		t.Error("expected to find code mark")
	}
}

func TestMarkdownToADF_InlineCodeWithBackticks(t *testing.T) {
	md := "Use `` `backticks` `` in code"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundCode := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeCode {
				foundCode = true
			}
		}
	}
	if !foundCode {
		t.Error("expected to find code mark")
	}
}

func TestMarkdownToADF_LinkWithSpecialCharsInURL(t *testing.T) {
	md := "[Search](https://example.com/search?q=hello+world&lang=en)"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				href := mark.Attrs["href"].(string)
				if !strings.Contains(href, "q=hello+world") {
					t.Errorf("expected URL with query params, got %q", href)
				}
			}
		}
	}
}

func TestMarkdownToADF_LinkWithParensInURL(t *testing.T) {
	md := "[Wiki](https://en.wikipedia.org/wiki/Go_(programming_language))"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
				href := mark.Attrs["href"].(string)
				if !strings.Contains(href, "Go_(programming_language)") {
					t.Errorf("expected URL with parens, got %q", href)
				}
			}
		}
	}
	if !foundLink {
		t.Error("expected to find link")
	}
}

func TestMarkdownToADF_AutoLink(t *testing.T) {
	md := "Visit <https://example.com> for more"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
			}
		}
	}
	if !foundLink {
		t.Error("expected autolink to produce link mark")
	}
}

func TestMarkdownToADF_EmailAutoLink(t *testing.T) {
	md := "Contact <test@example.com> for help"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
			}
		}
	}
	if !foundLink {
		t.Error("expected email autolink to produce link mark")
	}
}

func TestMarkdownToADF_DeeplyNestedList(t *testing.T) {
	md := "- Level 1\n  - Level 2\n    - Level 3\n      - Level 4"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := adf.Content[0]
	if list.Type != NodeTypeBulletList {
		t.Fatalf("expected bulletList, got %q", list.Type)
	}

	// Traverse to find depth
	depth := 1
	current := list.Content[0] // First item
	for {
		foundNested := false
		for _, child := range current.Content {
			if child.Type == NodeTypeBulletList {
				depth++
				if len(child.Content) > 0 {
					current = child.Content[0]
					foundNested = true
					break
				}
			}
		}
		if !foundNested {
			break
		}
	}
	if depth < 4 {
		t.Errorf("expected at least 4 levels of nesting, got %d", depth)
	}
}

func TestMarkdownToADF_MixedListTypes(t *testing.T) {
	md := "1. Ordered\n   - Unordered inside\n   - Another\n2. Back to ordered"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := adf.Content[0]
	if list.Type != NodeTypeOrderedList {
		t.Fatalf("expected orderedList, got %q", list.Type)
	}

	// First item should contain bullet list
	firstItem := list.Content[0]
	foundBullet := false
	for _, child := range firstItem.Content {
		if child.Type == NodeTypeBulletList {
			foundBullet = true
		}
	}
	if !foundBullet {
		t.Error("expected bullet list nested in ordered list")
	}
}

func TestMarkdownToADF_ListWithMultipleParagraphs(t *testing.T) {
	md := "- First item\n\n  With continuation\n\n- Second item"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still produce valid list structure
	if len(adf.Content) == 0 {
		t.Fatal("expected content")
	}
}

func TestMarkdownToADF_TaskListAllUnchecked(t *testing.T) {
	md := "- [ ] Item 1\n- [ ] Item 2\n- [ ] Item 3"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := adf.Content[0]
	if list.Type != NodeTypeTaskList {
		t.Fatalf("expected taskList, got %q", list.Type)
	}
	for i, item := range list.Content {
		state := item.Attrs["state"].(string)
		if state != TaskStateTODO {
			t.Errorf("item %d: expected TODO, got %q", i, state)
		}
	}
}

func TestMarkdownToADF_TaskListAllChecked(t *testing.T) {
	md := "- [x] Item 1\n- [x] Item 2\n- [x] Item 3"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := adf.Content[0]
	if list.Type != NodeTypeTaskList {
		t.Fatalf("expected taskList, got %q", list.Type)
	}
	for i, item := range list.Content {
		state := item.Attrs["state"].(string)
		if state != TaskStateDONE {
			t.Errorf("item %d: expected DONE, got %q", i, state)
		}
	}
}

func TestMarkdownToADF_TaskListUppercaseX(t *testing.T) {
	md := "- [X] Done with uppercase X"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list := adf.Content[0]
	if list.Type != NodeTypeTaskList {
		t.Fatalf("expected taskList, got %q", list.Type)
	}
	state := list.Content[0].Attrs["state"].(string)
	if state != TaskStateDONE {
		t.Errorf("expected uppercase X to be DONE, got %q", state)
	}
}

func TestMarkdownToADF_TableSingleColumn(t *testing.T) {
	md := "| Header |\n|--------|\n| Cell   |"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	table := adf.Content[0]
	if table.Type != NodeTypeTable {
		t.Fatalf("expected table, got %q", table.Type)
	}
	headerRow := table.Content[0]
	if len(headerRow.Content) != 1 {
		t.Errorf("expected 1 column, got %d", len(headerRow.Content))
	}
}

func TestMarkdownToADF_TableManyColumns(t *testing.T) {
	md := "| A | B | C | D | E |\n|---|---|---|---|---|\n| 1 | 2 | 3 | 4 | 5 |"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	table := adf.Content[0]
	headerRow := table.Content[0]
	if len(headerRow.Content) != 5 {
		t.Errorf("expected 5 columns, got %d", len(headerRow.Content))
	}
}

func TestMarkdownToADF_TableWithFormattingInCells(t *testing.T) {
	md := "| **Bold** | *Italic* | `Code` |\n|----------|----------|--------|\n| Normal   | ~~Strike~~ | [Link](url) |"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	table := adf.Content[0]
	if table.Type != NodeTypeTable {
		t.Fatalf("expected table, got %q", table.Type)
	}

	// Check header row has formatting
	headerRow := table.Content[0]
	firstCell := headerRow.Content[0]
	foundBold := false
	if para, ok := findParagraphInCell(firstCell); ok {
		for _, node := range para.Content {
			for _, mark := range node.Marks {
				if mark.Type == MarkTypeStrong {
					foundBold = true
				}
			}
		}
	}
	if !foundBold {
		t.Error("expected bold in table cell")
	}
}

func findParagraphInCell(cell ADFNode) (ADFNode, bool) {
	for _, child := range cell.Content {
		if child.Type == NodeTypeParagraph {
			return child, true
		}
	}
	return ADFNode{}, false
}

func TestMarkdownToADF_TableWithEmptyCells(t *testing.T) {
	md := "| A | B |\n|---|---|\n|   |   |"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	table := adf.Content[0]
	if table.Type != NodeTypeTable {
		t.Fatalf("expected table, got %q", table.Type)
	}
	// Should handle empty cells gracefully
	if len(table.Content) != 2 {
		t.Errorf("expected 2 rows, got %d", len(table.Content))
	}
}

func TestMarkdownToADF_BlockquoteMultipleParagraphs(t *testing.T) {
	md := "> First paragraph\n>\n> Second paragraph"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quote := adf.Content[0]
	if quote.Type != NodeTypeBlockquote {
		t.Fatalf("expected blockquote, got %q", quote.Type)
	}
	paraCount := 0
	for _, child := range quote.Content {
		if child.Type == NodeTypeParagraph {
			paraCount++
		}
	}
	if paraCount < 2 {
		t.Errorf("expected at least 2 paragraphs in blockquote, got %d", paraCount)
	}
}

func TestMarkdownToADF_NestedBlockquotes(t *testing.T) {
	md := "> Outer quote\n>> Nested quote"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quote := adf.Content[0]
	if quote.Type != NodeTypeBlockquote {
		t.Fatalf("expected blockquote, got %q", quote.Type)
	}
	// Should contain nested blockquote
	foundNested := false
	for _, child := range quote.Content {
		if child.Type == NodeTypeBlockquote {
			foundNested = true
		}
	}
	if !foundNested {
		t.Error("expected nested blockquote")
	}
}

func TestMarkdownToADF_BlockquoteWithList(t *testing.T) {
	md := "> Quote with list:\n> - Item 1\n> - Item 2"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quote := adf.Content[0]
	if quote.Type != NodeTypeBlockquote {
		t.Fatalf("expected blockquote, got %q", quote.Type)
	}
	foundList := false
	for _, child := range quote.Content {
		if child.Type == NodeTypeBulletList {
			foundList = true
		}
	}
	if !foundList {
		t.Error("expected list inside blockquote")
	}
}

func TestMarkdownToADF_BlockquoteWithCode(t *testing.T) {
	md := "> Quote with `inline code`"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	quote := adf.Content[0]
	foundCode := false
	for _, child := range quote.Content {
		if child.Type == NodeTypeParagraph {
			for _, node := range child.Content {
				for _, mark := range node.Marks {
					if mark.Type == MarkTypeCode {
						foundCode = true
					}
				}
			}
		}
	}
	if !foundCode {
		t.Error("expected code inside blockquote")
	}
}

func TestMarkdownToADF_MultipleHorizontalRules(t *testing.T) {
	md := "Above\n\n---\n\nMiddle\n\n***\n\nBelow\n\n___"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ruleCount := 0
	for _, node := range adf.Content {
		if node.Type == NodeTypeRule {
			ruleCount++
		}
	}
	if ruleCount != 3 {
		t.Errorf("expected 3 horizontal rules, got %d", ruleCount)
	}
}

func TestMarkdownToADF_HorizontalRuleVariants(t *testing.T) {
	variants := []string{
		"---",
		"***",
		"___",
		"- - -",
		"* * *",
		"_ _ _",
	}
	for _, md := range variants {
		adf, err := MarkdownToADF(md)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", md, err)
		}
		if len(adf.Content) == 0 {
			t.Errorf("expected content for %q", md)
			continue
		}
		if adf.Content[0].Type != NodeTypeRule {
			t.Errorf("expected rule for %q, got %q", md, adf.Content[0].Type)
		}
	}
}

func TestMarkdownToADF_UnicodeText(t *testing.T) {
	md := "Hello 世界! Привет мир! مرحبا بالعالم"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	var text string
	for _, node := range para.Content {
		text += node.Text
	}
	if !strings.Contains(text, "世界") {
		t.Error("expected Chinese characters")
	}
	if !strings.Contains(text, "Привет") {
		t.Error("expected Russian characters")
	}
	if !strings.Contains(text, "مرحبا") {
		t.Error("expected Arabic characters")
	}
}

func TestMarkdownToADF_Emoji(t *testing.T) {
	md := "Hello 👋 World 🌍!"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	var text string
	for _, node := range para.Content {
		text += node.Text
	}
	if !strings.Contains(text, "👋") || !strings.Contains(text, "🌍") {
		t.Errorf("expected emoji to be preserved, got %q", text)
	}
}

func TestMarkdownToADF_SpecialCharacters(t *testing.T) {
	md := "Special: < > & \" ' \\ / @ # $ % ^ & * ( ) [ ] { }"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(adf.Content) == 0 {
		t.Fatal("expected content")
	}
	// Should handle without error
}

func TestMarkdownToADF_HardBreak(t *testing.T) {
	md := "Line one  \nLine two"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundHardBreak := false
	for _, node := range para.Content {
		if node.Type == NodeTypeHardBreak {
			foundHardBreak = true
		}
	}
	if !foundHardBreak {
		t.Error("expected hard break (two trailing spaces)")
	}
}

func TestMarkdownToADF_HardBreakWithBackslash(t *testing.T) {
	md := "Line one\\\nLine two"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should handle backslash line break
	if len(adf.Content) == 0 {
		t.Fatal("expected content")
	}
}

func TestMarkdownToADF_EscapedCharacters(t *testing.T) {
	// Test that escaped asterisks don't produce formatting
	md := "Not \\*italic\\* here"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	// Should NOT have italic marks for escaped asterisks
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeEm {
				t.Errorf("escaped asterisks should not produce italic, got text %q with marks", node.Text)
			}
		}
	}
}

func TestMarkdownToADF_MixedDocument(t *testing.T) {
	md := `# Main Heading

This is a paragraph with **bold**, *italic*, and ` + "`code`" + `.

## Sub Heading

> A blockquote with some wisdom

- List item 1
- List item 2
  - Nested item

| Header | Value |
|--------|-------|
| A      | 1     |

---

` + "```go" + `
func main() {}
` + "```" + `

Final paragraph.`

	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	types := make(map[string]int)
	for _, node := range adf.Content {
		types[node.Type]++
	}

	if types[NodeTypeHeading] < 2 {
		t.Error("expected at least 2 headings")
	}
	if types[NodeTypeParagraph] < 2 {
		t.Error("expected at least 2 paragraphs")
	}
	if types[NodeTypeBlockquote] < 1 {
		t.Error("expected at least 1 blockquote")
	}
	if types[NodeTypeBulletList] < 1 {
		t.Error("expected at least 1 bullet list")
	}
	if types[NodeTypeTable] < 1 {
		t.Error("expected at least 1 table")
	}
	if types[NodeTypeRule] < 1 {
		t.Error("expected at least 1 rule")
	}
	if types[NodeTypeCodeBlock] < 1 {
		t.Error("expected at least 1 code block")
	}
}

func TestMarkdownToADF_VeryLongLine(t *testing.T) {
	longWord := strings.Repeat("a", 10000)
	md := "Start " + longWord + " end"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	var text string
	for _, node := range para.Content {
		text += node.Text
	}
	if len(text) < 10000 {
		t.Error("expected long text to be preserved")
	}
}

func TestMarkdownToADF_ManyParagraphs(t *testing.T) {
	var paragraphs []string
	for i := 0; i < 100; i++ {
		paragraphs = append(paragraphs, "Paragraph "+string(rune('A'+i%26)))
	}
	md := strings.Join(paragraphs, "\n\n")
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(adf.Content) != 100 {
		t.Errorf("expected 100 paragraphs, got %d", len(adf.Content))
	}
}

func TestMarkdownToADF_StrikethroughWithOtherMarks(t *testing.T) {
	md := "~~**bold strikethrough**~~"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundStrike := false
	foundBold := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeStrike {
				foundStrike = true
			}
			if mark.Type == MarkTypeStrong {
				foundBold = true
			}
		}
	}
	if !foundStrike || !foundBold {
		t.Error("expected both strikethrough and bold")
	}
}

func TestMarkdownToADF_LinkWithBoldText(t *testing.T) {
	md := "[**Bold Link**](https://example.com)"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
			}
		}
	}
	if !foundLink {
		t.Error("expected link with bold text")
	}
}

// =============================================================================
// ADF to Markdown Edge Cases
// =============================================================================

func TestADFToMarkdown_EmptyDocument(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{},
	}
	md := ADFToMarkdownFromStruct(adf)
	if md != "" {
		t.Errorf("expected empty string, got %q", md)
	}
}

func TestADFToMarkdown_NilDocument(t *testing.T) {
	md := ADFToMarkdownFromStruct(nil)
	if md != "" {
		t.Errorf("expected empty string for nil, got %q", md)
	}
}

func TestADFToMarkdown_EmptyJSON(t *testing.T) {
	md := ADFToMarkdown([]byte(""))
	if md != "" {
		t.Errorf("expected empty string for empty JSON, got %q", md)
	}
}

func TestADFToMarkdown_InvalidJSON(t *testing.T) {
	md := ADFToMarkdown([]byte("not valid json"))
	if md != "" {
		t.Errorf("expected empty string for invalid JSON, got %q", md)
	}
}

func TestADFToMarkdown_EmptyParagraph(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: NodeTypeParagraph, Content: []ADFNode{}},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should handle gracefully (empty or minimal output)
	_ = md // No error expected
}

func TestADFToMarkdown_EmptyTextNode(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: ""},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	_ = md // Should not panic
}

func TestADFToMarkdown_UnknownNodeType(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: "unknownType", Content: []ADFNode{{Type: NodeTypeText, Text: "hidden"}}},
			{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "visible"}}},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "visible") {
		t.Error("expected visible content to be present")
	}
}

func TestADFToMarkdown_UnknownMarkType(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "text", Marks: []ADFMark{{Type: "unknownMark"}}},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "text") {
		t.Error("expected text to be present despite unknown mark")
	}
}

func TestADFToMarkdown_HeadingLevelZero(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeHeading,
				Attrs:   map[string]any{"level": 0},
				Content: []ADFNode{{Type: NodeTypeText, Text: "Test"}},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should clamp to level 1
	if !strings.HasPrefix(md, "# ") {
		t.Errorf("expected level 0 to become level 1, got %q", md)
	}
}

func TestADFToMarkdown_HeadingLevelSeven(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeHeading,
				Attrs:   map[string]any{"level": 7},
				Content: []ADFNode{{Type: NodeTypeText, Text: "Test"}},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should clamp to level 6
	if !strings.HasPrefix(md, "###### ") {
		t.Errorf("expected level 7 to become level 6, got %q", md)
	}
}

func TestADFToMarkdown_HeadingNegativeLevel(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeHeading,
				Attrs:   map[string]any{"level": -1},
				Content: []ADFNode{{Type: NodeTypeText, Text: "Test"}},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should clamp to level 1
	if !strings.HasPrefix(md, "# ") {
		t.Errorf("expected negative level to become level 1, got %q", md)
	}
}

func TestADFToMarkdown_HeadingMissingLevel(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeHeading,
				Attrs:   map[string]any{}, // No level
				Content: []ADFNode{{Type: NodeTypeText, Text: "Test"}},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should default to level 1
	if !strings.HasPrefix(md, "# ") {
		t.Errorf("expected default level 1, got %q", md)
	}
}

func TestADFToMarkdown_HeadingLevelAsFloat(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeHeading,
				Attrs:   map[string]any{"level": 2.0}, // Float from JSON
				Content: []ADFNode{{Type: NodeTypeText, Text: "Test"}},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.HasPrefix(md, "## ") {
		t.Errorf("expected float level to work, got %q", md)
	}
}

func TestADFToMarkdown_MultipleMarks(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{
						Type: NodeTypeText,
						Text: "formatted",
						Marks: []ADFMark{
							{Type: MarkTypeStrong},
							{Type: MarkTypeEm},
							{Type: MarkTypeStrike},
						},
					},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should have all marks applied
	if !strings.Contains(md, "**") {
		t.Error("expected bold")
	}
	if !strings.Contains(md, "*") {
		t.Error("expected italic")
	}
	if !strings.Contains(md, "~~") {
		t.Error("expected strikethrough")
	}
}

func TestADFToMarkdown_Strikethrough(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "deleted", Marks: []ADFMark{{Type: MarkTypeStrike}}},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "~~deleted~~") {
		t.Errorf("expected strikethrough, got %q", md)
	}
}

func TestADFToMarkdown_CodeBlockNoLanguage(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeCodeBlock,
				Content: []ADFNode{{Type: NodeTypeText, Text: "code here"}},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.HasPrefix(md, "```\n") {
		t.Errorf("expected code block without language, got %q", md)
	}
}

func TestADFToMarkdown_CodeBlockEmptyContent(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:    NodeTypeCodeBlock,
				Attrs:   map[string]any{"language": "go"},
				Content: []ADFNode{},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should produce valid but empty code block
	if !strings.Contains(md, "```go") {
		t.Errorf("expected code block header, got %q", md)
	}
}

func TestADFToMarkdown_LinkMissingHref(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{
						Type:  NodeTypeText,
						Text:  "link text",
						Marks: []ADFMark{{Type: MarkTypeLink, Attrs: map[string]any{}}},
					},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should produce link with empty href
	if !strings.Contains(md, "[link text]()") {
		t.Errorf("expected link with empty href, got %q", md)
	}
}

func TestADFToMarkdown_NestedListsInADF(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeBulletList,
				Content: []ADFNode{
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Level 1"}}},
							{
								Type: NodeTypeBulletList,
								Content: []ADFNode{
									{
										Type: NodeTypeListItem,
										Content: []ADFNode{
											{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Level 2"}}},
											{
												Type: NodeTypeBulletList,
												Content: []ADFNode{
													{
														Type: NodeTypeListItem,
														Content: []ADFNode{
															{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Level 3"}}},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "- Level 1") {
		t.Error("expected Level 1")
	}
	if !strings.Contains(md, "  - Level 2") {
		t.Error("expected indented Level 2")
	}
	if !strings.Contains(md, "    - Level 3") {
		t.Error("expected double indented Level 3")
	}
}

func TestADFToMarkdown_EmptyList(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: NodeTypeBulletList, Content: []ADFNode{}},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should handle empty list gracefully
	_ = md
}

func TestADFToMarkdown_EmptyTable(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: NodeTypeTable, Content: []ADFNode{}},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should produce empty output for empty table
	if strings.Contains(md, "|") {
		t.Error("expected no table output for empty table")
	}
}

func TestADFToMarkdown_TableWithPipeInContent(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeTable,
				Content: []ADFNode{
					{
						Type: NodeTypeTableRow,
						Content: []ADFNode{
							{
								Type: NodeTypeTableHeader,
								Content: []ADFNode{
									{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "A | B"}}},
								},
							},
						},
					},
					{
						Type: NodeTypeTableRow,
						Content: []ADFNode{
							{
								Type: NodeTypeTableCell,
								Content: []ADFNode{
									{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "C | D"}}},
								},
							},
						},
					},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Pipes in content should be escaped
	if !strings.Contains(md, "\\|") {
		t.Errorf("expected escaped pipes, got %q", md)
	}
}

func TestADFToMarkdown_HardBreak(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Line 1"},
					{Type: NodeTypeHardBreak},
					{Type: NodeTypeText, Text: "Line 2"},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Hard break should produce two trailing spaces
	if !strings.Contains(md, "  \n") {
		t.Errorf("expected hard break, got %q", md)
	}
}

func TestADFToMarkdown_UnicodeContent(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Hello 世界 👋 Привет"},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "世界") {
		t.Error("expected Chinese characters")
	}
	if !strings.Contains(md, "👋") {
		t.Error("expected emoji")
	}
	if !strings.Contains(md, "Привет") {
		t.Error("expected Russian characters")
	}
}

func TestADFToMarkdown_TaskItemMissingState(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:  NodeTypeTaskList,
				Attrs: map[string]any{"localId": "test"},
				Content: []ADFNode{
					{
						Type:  NodeTypeTaskItem,
						Attrs: map[string]any{"localId": "item"}, // Missing state
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Item"}}},
						},
					},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Should default to unchecked
	if !strings.Contains(md, "[ ]") {
		t.Errorf("expected unchecked checkbox for missing state, got %q", md)
	}
}

func TestADFToMarkdown_BlockquoteNestedContent(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeBlockquote,
				Content: []ADFNode{
					{
						Type: NodeTypeParagraph,
						Content: []ADFNode{
							{Type: NodeTypeText, Text: "Line 1"},
						},
					},
					{
						Type: NodeTypeParagraph,
						Content: []ADFNode{
							{Type: NodeTypeText, Text: "Line 2"},
						},
					},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	lines := strings.Split(md, "\n")
	quotedLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "> ") {
			quotedLines++
		}
	}
	if quotedLines < 2 {
		t.Errorf("expected at least 2 quoted lines, got %d", quotedLines)
	}
}

func TestADFToMarkdown_RealWorldJiraADF(t *testing.T) {
	// Simulates ADF that might come from Jira
	adfJSON := `{
		"version": 1,
		"type": "doc",
		"content": [
			{
				"type": "paragraph",
				"content": [
					{"type": "text", "text": "This is a "},
					{"type": "text", "text": "Jira", "marks": [{"type": "strong"}]},
					{"type": "text", "text": " issue description."}
				]
			},
			{
				"type": "heading",
				"attrs": {"level": 2},
				"content": [{"type": "text", "text": "Steps to Reproduce"}]
				},
			{
				"type": "orderedList",
				"content": [
					{
						"type": "listItem",
						"content": [
							{"type": "paragraph", "content": [{"type": "text", "text": "Open the application"}]}
						]
					},
					{
						"type": "listItem",
						"content": [
							{"type": "paragraph", "content": [{"type": "text", "text": "Click the button"}]}
						]
					}
				]
			},
			{
				"type": "codeBlock",
				"attrs": {"language": "bash"},
				"content": [{"type": "text", "text": "npm run test"}]
			}
		]
	}`

	md := ADFToMarkdown([]byte(adfJSON))
	if !strings.Contains(md, "**Jira**") {
		t.Error("expected bold Jira")
	}
	if !strings.Contains(md, "## Steps to Reproduce") {
		t.Error("expected heading")
	}
	if !strings.Contains(md, "1. Open") {
		t.Error("expected ordered list")
	}
	if !strings.Contains(md, "```bash") {
		t.Error("expected code block")
	}
}

func TestADFToMarkdown_SpecialCharsEscaping(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Text with *asterisks* and _underscores_ and [brackets]"},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// These should be escaped to prevent formatting
	if !strings.Contains(md, "\\*") {
		t.Error("expected escaped asterisks")
	}
	if !strings.Contains(md, "\\_") {
		t.Error("expected escaped underscores")
	}
	if !strings.Contains(md, "\\[") {
		t.Error("expected escaped brackets")
	}
}

func TestADFToMarkdown_InlineCodeNoEscaping(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "*test*", Marks: []ADFMark{{Type: MarkTypeCode}}},
				},
			},
		},
	}
	md := ADFToMarkdownFromStruct(adf)
	// Code content should not be escaped
	if strings.Contains(md, "\\*") {
		t.Error("code content should not be escaped")
	}
	if !strings.Contains(md, "`*test*`") {
		t.Errorf("expected inline code, got %q", md)
	}
}

// =============================================================================
// Round-trip tests
// =============================================================================

func TestRoundTrip_Paragraph(t *testing.T) {
	original := "Hello, world!"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should contain the original text (may have escapes)
	if !strings.Contains(md, "Hello") || !strings.Contains(md, "world") {
		t.Errorf("round-trip failed: got %q", md)
	}
}

func TestRoundTrip_Heading(t *testing.T) {
	original := "## My Heading"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.HasPrefix(md, "## ") {
		t.Errorf("round-trip failed: expected '## ' prefix, got %q", md)
	}
}

func TestRoundTrip_CodeBlock(t *testing.T) {
	original := "```go\nfmt.Println(\"test\")\n```"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "```go") {
		t.Errorf("round-trip failed: expected '```go', got %q", md)
	}
	if !strings.Contains(md, "fmt.Println") {
		t.Errorf("round-trip failed: expected code content, got %q", md)
	}
}

func TestRoundTrip_BulletList(t *testing.T) {
	original := "- Item 1\n- Item 2\n- Item 3"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "- Item 1") {
		t.Error("round-trip failed: missing Item 1")
	}
	if !strings.Contains(md, "- Item 2") {
		t.Error("round-trip failed: missing Item 2")
	}
	if !strings.Contains(md, "- Item 3") {
		t.Error("round-trip failed: missing Item 3")
	}
}

func TestRoundTrip_OrderedList(t *testing.T) {
	original := "1. First\n2. Second\n3. Third"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "1.") || !strings.Contains(md, "First") {
		t.Error("round-trip failed: missing First")
	}
	if !strings.Contains(md, "2.") || !strings.Contains(md, "Second") {
		t.Error("round-trip failed: missing Second")
	}
}

func TestRoundTrip_TaskList(t *testing.T) {
	original := "- [ ] Todo\n- [x] Done"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "[ ]") {
		t.Error("round-trip failed: missing unchecked")
	}
	if !strings.Contains(md, "[x]") {
		t.Error("round-trip failed: missing checked")
	}
}

func TestRoundTrip_Table(t *testing.T) {
	original := "| A | B |\n|---|---|\n| 1 | 2 |"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "| A") {
		t.Error("round-trip failed: missing header A")
	}
	if !strings.Contains(md, "| B") {
		t.Error("round-trip failed: missing header B")
	}
	if !strings.Contains(md, "---") {
		t.Error("round-trip failed: missing separator")
	}
}

func TestRoundTrip_Blockquote(t *testing.T) {
	original := "> This is a quote"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.HasPrefix(md, ">") {
		t.Errorf("round-trip failed: expected blockquote, got %q", md)
	}
	if !strings.Contains(md, "quote") {
		t.Error("round-trip failed: missing quote content")
	}
}

func TestRoundTrip_Bold(t *testing.T) {
	original := "This is **bold** text"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "**bold**") {
		t.Errorf("round-trip failed: expected bold, got %q", md)
	}
}

func TestRoundTrip_Italic(t *testing.T) {
	original := "This is *italic* text"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "*italic*") {
		t.Errorf("round-trip failed: expected italic, got %q", md)
	}
}

func TestRoundTrip_Strikethrough(t *testing.T) {
	original := "This is ~~deleted~~ text"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "~~deleted~~") {
		t.Errorf("round-trip failed: expected strikethrough, got %q", md)
	}
}

func TestRoundTrip_InlineCode(t *testing.T) {
	original := "Use `code` here"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "`code`") {
		t.Errorf("round-trip failed: expected inline code, got %q", md)
	}
}

func TestRoundTrip_Link(t *testing.T) {
	original := "Visit [Google](https://google.com) now"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "[Google]") {
		t.Error("round-trip failed: missing link text")
	}
	if !strings.Contains(md, "(https://google.com)") {
		t.Error("round-trip failed: missing URL")
	}
}

func TestRoundTrip_HorizontalRule(t *testing.T) {
	original := "Above\n\n---\n\nBelow"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "---") {
		t.Errorf("round-trip failed: expected horizontal rule, got %q", md)
	}
}

func TestRoundTrip_NestedList(t *testing.T) {
	original := "- Level 1\n  - Level 2\n    - Level 3"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "Level 1") {
		t.Error("round-trip failed: missing Level 1")
	}
	if !strings.Contains(md, "Level 2") {
		t.Error("round-trip failed: missing Level 2")
	}
	if !strings.Contains(md, "Level 3") {
		t.Error("round-trip failed: missing Level 3")
	}
}

func TestRoundTrip_ComplexDocument(t *testing.T) {
	original := `# Title

This has **bold** and *italic*.

## Section

- Item 1
- Item 2

> A quote

` + "```go" + `
code
` + "```"

	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)

	// Verify key elements survived round-trip
	if !strings.Contains(md, "# Title") {
		t.Error("round-trip failed: missing h1")
	}
	if !strings.Contains(md, "**bold**") {
		t.Error("round-trip failed: missing bold")
	}
	if !strings.Contains(md, "*italic*") {
		t.Error("round-trip failed: missing italic")
	}
	if !strings.Contains(md, "## Section") {
		t.Error("round-trip failed: missing h2")
	}
	if !strings.Contains(md, "- Item") {
		t.Error("round-trip failed: missing list")
	}
	if !strings.Contains(md, ">") {
		t.Error("round-trip failed: missing quote")
	}
	if !strings.Contains(md, "```go") {
		t.Error("round-trip failed: missing code block")
	}
}

func TestRoundTrip_Unicode(t *testing.T) {
	original := "Hello 世界 👋"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "世界") {
		t.Error("round-trip failed: missing Chinese")
	}
	if !strings.Contains(md, "👋") {
		t.Error("round-trip failed: missing emoji")
	}
}

func TestRoundTrip_MultipleFormats(t *testing.T) {
	original := "***bold and italic***"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should have both bold and italic markers
	if !strings.Contains(md, "**") {
		t.Error("round-trip failed: missing bold")
	}
	if !strings.Contains(md, "*") {
		t.Error("round-trip failed: missing italic")
	}
}

// =============================================================================
// Expected Normalization Tests (verifying known differences are handled)
// =============================================================================

func TestNormalization_AlternativeBoldSyntax(t *testing.T) {
	// __bold__ should normalize to **bold** after round-trip
	original := "This is __bold__ text"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should contain ** style bold (normalized from __)
	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected __bold__ to normalize to **bold**, got %q", md)
	}
}

func TestNormalization_AlternativeItalicSyntax(t *testing.T) {
	// _italic_ should normalize to *italic* after round-trip
	original := "This is _italic_ text"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should contain * style italic (normalized from _)
	if !strings.Contains(md, "*italic*") {
		t.Errorf("expected _italic_ to normalize to *italic*, got %q", md)
	}
}

func TestNormalization_HorizontalRuleVariants(t *testing.T) {
	// All HR variants should normalize to ---
	variants := []string{"***", "___", "- - -", "* * *", "_ _ _"}
	for _, variant := range variants {
		adf, err := MarkdownToADF(variant)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", variant, err)
		}

		md := ADFToMarkdownFromStruct(adf)
		if !strings.Contains(md, "---") {
			t.Errorf("expected %q to normalize to ---, got %q", variant, md)
		}
	}
}

func TestNormalization_SoftLineBreakBecomesSpace(t *testing.T) {
	// Soft line break (single newline in paragraph) should become space
	original := "Line one\nLine two"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be single paragraph
	if len(adf.Content) != 1 {
		t.Errorf("expected 1 paragraph, got %d", len(adf.Content))
	}

	// Text should be joined with space, not newline
	md := ADFToMarkdownFromStruct(adf)
	if strings.Contains(md, "  \n") {
		t.Error("soft line break should not become hard break")
	}
}

func TestNormalization_WhitespaceInText(t *testing.T) {
	// Multiple spaces may be normalized
	original := "Word   with   spaces"
	adf, err := MarkdownToADF(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should contain the words, spaces may be normalized
	if !strings.Contains(md, "Word") || !strings.Contains(md, "spaces") {
		t.Errorf("expected words to be preserved, got %q", md)
	}
}

func TestNormalization_TaskListLocalIdRegenerated(t *testing.T) {
	// Task list localId is regenerated each time, so two conversions differ
	original := "- [ ] Task"
	adf1, _ := MarkdownToADF(original)
	adf2, _ := MarkdownToADF(original)

	// Both should have localId but they should be different UUIDs
	id1 := adf1.Content[0].Attrs["localId"].(string)
	id2 := adf2.Content[0].Attrs["localId"].(string)

	if id1 == id2 {
		t.Error("expected different localIds for each conversion")
	}

	// But semantic content should be the same
	md1 := ADFToMarkdownFromStruct(adf1)
	md2 := ADFToMarkdownFromStruct(adf2)
	if md1 != md2 {
		t.Errorf("expected same markdown output, got %q and %q", md1, md2)
	}
}

// =============================================================================
// Unsupported ADF Elements (should skip gracefully)
// =============================================================================

func TestADFToMarkdown_MediaNodeSkipped(t *testing.T) {
	// Media/image nodes should be skipped (out of scope)
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: "mediaSingle",
				Content: []ADFNode{
					{
						Type: "media",
						Attrs: map[string]any{
							"type": "file",
							"id":   "abc-123",
						},
					},
				},
			},
			{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "After image"}}},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should skip media, but include paragraph
	if !strings.Contains(md, "After image") {
		t.Error("expected paragraph after skipped media")
	}
}

func TestADFToMarkdown_MentionNodeSkipped(t *testing.T) {
	// Mention nodes should be skipped
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Hello "},
					{
						Type: "mention",
						Attrs: map[string]any{
							"id":   "user-123",
							"text": "@john",
						},
					},
					{Type: NodeTypeText, Text: " there"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	// Should have Hello and there, mention skipped
	if !strings.Contains(md, "Hello") || !strings.Contains(md, "there") {
		t.Errorf("expected surrounding text, got %q", md)
	}
}

func TestADFToMarkdown_PanelNodeSkipped(t *testing.T) {
	// Panel nodes should be skipped
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:  "panel",
				Attrs: map[string]any{"panelType": "info"},
				Content: []ADFNode{
					{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Panel content"}}},
				},
			},
			{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "After panel"}}},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	// Panel skipped, but after paragraph should exist
	if !strings.Contains(md, "After panel") {
		t.Error("expected paragraph after skipped panel")
	}
}

func TestADFToMarkdown_ExpandNodeSkipped(t *testing.T) {
	// Expand nodes should be skipped
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:  "expand",
				Attrs: map[string]any{"title": "Click to expand"},
				Content: []ADFNode{
					{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Hidden content"}}},
				},
			},
			{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Visible content"}}},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "Visible content") {
		t.Error("expected visible content after skipped expand")
	}
}

func TestADFToMarkdown_LayoutNodeSkipped(t *testing.T) {
	// Layout nodes should be skipped
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: "layoutSection",
				Content: []ADFNode{
					{
						Type: "layoutColumn",
						Attrs: map[string]any{"width": 50},
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Column 1"}}},
						},
					},
				},
			},
			{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "After layout"}}},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "After layout") {
		t.Error("expected content after skipped layout")
	}
}

func TestADFToMarkdown_InlineCardSkipped(t *testing.T) {
	// Jira issue link inline cards should be skipped
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "See issue "},
					{
						Type: "inlineCard",
						Attrs: map[string]any{
							"url": "https://jira.example.com/browse/ABC-123",
						},
					},
					{Type: NodeTypeText, Text: " for details"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "See issue") || !strings.Contains(md, "for details") {
		t.Errorf("expected surrounding text, got %q", md)
	}
}

func TestADFToMarkdown_EmojiNodeSkipped(t *testing.T) {
	// Jira emoji nodes (not unicode) should be skipped
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Great job "},
					{
						Type: "emoji",
						Attrs: map[string]any{
							"shortName": ":thumbsup:",
							"text":      "👍",
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "Great job") {
		t.Errorf("expected text before skipped emoji, got %q", md)
	}
}

// =============================================================================
// Additional Markdown Edge Cases
// =============================================================================

func TestMarkdownToADF_SetextHeading(t *testing.T) {
	// Setext-style headings (underline style)
	md := "Heading 1\n=========\n\nHeading 2\n---------"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	headingCount := 0
	for _, node := range adf.Content {
		if node.Type == NodeTypeHeading {
			headingCount++
		}
	}
	if headingCount != 2 {
		t.Errorf("expected 2 setext headings, got %d", headingCount)
	}
}

func TestMarkdownToADF_EmptyHeading(t *testing.T) {
	md := "## "
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should handle empty heading gracefully
	if len(adf.Content) > 0 {
		if adf.Content[0].Type == NodeTypeHeading {
			// Empty heading is valid
		}
	}
}

func TestMarkdownToADF_TableWithoutHeader(t *testing.T) {
	// GFM requires header row, but test that we handle edge case
	md := "| A | B |\n| C | D |"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not crash, may produce paragraph instead of table
	if len(adf.Content) == 0 {
		t.Error("expected some content")
	}
}

func TestMarkdownToADF_TableOnlyHeader(t *testing.T) {
	md := "| Header 1 | Header 2 |\n|----------|----------|"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) == 0 {
		t.Fatal("expected content")
	}

	table := adf.Content[0]
	if table.Type != NodeTypeTable {
		t.Fatalf("expected table, got %q", table.Type)
	}
	// Should have just header row
	if len(table.Content) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(table.Content))
	}
}

func TestMarkdownToADF_OrderedListStartNumber(t *testing.T) {
	// Ordered list starting from non-1 number
	md := "5. Fifth\n6. Sixth\n7. Seventh"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	list := adf.Content[0]
	if list.Type != NodeTypeOrderedList {
		t.Fatalf("expected orderedList, got %q", list.Type)
	}
	// ADF doesn't support start number, but content should be preserved
	if len(list.Content) != 3 {
		t.Errorf("expected 3 items, got %d", len(list.Content))
	}
}

func TestMarkdownToADF_ConsecutiveLists(t *testing.T) {
	// Two lists back to back
	md := "- Bullet 1\n- Bullet 2\n\n1. Ordered 1\n2. Ordered 2"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 2 {
		t.Fatalf("expected 2 lists, got %d", len(adf.Content))
	}

	if adf.Content[0].Type != NodeTypeBulletList {
		t.Errorf("expected first list to be bullet, got %q", adf.Content[0].Type)
	}
	if adf.Content[1].Type != NodeTypeOrderedList {
		t.Errorf("expected second list to be ordered, got %q", adf.Content[1].Type)
	}
}

func TestMarkdownToADF_ListItemWithMultipleBlocks(t *testing.T) {
	// List item with paragraph and code block
	md := "- Item with code:\n\n  ```go\n  code\n  ```\n\n- Next item"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should parse without error
	if len(adf.Content) == 0 {
		t.Error("expected content")
	}
}

func TestMarkdownToADF_InlineHTMLBreak(t *testing.T) {
	// <br> tag for line break
	md := "Line one<br>Line two"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should parse without error (HTML extracted or skipped per DR-007)
	if len(adf.Content) == 0 {
		t.Error("expected content")
	}
}

func TestMarkdownToADF_InlineHTMLSpan(t *testing.T) {
	// Inline HTML tags - text should be extracted
	md := "Text with <span style=\"color:red\">colored</span> word"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	var text string
	for _, node := range para.Content {
		text += node.Text
	}
	// Should contain the text, tags discarded
	if !strings.Contains(text, "colored") {
		t.Errorf("expected 'colored' text extracted from HTML, got %q", text)
	}
}

func TestMarkdownToADF_MultipleConsecutiveHardBreaks(t *testing.T) {
	md := "Line one  \n  \n  \nLine two"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should handle multiple hard breaks
	if len(adf.Content) == 0 {
		t.Error("expected content")
	}
}

func TestMarkdownToADF_TextOnlyWhitespace(t *testing.T) {
	md := "Before\n\n   \n\nAfter"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Whitespace-only "paragraph" should be skipped
	for _, node := range adf.Content {
		if node.Type == NodeTypeParagraph {
			var text string
			for _, child := range node.Content {
				text += child.Text
			}
			text = strings.TrimSpace(text)
			if text == "" {
				t.Error("empty paragraph should be skipped")
			}
		}
	}
}

func TestMarkdownToADF_LinkWithEmptyText(t *testing.T) {
	md := "[](https://example.com)"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should handle gracefully - link with URL as text
	if len(adf.Content) == 0 {
		t.Error("expected content")
	}
}

func TestMarkdownToADF_LinkWithTitle(t *testing.T) {
	md := "[Link](https://example.com \"Title text\")"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
				href := mark.Attrs["href"].(string)
				if href != "https://example.com" {
					t.Errorf("expected URL, got %q", href)
				}
			}
		}
	}
	if !foundLink {
		t.Error("expected link")
	}
}

func TestMarkdownToADF_ReferenceLink(t *testing.T) {
	md := "[Link text][ref]\n\n[ref]: https://example.com"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should resolve reference link
	para := adf.Content[0]
	foundLink := false
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				foundLink = true
			}
		}
	}
	if !foundLink {
		t.Error("expected reference link to be resolved")
	}
}

func TestMarkdownToADF_FootnoteStyleLink(t *testing.T) {
	md := "Check [this][1] and [that][2].\n\n[1]: https://one.com\n[2]: https://two.com"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	linkCount := 0
	for _, node := range para.Content {
		for _, mark := range node.Marks {
			if mark.Type == MarkTypeLink {
				linkCount++
			}
		}
	}
	if linkCount != 2 {
		t.Errorf("expected 2 reference links, got %d", linkCount)
	}
}

func TestMarkdownToADF_ImageAltTextPreserved(t *testing.T) {
	// Images are out of scope, but alt text should be preserved
	md := "![Alt text](https://example.com/image.png)"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Alt text should be in output as regular text
	para := adf.Content[0]
	var text string
	for _, node := range para.Content {
		text += node.Text
	}
	if !strings.Contains(text, "Alt text") {
		t.Errorf("expected alt text to be preserved, got %q", text)
	}
}

func TestMarkdownToADF_DefinitionList(t *testing.T) {
	// Definition lists aren't standard Markdown, should parse as regular text
	md := "Term\n: Definition"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not crash
	if len(adf.Content) == 0 {
		t.Error("expected content")
	}
}

// =============================================================================
// Additional ADF Edge Cases
// =============================================================================

func TestADFToMarkdown_NestedBlockquotes(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeBlockquote,
				Content: []ADFNode{
					{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Outer"}}},
					{
						Type: NodeTypeBlockquote,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Inner"}}},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "> Outer") {
		t.Error("expected outer quote")
	}
	if !strings.Contains(md, "> > Inner") {
		t.Errorf("expected nested quote with >> prefix, got %q", md)
	}
}

func TestADFToMarkdown_OrderedListManyItems(t *testing.T) {
	// Ordered list with 10+ items to verify numbering
	var items []ADFNode
	for i := 1; i <= 15; i++ {
		items = append(items, ADFNode{
			Type: NodeTypeListItem,
			Content: []ADFNode{
				{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Item"}}},
			},
		})
	}

	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: NodeTypeOrderedList, Content: items},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "10.") {
		t.Error("expected 10. for 10th item")
	}
	if !strings.Contains(md, "15.") {
		t.Error("expected 15. for 15th item")
	}
}

func TestADFToMarkdown_MixedNestedLists(t *testing.T) {
	// Ordered list containing bullet list
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeOrderedList,
				Content: []ADFNode{
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "First"}}},
							{
								Type: NodeTypeBulletList,
								Content: []ADFNode{
									{
										Type: NodeTypeListItem,
										Content: []ADFNode{
											{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Bullet A"}}},
										},
									},
								},
							},
						},
					},
					{
						Type: NodeTypeListItem,
						Content: []ADFNode{
							{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Second"}}},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "1. First") {
		t.Error("expected ordered list item")
	}
	if !strings.Contains(md, "  - Bullet A") {
		t.Errorf("expected nested bullet list, got %q", md)
	}
	if !strings.Contains(md, "2. Second") {
		t.Error("expected second ordered item")
	}
}

func TestADFToMarkdown_TableWithManyRows(t *testing.T) {
	var rows []ADFNode
	// Header row
	rows = append(rows, ADFNode{
		Type: NodeTypeTableRow,
		Content: []ADFNode{
			{Type: NodeTypeTableHeader, Content: []ADFNode{{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "H1"}}}}},
			{Type: NodeTypeTableHeader, Content: []ADFNode{{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "H2"}}}}},
		},
	})
	// 10 data rows
	for i := 1; i <= 10; i++ {
		rows = append(rows, ADFNode{
			Type: NodeTypeTableRow,
			Content: []ADFNode{
				{Type: NodeTypeTableCell, Content: []ADFNode{{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "A"}}}}},
				{Type: NodeTypeTableCell, Content: []ADFNode{{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "B"}}}}},
			},
		})
	}

	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{Type: NodeTypeTable, Content: rows},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	lines := strings.Split(md, "\n")
	// Should have header + separator + 10 data rows = 12 lines
	tableLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "|") {
			tableLines++
		}
	}
	if tableLines != 12 {
		t.Errorf("expected 12 table lines, got %d", tableLines)
	}
}

func TestADFToMarkdown_ComplexNestedStructure(t *testing.T) {
	// Blockquote containing list containing formatted text
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeBlockquote,
				Content: []ADFNode{
					{Type: NodeTypeParagraph, Content: []ADFNode{{Type: NodeTypeText, Text: "Quote intro:"}}},
					{
						Type: NodeTypeBulletList,
						Content: []ADFNode{
							{
								Type: NodeTypeListItem,
								Content: []ADFNode{
									{
										Type: NodeTypeParagraph,
										Content: []ADFNode{
											{Type: NodeTypeText, Text: "Bold", Marks: []ADFMark{{Type: MarkTypeStrong}}},
											{Type: NodeTypeText, Text: " item"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "> Quote intro") {
		t.Error("expected blockquote text")
	}
	if !strings.Contains(md, "**Bold**") {
		t.Error("expected bold in list inside blockquote")
	}
}

func TestADFToMarkdown_CodeBlockWithBackticks(t *testing.T) {
	// Code block containing backticks
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type:  NodeTypeCodeBlock,
				Attrs: map[string]any{"language": "md"},
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Use `code` in markdown"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	if !strings.Contains(md, "```md") {
		t.Error("expected code block with language")
	}
	if !strings.Contains(md, "`code`") {
		t.Error("expected backticks preserved in code block")
	}
}

func TestADFToMarkdown_TextWithBackslashes(t *testing.T) {
	adf := &ADF{
		Version: 1,
		Type:    NodeTypeDoc,
		Content: []ADFNode{
			{
				Type: NodeTypeParagraph,
				Content: []ADFNode{
					{Type: NodeTypeText, Text: "Path: C:\\Users\\name"},
				},
			},
		},
	}

	md := ADFToMarkdownFromStruct(adf)
	// Backslashes should NOT be escaped - they only have special meaning
	// before certain characters in markdown, and escaping them causes
	// double-escaping on round-trip
	if md != "Path: C:\\Users\\name" {
		t.Errorf("expected backslashes preserved, got %q", md)
	}
}

// Test ADF mark compatibility - code mark can only combine with link
// Per ADF spec: https://developer.atlassian.com/cloud/jira/platform/apis/document/marks/code/

func TestMarkdownToADF_CodeInBold_DropsIncompatibleMark(t *testing.T) {
	// ADF does not allow code + strong together, code takes precedence
	md := "**`code`**"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(adf.Content) != 1 {
		t.Fatalf("expected 1 content node, got %d", len(adf.Content))
	}

	para := adf.Content[0]
	if len(para.Content) != 1 {
		t.Fatalf("expected 1 text node, got %d", len(para.Content))
	}

	text := para.Content[0]
	if text.Text != "code" {
		t.Errorf("expected text 'code', got %q", text.Text)
	}

	// Should have only code mark, not strong (ADF incompatibility)
	if len(text.Marks) != 1 {
		t.Errorf("expected 1 mark, got %d", len(text.Marks))
	}
	if text.Marks[0].Type != MarkTypeCode {
		t.Errorf("expected code mark, got %q", text.Marks[0].Type)
	}
}

func TestMarkdownToADF_CodeInItalic_DropsIncompatibleMark(t *testing.T) {
	// ADF does not allow code + em together, code takes precedence
	md := "*`code`*"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	text := para.Content[0]

	// Should have only code mark, not em
	if len(text.Marks) != 1 {
		t.Errorf("expected 1 mark, got %d", len(text.Marks))
	}
	if text.Marks[0].Type != MarkTypeCode {
		t.Errorf("expected code mark, got %q", text.Marks[0].Type)
	}
}

func TestMarkdownToADF_CodeInStrikethrough_DropsIncompatibleMark(t *testing.T) {
	// ADF does not allow code + strike together, code takes precedence
	md := "~~`code`~~"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	text := para.Content[0]

	// Should have only code mark, not strike
	if len(text.Marks) != 1 {
		t.Errorf("expected 1 mark, got %d", len(text.Marks))
	}
	if text.Marks[0].Type != MarkTypeCode {
		t.Errorf("expected code mark, got %q", text.Marks[0].Type)
	}
}

func TestMarkdownToADF_BoldWithCode_PreservesNonCodeMarks(t *testing.T) {
	// Bold text around code should work, with separate nodes
	md := "**bold `code` more bold**"
	adf, err := MarkdownToADF(md)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	para := adf.Content[0]
	if len(para.Content) != 3 {
		t.Fatalf("expected 3 nodes (bold, code, bold), got %d", len(para.Content))
	}

	// First node: bold text
	if para.Content[0].Text != "bold " {
		t.Errorf("expected 'bold ', got %q", para.Content[0].Text)
	}
	if len(para.Content[0].Marks) != 1 || para.Content[0].Marks[0].Type != MarkTypeStrong {
		t.Errorf("expected strong mark on first node")
	}

	// Second node: code only (no strong)
	if para.Content[1].Text != "code" {
		t.Errorf("expected 'code', got %q", para.Content[1].Text)
	}
	if len(para.Content[1].Marks) != 1 || para.Content[1].Marks[0].Type != MarkTypeCode {
		t.Errorf("expected only code mark on second node, got %v", para.Content[1].Marks)
	}

	// Third node: bold text
	if para.Content[2].Text != " more bold" {
		t.Errorf("expected ' more bold', got %q", para.Content[2].Text)
	}
	if len(para.Content[2].Marks) != 1 || para.Content[2].Marks[0].Type != MarkTypeStrong {
		t.Errorf("expected strong mark on third node")
	}
}
