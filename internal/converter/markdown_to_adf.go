package converter

import (
	"bytes"
	"strings"

	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// MarkdownToADF converts Markdown text to Atlassian Document Format.
func MarkdownToADF(markdown string) (*ADF, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	source := []byte(markdown)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	content := walkNode(doc, source)

	return &ADF{
		Version: ADFVersion,
		Type:    NodeTypeDoc,
		Content: content,
	}, nil
}

// walkNode recursively walks the goldmark AST and converts nodes to ADF.
func walkNode(n ast.Node, source []byte) []ADFNode {
	var nodes []ADFNode

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if node := convertNode(child, source); node != nil {
			nodes = append(nodes, *node)
		}
	}

	return nodes
}

// convertNode converts a single goldmark AST node to an ADF node.
func convertNode(n ast.Node, source []byte) *ADFNode {
	switch node := n.(type) {
	case *ast.Paragraph:
		return convertParagraph(node, source)
	case *ast.TextBlock:
		return convertTextBlock(node, source)
	case *ast.Heading:
		return convertHeading(node, source)
	case *ast.CodeBlock:
		return convertCodeBlock(node, source)
	case *ast.FencedCodeBlock:
		return convertFencedCodeBlock(node, source)
	case *ast.Blockquote:
		return convertBlockquote(node, source)
	case *ast.List:
		return convertList(node, source)
	case *extast.Table:
		return convertTable(node, source)
	case *ast.ThematicBreak:
		return &ADFNode{Type: NodeTypeRule}
	case *ast.HTMLBlock:
		return convertHTMLBlock(node, source)
	default:
		return nil
	}
}

func convertParagraph(n *ast.Paragraph, source []byte) *ADFNode {
	content := convertInlineNodes(n, source)
	if len(content) == 0 {
		return nil
	}
	return &ADFNode{
		Type:    NodeTypeParagraph,
		Content: content,
	}
}

func convertTextBlock(n *ast.TextBlock, source []byte) *ADFNode {
	content := convertInlineNodes(n, source)
	if len(content) == 0 {
		return nil
	}
	return &ADFNode{
		Type:    NodeTypeParagraph,
		Content: content,
	}
}

func convertHeading(n *ast.Heading, source []byte) *ADFNode {
	return &ADFNode{
		Type:    NodeTypeHeading,
		Attrs:   map[string]any{"level": n.Level},
		Content: convertInlineNodes(n, source),
	}
}

func convertCodeBlock(n *ast.CodeBlock, source []byte) *ADFNode {
	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write(line.Value(source))
	}
	text := buf.String()
	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}
	return &ADFNode{
		Type: NodeTypeCodeBlock,
		Content: []ADFNode{
			{Type: NodeTypeText, Text: text},
		},
	}
}

func convertFencedCodeBlock(n *ast.FencedCodeBlock, source []byte) *ADFNode {
	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write(line.Value(source))
	}
	text := buf.String()
	if len(text) > 0 && text[len(text)-1] == '\n' {
		text = text[:len(text)-1]
	}

	// Jira ADF rejects empty text nodes in code blocks - use space placeholder
	if strings.TrimSpace(text) == "" {
		text = " "
	}

	node := &ADFNode{
		Type: NodeTypeCodeBlock,
		Content: []ADFNode{
			{Type: NodeTypeText, Text: text},
		},
	}

	if lang := string(n.Language(source)); lang != "" {
		node.Attrs = map[string]any{"language": lang}
	}

	return node
}

func convertBlockquote(n *ast.Blockquote, source []byte) *ADFNode {
	// ADF does not support nested blockquotes - flatten them
	content := flattenBlockquoteContent(n, source)
	return &ADFNode{
		Type:    NodeTypeBlockquote,
		Content: content,
	}
}

// flattenBlockquoteContent extracts content from blockquotes, flattening any nested blockquotes.
func flattenBlockquoteContent(n ast.Node, source []byte) []ADFNode {
	var nodes []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if bq, ok := child.(*ast.Blockquote); ok {
			// Flatten nested blockquote - include its content directly
			nodes = append(nodes, flattenBlockquoteContent(bq, source)...)
		} else if node := convertNode(child, source); node != nil {
			nodes = append(nodes, *node)
		}
	}
	return nodes
}

