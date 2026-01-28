package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// AttachmentAddResult represents the result of adding attachments.
type AttachmentAddResult struct {
	IssueKey    string           `json:"issueKey"`
	Attachments []AttachmentInfo `json:"attachments"`
}

var issueAttachmentAddCmd = &cobra.Command{
	Use:   "add <issue-key> <file> [file...]",
	Short: "Upload attachments",
	Long:  "Upload one or more files to an issue.",
	Example: `  ajira issue attachment add PROJ-123 screenshot.png
  ajira issue attachment add PROJ-123 file1.pdf file2.png file3.doc
  ajira issue attachment add PROJ-123 *.log`,
	Args:         cobra.MinimumNArgs(2),
	SilenceUsage: true,
	RunE:         runIssueAttachmentAdd,
}

func init() {
	issueAttachmentCmd.AddCommand(issueAttachmentAddCmd)
}

func runIssueAttachmentAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]
	filePaths := args[1:]

	// Validate all files exist before proceeding
	for _, path := range filePaths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", path)
			}
			return fmt.Errorf("cannot access file %s: %w", path, err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Dry-run mode
	if DryRun() {
		if len(filePaths) == 1 {
			PrintDryRun(fmt.Sprintf("upload %s to %s", filePaths[0], issueKey))
		} else {
			PrintDryRun(fmt.Sprintf("upload %d files to %s", len(filePaths), issueKey))
		}
		return nil
	}

	client := api.NewClient(cfg)

	attachments, err := uploadAttachments(ctx, client, issueKey, filePaths)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to upload attachments: %w", err)
	}

	if JSONOutput() {
		result := AttachmentAddResult{
			IssueKey:    issueKey,
			Attachments: attachments,
		}
		PrintSuccessJSON(result)
	} else {
		PrintSuccess(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

func uploadAttachments(ctx context.Context, client *api.Client, issueKey string, filePaths []string) ([]AttachmentInfo, error) {
	// Build multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", filePath, err)
		}

		filename := filepath.Base(filePath)
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to create form field: %w", err)
		}

		_, err = io.Copy(part, file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize multipart form: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/attachments", issueKey)
	respBody, err := client.PostMultipart(ctx, path, writer.FormDataContentType(), body.Bytes())
	if err != nil {
		return nil, err
	}

	// Parse response (array of attachment objects)
	var resp []attachmentValue
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var attachments []AttachmentInfo
	for _, a := range resp {
		info := AttachmentInfo{
			ID:       a.ID,
			Filename: a.Filename,
			Size:     a.Size,
			MimeType: a.MimeType,
			Created:  a.Created,
			Content:  a.Content,
		}
		if a.Author != nil {
			info.Author = a.Author.DisplayName
		}
		attachments = append(attachments, info)
	}

	return attachments, nil
}
