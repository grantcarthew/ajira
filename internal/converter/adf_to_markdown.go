package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ADFToMarkdown converts Atlassian Document Format to Markdown.
// It uses best-effort conversion, skipping unsupported node types.
func ADFToMarkdown(adfJSON []byte) string {
	var doc ADF
	if err := json.Unmarshal(adfJSON, &doc); err != nil {
		return ""
	}

	return renderNodes(doc.Content, 0)
}

// ADFToMarkdownFromStruct converts an ADF struct to Markdown.
func ADFToMarkdownFromStruct(doc *ADF) string {
	if doc == nil {
		return ""
	}
	return renderNodes(doc.Content, 0)
}

// renderNodes renders a slice of ADF nodes to Markdown.
func renderNodes(nodes []ADFNode, depth int) string {
	var parts []string
	for i, node := range nodes {
		rendered := renderNode(node, depth)
		if rendered != "" {
			parts = append(parts, rendered)
		}
		// Add blank line between block elements
		if i < len(nodes)-1 && isBlockElement(node) && isBlockElement(nodes[i+1]) {
			// Always add blank line between different block types
			// This includes: heading->list, list->heading, paragraph->list, etc.
			parts = append(parts, "")
		}
	}
	return strings.Join(parts, "\n")
}

// isBlockElement returns true if the node is a block-level element.
func isBlockElement(node ADFNode) bool {
	switch node.Type {
	case NodeTypeParagraph, NodeTypeHeading, NodeTypeCodeBlock,
		NodeTypeBlockquote, NodeTypeBulletList, NodeTypeOrderedList,
		NodeTypeTaskList, NodeTypeTable, NodeTypeRule:
		return true
	default:
		return false
	}
}

// isListItem returns true if the node is a list or list item.
func isListItem(node ADFNode) bool {
	switch node.Type {
	case NodeTypeBulletList, NodeTypeOrderedList, NodeTypeTaskList,
		NodeTypeListItem, NodeTypeTaskItem:
		return true
	default:
		return false
	}
}

// renderNode renders a single ADF node to Markdown.
func renderNode(node ADFNode, depth int) string {
	switch node.Type {
	case NodeTypeParagraph:
		return renderParagraph(node)
	case NodeTypeHeading:
		return renderHeading(node)
	case NodeTypeCodeBlock:
		return renderCodeBlock(node)
	case NodeTypeBlockquote:
		return renderBlockquote(node, depth)
	case NodeTypeBulletList:
		return renderBulletList(node, depth)
	case NodeTypeOrderedList:
		return renderOrderedList(node, depth)
	case NodeTypeTaskList:
		return renderTaskList(node, depth)
	case NodeTypeTable:
		return renderTable(node)
	case NodeTypeRule:
		return "---"
	case NodeTypeText:
		return renderText(node)
	case NodeTypeHardBreak:
		return "  \n"
	default:
		// Skip unsupported node types gracefully
		return ""
	}
}

func renderParagraph(node ADFNode) string {
	return renderInlineContent(node.Content)
}

func renderHeading(node ADFNode) string {
	level := 1
	if l, ok := node.Attrs["level"].(float64); ok {
		level = int(l)
	} else if l, ok := node.Attrs["level"].(int); ok {
		level = l
	}
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}

	prefix := strings.Repeat("#", level) + " "
	return prefix + renderInlineContent(node.Content)
}

func renderCodeBlock(node ADFNode) string {
	lang := ""
	if l, ok := node.Attrs["language"].(string); ok {
		lang = l
	}

	var code string
	for _, child := range node.Content {
		if child.Type == NodeTypeText {
			code += child.Text
		}
	}

	return fmt.Sprintf("```%s\n%s\n```", lang, code)
}

func renderBlockquote(node ADFNode, depth int) string {
	inner := renderNodes(node.Content, depth)
	lines := strings.Split(inner, "\n")
	var quoted []string
	for _, line := range lines {
		quoted = append(quoted, "> "+line)
	}
	return strings.Join(quoted, "\n")
}