func convertList(n *ast.List, source []byte) *ADFNode {
	// Check if this is a task list by examining the first item
	isTaskList := false
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if item, ok := child.(*ast.ListItem); ok {
			// Check if list item has checkbox attribute (goldmark GFM style)
			if item.ChildCount() > 0 {
				if para, ok := item.FirstChild().(*ast.Paragraph); ok {
					if para.ChildCount() > 0 {
						if _, ok := para.FirstChild().(*extast.TaskCheckBox); ok {
							isTaskList = true
							break
						}
					}
				}
				// Also check TextBlock which goldmark may use
				if tb, ok := item.FirstChild().(*ast.TextBlock); ok {
					if tb.ChildCount() > 0 {
						if _, ok := tb.FirstChild().(*extast.TaskCheckBox); ok {
							isTaskList = true
							break
						}
					}
				}
			}
		}
	}

	if isTaskList {
		return convertTaskList(n, source)
	}

	nodeType := NodeTypeBulletList
	if n.IsOrdered() {
		nodeType = NodeTypeOrderedList
	}

	var items []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if item, ok := child.(*ast.ListItem); ok {
			items = append(items, *convertListItem(item, source))
		}
	}

	return &ADFNode{
		Type:    nodeType,
		Content: items,
	}
}

func convertListItem(n *ast.ListItem, source []byte) *ADFNode {
	return &ADFNode{
		Type:    NodeTypeListItem,
		Content: walkNode(n, source),
	}
}

func convertTaskList(n *ast.List, source []byte) *ADFNode {
	var items []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if item, ok := child.(*ast.ListItem); ok {
			items = append(items, *convertTaskItem(item, source))
		}
	}

	return &ADFNode{
		Type:    NodeTypeTaskList,
		Attrs:   map[string]any{"localId": uuid.New().String()},
		Content: items,
	}
}

func convertTaskItem(n *ast.ListItem, source []byte) *ADFNode {
	state := TaskStateTODO

	// Find checkbox and determine state
	// Per ADF spec, taskItem content should be inline nodes directly, not wrapped in paragraphs
	var inlineContent []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch block := child.(type) {
		case *ast.Paragraph:
			// Extract inline nodes from paragraph
			for inline := block.FirstChild(); inline != nil; inline = inline.NextSibling() {
				if cb, ok := inline.(*extast.TaskCheckBox); ok {
					if cb.IsChecked {
						state = TaskStateDONE
					}
					continue // Skip the checkbox itself
				}
				inlineContent = append(inlineContent, convertInlineNodeMulti(inline, source)...)
			}
		case *ast.TextBlock:
			// Extract inline nodes from text block
			for inline := block.FirstChild(); inline != nil; inline = inline.NextSibling() {
				if cb, ok := inline.(*extast.TaskCheckBox); ok {
					if cb.IsChecked {
						state = TaskStateDONE
					}
					continue // Skip the checkbox itself
				}
				inlineContent = append(inlineContent, convertInlineNodeMulti(inline, source)...)
			}
		}
		// Note: We don't process other block types for task items as they should only contain inline content
	}

	return &ADFNode{
		Type:    NodeTypeTaskItem,
		Attrs:   map[string]any{"localId": uuid.New().String(), "state": state},
		Content: inlineContent,
	}
}

func convertTable(n *extast.Table, source []byte) *ADFNode {
	var rows []ADFNode

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if row, ok := child.(*extast.TableRow); ok {
			rows = append(rows, *convertTableRow(row, source, false))
		} else if header, ok := child.(*extast.TableHeader); ok {
			rows = append(rows, *convertTableRow(header, source, true))
		}
	}

	return &ADFNode{
		Type:    NodeTypeTable,
		Content: rows,
	}
}

func convertTableRow(n ast.Node, source []byte, isHeader bool) *ADFNode {
	var cells []ADFNode
	cellType := NodeTypeTableCell
	if isHeader {
		cellType = NodeTypeTableHeader
	}

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if cell, ok := child.(*extast.TableCell); ok {
			cells = append(cells, ADFNode{
				Type: cellType,
				Content: []ADFNode{
					{
						Type:    NodeTypeParagraph,
						Content: convertInlineNodes(cell, source),
					},
				},
			})
		}
	}

	return &ADFNode{
		Type:    NodeTypeTableRow,
		Content: cells,
	}
}

