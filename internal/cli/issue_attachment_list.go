package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// attachmentListResponse represents the Jira API response for attachments.
type attachmentListResponse struct {
	Fields struct {
		Attachment []attachmentValue `json:"attachment"`
	} `json:"fields"`
}

type attachmentValue struct {
	ID       string     `json:"id"`
	Filename string     `json:"filename"`
	Size     int64      `json:"size"`
	MimeType string     `json:"mimeType"`
	Author   *userField `json:"author"`
	Created  string     `json:"created"`
	Content  string     `json:"content"`
	Self     string     `json:"self"`
}

var issueAttachmentListCmd = &cobra.Command{
	Use:   "list <issue-key>",
	Short: "List attachments",
	Long:  "List all attachments for an issue with ID, filename, size, author, and date.",
	Example: `  ajira issue attachment list PROJ-123          # List all attachments
  ajira issue attachment list PROJ-123 --json   # JSON output`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueAttachmentList,
}

func init() {
	issueAttachmentCmd.AddCommand(issueAttachmentListCmd)
}

func runIssueAttachmentList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	attachments, err := getAttachments(ctx, client, issueKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch attachments: %w", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(attachments, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		printAttachmentList(issueKey, attachments)
	}

	return nil
}

func getAttachments(ctx context.Context, client *api.Client, key string) ([]AttachmentInfo, error) {
	path := fmt.Sprintf("/issue/%s?fields=attachment", key)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp attachmentListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var attachments []AttachmentInfo
	for _, a := range resp.Fields.Attachment {
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

func printAttachmentList(issueKey string, attachments []AttachmentInfo) {
	if len(attachments) == 0 {
		fmt.Printf("No attachments for %s\n", issueKey)
		return
	}

	fmt.Printf("Attachments for %s (%d):\n\n", issueKey, len(attachments))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tFilename\tSize\tAuthor\tCreated")
	for _, a := range attachments {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			a.ID,
			a.Filename,
			FormatFileSize(a.Size),
			a.Author,
			formatDateTime(a.Created),
		)
	}
	w.Flush()
}
