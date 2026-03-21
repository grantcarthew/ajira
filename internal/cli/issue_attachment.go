package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gcarthew/ajira/internal/api"
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
	return strconv.FormatInt(value, 10) + " " + unit
}

func formatFloat1(f float64) string {
	// Truncate to 1 decimal place, trimming trailing zero.
	i := int64(f * 10)
	whole, frac := i/10, i%10
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac, 10)
}

func formatFloat2(f float64) string {
	// Truncate to 2 decimal places, trimming trailing zeros.
	i := int64(f * 100)
	whole, frac := i/100, i%100
	switch {
	case frac == 0:
		return strconv.FormatInt(whole, 10)
	case frac%10 == 0:
		return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac/10, 10)
	case frac < 10:
		return strconv.FormatInt(whole, 10) + ".0" + strconv.FormatInt(frac, 10)
	default:
		return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac, 10)
	}
}

// validateAttachmentOwnership checks that every requested attachment ID belongs
// to the given issue. All IDs are validated before any action is taken, so a
// bad ID in a multi-ID operation will not result in partial changes.
func validateAttachmentOwnership(ctx context.Context, client *api.Client, issueKey string, ids ...string) error {
	attachments, err := getAttachments(ctx, client, issueKey)
	if err != nil {
		return fmt.Errorf("failed to fetch attachments for %s: %w", issueKey, err)
	}

	owned := make(map[string]bool, len(attachments))
	for _, a := range attachments {
		owned[a.ID] = true
	}

	for _, id := range ids {
		if !owned[id] {
			return fmt.Errorf("attachment %s does not belong to issue %s", id, issueKey)
		}
	}
	return nil
}