func convertHTMLBlock(n *ast.HTMLBlock, source []byte) *ADFNode {
	// Extract text content from HTML, discard tags
	var buf bytes.Buffer
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write(line.Value(source))
	}

	text := extractTextFromHTML(buf.String())
	if text == "" {
		return nil
	}

	return &ADFNode{
		Type: NodeTypeParagraph,
		Content: []ADFNode{
			{Type: NodeTypeText, Text: text},
		},
	}
}

// convertInlineNodes converts all inline children of a block node.
func convertInlineNodes(n ast.Node, source []byte) []ADFNode {
	var nodes []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		nodes = append(nodes, convertInlineNodeMulti(child, source)...)
	}
	return nodes
}

// convertInlineNodeMulti converts a single inline node to one or more ADF nodes.
// This handles nested marks correctly by returning separate nodes for each mark combination.
func convertInlineNodeMulti(n ast.Node, source []byte) []ADFNode {
	switch node := n.(type) {
	case *ast.Text:
		if n := convertText(node, source); n != nil {
			return []ADFNode{*n}
		}
		return nil
	case *ast.String:
		return []ADFNode{{Type: NodeTypeText, Text: string(node.Value)}}
	case *ast.CodeSpan:
		if n := convertCodeSpan(node, source); n != nil {
			return []ADFNode{*n}
		}
		return nil
	case *ast.Emphasis:
		return convertEmphasisMulti(node, source)
	case *extast.Strikethrough:
		return convertStrikethroughMulti(node, source)
	case *ast.Link:
		if n := convertLink(node, source); n != nil {
			return []ADFNode{*n}
		}
		return nil
	case *ast.AutoLink:
		if n := convertAutoLink(node, source); n != nil {
			return []ADFNode{*n}
		}
		return nil
	case *ast.RawHTML:
		if n := convertRawHTML(node, source); n != nil {
			return []ADFNode{*n}
		}
		return nil
	case *ast.HTMLBlock:
		return nil // Handled at block level
	case *ast.Image:
		// Images out of scope, return alt text
		return []ADFNode{{Type: NodeTypeText, Text: string(node.Text(source))}}
	default:
		return nil
	}
}

// convertInlineNode converts a single inline node to an ADF text node with marks.
// Deprecated: Use convertInlineNodeMulti for proper nested mark handling.
func convertInlineNode(n ast.Node, source []byte) *ADFNode {
	nodes := convertInlineNodeMulti(n, source)
	if len(nodes) == 1 {
		return &nodes[0]
	}
	// For backward compatibility, concatenate if multiple nodes
	if len(nodes) > 1 {
		var text string
		var marks []ADFMark
		for _, node := range nodes {
			text += node.Text
			marks = append(marks, node.Marks...)
		}
		return &ADFNode{Type: NodeTypeText, Text: text, Marks: marks}
	}
	return nil
}

// addMarkToNodes adds a mark to all text nodes in the slice.
// It respects ADF mark compatibility rules - the 'code' mark can only
// combine with 'link', so other marks are skipped for code spans.
func addMarkToNodes(nodes []ADFNode, mark ADFMark) []ADFNode {
	result := make([]ADFNode, len(nodes))
	for i, node := range nodes {
		result[i] = node
		if node.Type == NodeTypeText {
			// Check if node has a code mark - code can only combine with link
			if hasCodeMark(node.Marks) && !isCodeCompatibleMark(mark) {
				continue // Skip incompatible marks for code spans
			}
			// Prepend the mark so outer marks come first
			result[i].Marks = append([]ADFMark{mark}, node.Marks...)
		}
	}
	return result
}

// hasCodeMark checks if the marks slice contains a code mark.
func hasCodeMark(marks []ADFMark) bool {
	for _, m := range marks {
		if m.Type == MarkTypeCode {
			return true
		}
	}
	return false
}

// isCodeCompatibleMark checks if a mark can be combined with code.
// Per ADF spec, code can only combine with link.
func isCodeCompatibleMark(mark ADFMark) bool {
	return mark.Type == MarkTypeLink
}

