// Package converter provides bidirectional conversion between Markdown and
// Atlassian Document Format (ADF).
package converter

// ADF represents an Atlassian Document Format document.
type ADF struct {
	Version int       `json:"version"`
	Type    string    `json:"type"`
	Content []ADFNode `json:"content"`
}

// ADFNode represents a node in the ADF document tree.
type ADFNode struct {
	Type    string            `json:"type"`
	Attrs   map[string]any    `json:"attrs,omitempty"`
	Content []ADFNode         `json:"content,omitempty"`
	Marks   []ADFMark         `json:"marks,omitempty"`
	Text    string            `json:"text,omitempty"`
}

// ADFMark represents a text formatting mark in ADF.
type ADFMark struct {
	Type  string         `json:"type"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

// ADF node types.
const (
	NodeTypeDoc         = "doc"
	NodeTypeParagraph   = "paragraph"
	NodeTypeText        = "text"
	NodeTypeHeading     = "heading"
	NodeTypeCodeBlock   = "codeBlock"
	NodeTypeBlockquote  = "blockquote"
	NodeTypeBulletList  = "bulletList"
	NodeTypeOrderedList = "orderedList"
	NodeTypeListItem    = "listItem"
	NodeTypeTaskList    = "taskList"
	NodeTypeTaskItem    = "taskItem"
	NodeTypeTable       = "table"
	NodeTypeTableRow    = "tableRow"
	NodeTypeTableHeader = "tableHeader"
	NodeTypeTableCell   = "tableCell"
	NodeTypeRule        = "rule"
	NodeTypeHardBreak   = "hardBreak"
)

// ADF mark types.
const (
	MarkTypeStrong = "strong"
	MarkTypeEm     = "em"
	MarkTypeStrike = "strike"
	MarkTypeCode   = "code"
	MarkTypeLink   = "link"
)

// Task item states.
const (
	TaskStateTODO = "TODO"
	TaskStateDONE = "DONE"
)

// ADFVersion is the ADF document version.
const ADFVersion = 1