func renderBulletList(node ADFNode, depth int) string {
	var items []string
	indent := strings.Repeat("  ", depth)

	for _, item := range node.Content {
		if item.Type == NodeTypeListItem {
			content := renderListItemContent(item, depth)
			items = append(items, indent+"- "+content)
		}
	}

	return strings.Join(items, "\n")
}

func renderOrderedList(node ADFNode, depth int) string {
	var items []string
	indent := strings.Repeat("  ", depth)

	for i, item := range node.Content {
		if item.Type == NodeTypeListItem {
			content := renderListItemContent(item, depth)
			items = append(items, fmt.Sprintf("%s%d. %s", indent, i+1, content))
		}
	}

	return strings.Join(items, "\n")
}

func renderTaskList(node ADFNode, depth int) string {
	var items []string
	indent := strings.Repeat("  ", depth)

	for _, item := range node.Content {
		if item.Type == NodeTypeTaskItem {
			checkbox := "[ ]"
			if state, ok := item.Attrs["state"].(string); ok && state == TaskStateDONE {
				checkbox = "[x]"
			}
			content := renderListItemContent(item, depth)
			items = append(items, fmt.Sprintf("%s- %s %s", indent, checkbox, content))
		}
	}

	return strings.Join(items, "\n")
}

func renderListItemContent(item ADFNode, depth int) string {
	var parts []string

	for i, child := range item.Content {
		switch child.Type {
		case NodeTypeParagraph:
			// First paragraph is inline with list marker
			if i == 0 {
				parts = append(parts, renderInlineContent(child.Content))
			} else {
				// Subsequent paragraphs need indent
				indent := strings.Repeat("  ", depth+1)
				parts = append(parts, "\n"+indent+renderInlineContent(child.Content))
			}
		case NodeTypeBulletList:
			parts = append(parts, "\n"+renderBulletList(child, depth+1))
		case NodeTypeOrderedList:
			parts = append(parts, "\n"+renderOrderedList(child, depth+1))
		case NodeTypeTaskList:
			parts = append(parts, "\n"+renderTaskList(child, depth+1))
		default:
			rendered := renderNode(child, depth+1)
			if rendered != "" {
				parts = append(parts, rendered)
			}
		}
	}

	return strings.Join(parts, "")
}

func renderTable(node ADFNode) string {
	var rows [][]string
	var isHeader []bool

	for _, row := range node.Content {
		if row.Type != NodeTypeTableRow {
			continue
		}

		var cells []string
		headerRow := false

		for _, cell := range row.Content {
			if cell.Type == NodeTypeTableHeader {
				headerRow = true
			}
			cellContent := ""
			for _, child := range cell.Content {
				if child.Type == NodeTypeParagraph {
					cellContent += renderInlineContent(child.Content)
				}
			}
			// Escape pipes in cell content
			cellContent = strings.ReplaceAll(cellContent, "|", "\\|")
			cells = append(cells, cellContent)
		}

		rows = append(rows, cells)
		isHeader = append(isHeader, headerRow)
	}

	if len(rows) == 0 {
		return ""
	}

	var lines []string

	// Render first row (header)
	if len(rows) > 0 {
		lines = append(lines, "| "+strings.Join(rows[0], " | ")+" |")

		// Add separator after header
		if len(isHeader) > 0 && isHeader[0] {
			var sep []string
			for range rows[0] {
				sep = append(sep, "---")
			}
			lines = append(lines, "| "+strings.Join(sep, " | ")+" |")
		}
	}

	// Render remaining rows
	for i := 1; i < len(rows); i++ {
		lines = append(lines, "| "+strings.Join(rows[i], " | ")+" |")
	}

	return strings.Join(lines, "\n")
}

// renderInlineContent renders inline content nodes to Markdown.
func renderInlineContent(nodes []ADFNode) string {
	// Merge adjacent text nodes with identical marks before rendering
	// This prevents over-escaping of underscores split across nodes by goldmark
	merged := mergeAdjacentTextNodes(nodes)

	var parts []string
	for _, node := range merged {
		parts = append(parts, renderInlineNode(node))
	}
	return strings.Join(parts, "")
}

