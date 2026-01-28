package cli

import (
	"github.com/spf13/cobra"
)

var issueAttachmentCmd = &cobra.Command{
	Use:     "attachment",
	Aliases: []string{"attachments"},
	Short:   "Manage attachments",
	Long:    "Commands for managing issue file attachments.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	issueCmd.AddCommand(issueAttachmentCmd)
}

// AttachmentInfo represents an attachment for display.
type AttachmentInfo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Author   string `json:"author"`
	Created  string `json:"created"`
	Content  string `json:"content,omitempty"`
}

// FormatFileSize formats a file size in bytes to a human-readable string.
// Uses 1024-based units with SI prefixes (KB, MB, GB).
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return formatSize(float64(bytes)/float64(GB), "GB")
	case bytes >= MB:
		return formatSize(float64(bytes)/float64(MB), "MB")
	case bytes >= KB:
		return formatSize(float64(bytes)/float64(KB), "KB")
	default:
		return formatSizeInt(bytes, "B")
	}
}

func formatSize(value float64, unit string) string {
	if value >= 100 {
		return formatSizeInt(int64(value), unit)
	}
	if value >= 10 {
		return formatFloat1(value) + " " + unit
	}
	return formatFloat2(value) + " " + unit
}

func formatSizeInt(value int64, unit string) string {
	return intToStr(value) + " " + unit
}

func formatFloat1(f float64) string {
	// Format with 1 decimal place, trimming trailing zero
	i := int64(f * 10)
	whole := i / 10
	frac := i % 10
	if frac == 0 {
		return intToStr(whole)
	}
	return intToStr(whole) + "." + intToStr(frac)
}

func formatFloat2(f float64) string {
	// Format with up to 2 decimal places, trimming trailing zeros
	i := int64(f * 100)
	whole := i / 100
	frac := i % 100
	if frac == 0 {
		return intToStr(whole)
	}
	if frac%10 == 0 {
		return intToStr(whole) + "." + intToStr(frac/10)
	}
	if frac < 10 {
		return intToStr(whole) + ".0" + intToStr(frac)
	}
	return intToStr(whole) + "." + intToStr(frac)
}

func intToStr(i int64) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + intToStr(-i)
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