func convertText(n *ast.Text, source []byte) *ADFNode {
	text := string(n.Segment.Value(source))

	// Handle hard breaks (explicit line breaks)
	if n.HardLineBreak() {
		return &ADFNode{Type: NodeTypeHardBreak}
	}

	// Handle soft breaks (become spaces per Markdown spec)
	if n.SoftLineBreak() {
		text = text + " "
	}

	if text == "" {
		return nil
	}

	return &ADFNode{Type: NodeTypeText, Text: text}
}

func convertCodeSpan(n *ast.CodeSpan, source []byte) *ADFNode {
	var buf bytes.Buffer
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			buf.Write(text.Segment.Value(source))
		}
	}

	return &ADFNode{
		Type:  NodeTypeText,
		Text:  buf.String(),
		Marks: []ADFMark{{Type: MarkTypeCode}},
	}
}

// convertEmphasisMulti converts emphasis to multiple ADF nodes, preserving nested marks.
func convertEmphasisMulti(n *ast.Emphasis, source []byte) []ADFNode {
	markType := MarkTypeEm
	if n.Level == 2 {
		markType = MarkTypeStrong
	}

	// Convert all children and add the emphasis mark to each
	var nodes []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		nodes = append(nodes, convertInlineNodeMulti(child, source)...)
	}

	return addMarkToNodes(nodes, ADFMark{Type: markType})
}

// convertEmphasis converts emphasis - kept for backward compatibility.
func convertEmphasis(n *ast.Emphasis, source []byte) *ADFNode {
	nodes := convertEmphasisMulti(n, source)
	if len(nodes) == 1 {
		return &nodes[0]
	}
	if len(nodes) > 1 {
		// This shouldn't happen in normal use, but handle gracefully
		var text string
		var marks []ADFMark
		for _, node := range nodes {
			text += node.Text
			if len(marks) == 0 {
				marks = node.Marks
			}
		}
		return &ADFNode{Type: NodeTypeText, Text: text, Marks: marks}
	}
	return nil
}

// convertStrikethroughMulti converts strikethrough to multiple ADF nodes, preserving nested marks.
func convertStrikethroughMulti(n *extast.Strikethrough, source []byte) []ADFNode {
	var nodes []ADFNode
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		nodes = append(nodes, convertInlineNodeMulti(child, source)...)
	}

	return addMarkToNodes(nodes, ADFMark{Type: MarkTypeStrike})
}

// convertStrikethrough converts strikethrough - kept for backward compatibility.
func convertStrikethrough(n *extast.Strikethrough, source []byte) *ADFNode {
	nodes := convertStrikethroughMulti(n, source)
	if len(nodes) == 1 {
		return &nodes[0]
	}
	if len(nodes) > 1 {
		var text string
		var marks []ADFMark
		for _, node := range nodes {
			text += node.Text
			if len(marks) == 0 {
				marks = node.Marks
			}
		}
		return &ADFNode{Type: NodeTypeText, Text: text, Marks: marks}
	}
	return nil
}

func convertLink(n *ast.Link, source []byte) *ADFNode {
	var textContent string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			textContent += string(text.Segment.Value(source))
		}
	}

	if textContent == "" {
		textContent = string(n.Destination)
	}

	return &ADFNode{
		Type: NodeTypeText,
		Text: textContent,
		Marks: []ADFMark{
			{
				Type:  MarkTypeLink,
				Attrs: map[string]any{"href": string(n.Destination)},
			},
		},
	}
}

func convertAutoLink(n *ast.AutoLink, source []byte) *ADFNode {
	url := string(n.URL(source))
	return &ADFNode{
		Type: NodeTypeText,
		Text: url,
		Marks: []ADFMark{
			{
				Type:  MarkTypeLink,
				Attrs: map[string]any{"href": url},
			},
		},
	}
}

func convertRawHTML(n *ast.RawHTML, source []byte) *ADFNode {
	// Extract text content from inline HTML
	var buf bytes.Buffer
	for i := 0; i < n.Segments.Len(); i++ {
		seg := n.Segments.At(i)
		buf.Write(seg.Value(source))
	}

	text := extractTextFromHTML(buf.String())
	if text == "" {
		return nil
	}

	return &ADFNode{Type: NodeTypeText, Text: text}
}

// extractTextFromHTML extracts text content from HTML, discarding tags.
func extractTextFromHTML(html string) string {
	var result bytes.Buffer
	inTag := false

	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}