// mergeAdjacentTextNodes merges consecutive text nodes that have identical marks.
func mergeAdjacentTextNodes(nodes []ADFNode) []ADFNode {
	if len(nodes) == 0 {
		return nodes
	}

	var result []ADFNode
	for _, node := range nodes {
		if node.Type != NodeTypeText {
			result = append(result, node)
			continue
		}

		// Check if we can merge with the previous node
		if len(result) > 0 {
			prev := &result[len(result)-1]
			if prev.Type == NodeTypeText && marksEqual(prev.Marks, node.Marks) {
				// Merge: append text to previous node
				prev.Text += node.Text
				continue
			}
		}

		result = append(result, node)
	}
	return result
}

// marksEqual returns true if two mark slices are equivalent.
func marksEqual(a, b []ADFMark) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type {
			return false
		}
	}
	return true
}

// renderInlineNode renders a single inline node to Markdown.
func renderInlineNode(node ADFNode) string {
	switch node.Type {
	case NodeTypeText:
		return renderText(node)
	case NodeTypeHardBreak:
		return "  \n"
	default:
		return ""
	}
}

// renderText renders a text node with its marks to Markdown.
func renderText(node ADFNode) string {
	text := escapeMarkdown(node.Text)

	// Apply marks in order
	for _, mark := range node.Marks {
		text = applyMark(text, mark)
	}

	return text
}

// applyMark wraps text with the appropriate Markdown syntax for a mark.
func applyMark(text string, mark ADFMark) string {
	switch mark.Type {
	case MarkTypeStrong:
		return "**" + text + "**"
	case MarkTypeEm:
		return "*" + text + "*"
	case MarkTypeStrike:
		return "~~" + text + "~~"
	case MarkTypeCode:
		// Don't escape inside code
		return "`" + strings.ReplaceAll(text, "\\", "") + "`"
	case MarkTypeLink:
		href := ""
		if h, ok := mark.Attrs["href"].(string); ok {
			href = h
		}
		// Unescape text for link text
		linkText := strings.ReplaceAll(text, "\\", "")
		return fmt.Sprintf("[%s](%s)", linkText, href)
	default:
		return text
	}
}

// escapeMarkdown escapes special Markdown characters in text.
// Uses minimal escaping to avoid over-escaping content that roundtrips through ADF.
func escapeMarkdown(text string) string {
	// Minimal set of characters that need escaping in inline text:
	// - ` (backticks) - would create inline code
	// - * (asterisks) - would create bold/italic
	// - [ (open bracket) - would start a link
	//
	// Characters we intentionally DON'T escape:
	// - \ (backslash) - escaping this causes double-escaping on roundtrip
	// - | (pipe) - only meaningful in tables, handled separately
	// - _ (underscore) - only triggers emphasis at word boundaries
	// - ] (close bracket) - only meaningful after [
	// - #, +, -, !, . - only special at line start
	runes := []rune(text)
	var result strings.Builder
	for i, r := range runes {
		// Check if previous character was a backslash - if so, this char is already escaped
		alreadyEscaped := i > 0 && runes[i-1] == '\\'

		switch r {
		case '`', '*', '[':
			if !alreadyEscaped {
				result.WriteRune('\\')
			}
			result.WriteRune(r)
		case '_':
			// Only escape underscore if it could trigger emphasis
			// Underscores only trigger emphasis at word boundaries
			// Don't escape if: surrounded by word chars OR part of identifier-like text (with underscores)
			prevIsWord := i > 0 && (isWordChar(runes[i-1]) || runes[i-1] == '_')
			nextIsWord := i < len(runes)-1 && (isWordChar(runes[i+1]) || runes[i+1] == '_')
			// If underscore is between word-like characters, don't escape
			if prevIsWord && nextIsWord {
				result.WriteRune(r)
			} else if !alreadyEscaped {
				result.WriteRune('\\')
				result.WriteRune(r)
			} else {
				result.WriteRune(r)
			}
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

// isWordChar returns true if r is a letter or digit (word character).
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
