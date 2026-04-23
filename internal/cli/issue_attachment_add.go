package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/grantcarthew/ajira/internal/api"
	"github.com/grantcarthew/ajira/internal/config"
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

// countingWriter counts bytes written to it without storing them.
type countingWriter struct{ n int64 }

func (w *countingWriter) Write(p []byte) (int, error) {
	w.n += int64(len(p))
	return len(p), nil
}

// calcMultipartSize does a dry run through the multipart writer to compute the
// exact Content-Length without reading file content. File sizes are added
// directly from the provided FileInfo slice. Returns total byte count and the
// boundary string to reuse on the real writer.
func calcMultipartSize(filePaths []string, infos []os.FileInfo) (int64, string, error) {
	cw := &countingWriter{}
	mw := multipart.NewWriter(cw)
	for i, fp := range filePaths {
		if _, err := mw.CreateFormFile("file", filepath.Base(fp)); err != nil {
			return 0, "", err
		}
		cw.n += infos[i].Size()
	}
	if err := mw.Close(); err != nil {
		return 0, "", err
	}
	return cw.n, mw.Boundary(), nil
}

func uploadAttachments(ctx context.Context, client *api.Client, issueKey string, filePaths []string) ([]AttachmentInfo, error) {
	// Stat all files upfront so we can compute Content-Length before streaming.
	infos := make([]os.FileInfo, len(filePaths))
	for i, fp := range filePaths {
		fi, err := os.Stat(fp)
		if err != nil {
			return nil, fmt.Errorf("cannot stat %s: %w", fp, err)
		}
		infos[i] = fi
	}

	// Dry-run: measure exact multipart body size and capture the boundary.
	totalSize, boundary, err := calcMultipartSize(filePaths, infos)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate upload size: %w", err)
	}

	// Stream the multipart body through a pipe so nothing is buffered in memory.
	pr, pw := io.Pipe()
	go func() {
		mw := multipart.NewWriter(pw)
		if err := mw.SetBoundary(boundary); err != nil {
			pw.CloseWithError(err)
			return
		}
		for _, fp := range filePaths {
			part, err := mw.CreateFormFile("file", filepath.Base(fp))
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			f, err := os.Open(fp)
			if err != nil {
				pw.CloseWithError(fmt.Errorf("failed to open %s: %w", fp, err))
				return
			}
			_, err = io.Copy(part, f)
			f.Close()
			if err != nil {
				pw.CloseWithError(err)
				return
			}
		}
		if err := mw.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
		pw.Close()
	}()

	contentType := "multipart/form-data; boundary=" + boundary
	path := fmt.Sprintf("/issue/%s/attachments", issueKey)
	respBody, err := client.PostMultipart(ctx, path, contentType, pr, totalSize)
	if err != nil {
		pr.CloseWithError(err) // unblock goroutine if still writing
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
