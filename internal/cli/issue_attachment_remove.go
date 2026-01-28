package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// AttachmentRemoveResult represents the result of removing attachments.
type AttachmentRemoveResult struct {
	IssueKey string   `json:"issueKey"`
	Removed  []string `json:"removed"`
	Count    int      `json:"count"`
}

var issueAttachmentRemoveCmd = &cobra.Command{
	Use:   "remove <issue-key> <attachment-id> [attachment-id...]",
	Short: "Delete attachments",
	Long:  "Delete one or more attachments from an issue.",
	Example: `  ajira issue attachment remove PROJ-123 10001
  ajira issue attachment remove PROJ-123 10001 10002 10003
  ajira issue attachment remove PROJ-123 10001 --dry-run  # Preview deletion`,
	Args:         cobra.MinimumNArgs(2),
	SilenceUsage: true,
	RunE:         runIssueAttachmentRemove,
}

func init() {
	issueAttachmentCmd.AddCommand(issueAttachmentRemoveCmd)
}

func runIssueAttachmentRemove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]
	attachmentIDs := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Dry-run mode
	if DryRun() {
		if len(attachmentIDs) == 1 {
			PrintDryRun(fmt.Sprintf("remove attachment %s from %s", attachmentIDs[0], issueKey))
		} else {
			PrintDryRun(fmt.Sprintf("remove %d attachments from %s", len(attachmentIDs), issueKey))
		}
		return nil
	}

	client := api.NewClient(cfg)

	// Delete each attachment
	var removed []string
	for _, id := range attachmentIDs {
		if err := deleteAttachment(ctx, client, id); err != nil {
			if apiErr, ok := err.(*api.APIError); ok {
				return fmt.Errorf("API error removing %s: %w", id, apiErr)
			}
			return fmt.Errorf("failed to remove attachment %s: %w", id, err)
		}
		removed = append(removed, id)
	}

	if JSONOutput() {
		result := AttachmentRemoveResult{
			IssueKey: issueKey,
			Removed:  removed,
			Count:    len(removed),
		}
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		PrintSuccess(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

func deleteAttachment(ctx context.Context, client *api.Client, attachmentID string) error {
	path := fmt.Sprintf("/attachment/%s", attachmentID)
	_, err := client.Delete(ctx, path)
	return err
}
