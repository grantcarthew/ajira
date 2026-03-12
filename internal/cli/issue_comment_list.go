package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

var (
	commentListLimit int
)

var issueCommentListCmd = &cobra.Command{
	Use:   "list <issue-key>",
	Short: "List comments",
	Long:  "List comments for an issue with ID, author, date, and body.",
	Example: `  ajira issue comment list PROJ-123          # List 5 most recent comments
  ajira issue comment list PROJ-123 -l 20   # List 20 most recent comments
  ajira issue comment list PROJ-123 --json  # JSON output`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueCommentList,
}

func init() {
	issueCommentListCmd.Flags().IntVarP(&commentListLimit, "limit", "l", 5, "Maximum number of comments to show")

	issueCommentCmd.AddCommand(issueCommentListCmd)
}

func runIssueCommentList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	comments, total, err := getComments(ctx, client, issueKey, commentListLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(comments, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		printCommentList(issueKey, comments, total)
	}

	return nil
}

func printCommentList(issueKey string, comments []CommentInfo, total int) {
	if len(comments) == 0 {
		fmt.Printf("No comments for %s\n", issueKey)
		return
	}

	if total > len(comments) {
		fmt.Printf("Comments for %s (%d of %d):\n", issueKey, len(comments), total)
	} else {
		fmt.Printf("Comments for %s (%d):\n", issueKey, len(comments))
	}

	for _, c := range comments {
		fmt.Println()
		fmt.Printf("[%s] [%s] %s:\n", formatDateTime(c.Created), c.ID, c.Author)
		fmt.Print(RenderMarkdown(c.Body))
	}

	if total > len(comments) {
		fmt.Fprintf(os.Stderr, "\nShowing %d of %d comments. Use -l %d to see all.\n", len(comments), total, total)
	}
}
