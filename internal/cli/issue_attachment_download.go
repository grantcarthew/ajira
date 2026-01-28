package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// AttachmentDownloadResult represents the result of downloading an attachment.
type AttachmentDownloadResult struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Output   string `json:"output"`
}

// attachmentMetaResponse represents the metadata response for a single attachment.
// Note: Jira API returns ID as a number for this endpoint, not a string.
type attachmentMetaResponse struct {
	ID       json.Number `json:"id"`
	Filename string      `json:"filename"`
	Size     int64       `json:"size"`
	MimeType string      `json:"mimeType"`
	Author   *userField  `json:"author"`
	Created  string      `json:"created"`
	Content  string      `json:"content"`
}

var (
	downloadOutput string
)

var issueAttachmentDownloadCmd = &cobra.Command{
	Use:   "download <issue-key> <attachment-id>",
	Short: "Download attachment",
	Long:  "Download an attachment to the current directory.",
	Example: `  ajira issue attachment download PROJ-123 10001              # Download to original filename
  ajira issue attachment download PROJ-123 10001 -o custom.pdf  # Download with custom name`,
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE:         runIssueAttachmentDownload,
}

func init() {
	issueAttachmentDownloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", "", "Output filename (default: original filename)")

	issueAttachmentCmd.AddCommand(issueAttachmentDownloadCmd)
}

func runIssueAttachmentDownload(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	_ = args[0] // issueKey - used for context but not required by API
	attachmentID := args[1]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Get attachment metadata to determine filename
	meta, err := getAttachmentMeta(ctx, client, attachmentID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch attachment metadata: %w", err)
	}

	// Determine output filename
	outputFile := downloadOutput
	if outputFile == "" {
		outputFile = meta.Filename
	}

	// Dry-run mode
	if DryRun() {
		PrintDryRun(fmt.Sprintf("download attachment %s to %s", attachmentID, outputFile))
		return nil
	}

	// Download attachment content
	content, _, err := client.GetRaw(ctx, fmt.Sprintf("/attachment/content/%s", attachmentID))
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to download attachment: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if JSONOutput() {
		result := AttachmentDownloadResult{
			ID:       meta.ID.String(),
			Filename: meta.Filename,
			Size:     meta.Size,
			Output:   outputFile,
		}
		PrintSuccessJSON(result)
	} else {
		PrintSuccess(fmt.Sprintf("Downloaded %s (%s)", outputFile, FormatFileSize(meta.Size)))
	}

	return nil
}

func getAttachmentMeta(ctx context.Context, client *api.Client, attachmentID string) (*attachmentMetaResponse, error) {
	path := fmt.Sprintf("/attachment/%s", attachmentID)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var meta attachmentMetaResponse
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &meta, nil
}
