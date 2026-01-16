package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/converter"
	"github.com/spf13/cobra"
)

// CommentResult represents the result of adding a comment.
type CommentResult struct {
	ID      string `json:"id"`
	Self    string `json:"self"`
	Created string `json:"created"`
}

// commentAddRequest represents the request body for adding a comment.
type commentAddRequest struct {
	Body *converter.ADF `json:"body"`
}

var (
	commentBody  string
	commentFile  string
	commentStdin bool
)

var issueCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage comments",
	Long:  "Commands for managing issue comments.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var issueCommentEditCmd = &cobra.Command{
	Use:   "edit <issue-key> <comment-id> [text]",
	Short: "Edit comment",
	Long:  "Edit an existing comment. Use 'issue view -c N' to find comment IDs.",
	Example: `  ajira issue comment edit PROJ-123 12345 "Updated text"   # Inline
  ajira issue comment edit PROJ-123 12345 -b "New text"    # Via --body
  ajira issue comment edit PROJ-123 12345 -f comment.md    # From file
  echo "text" | ajira issue comment edit PROJ-123 12345 -f - # From stdin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 || len(args) > 3 {
			return fmt.Errorf("requires 2 or 3 arguments: <issue-key> <comment-id> [text]")
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueCommentEdit,
}

var issueCommentAddCmd = &cobra.Command{
	Use:   "add <issue-key> [text]",
	Short: "Add comment",
	Long:  "Add a comment to an issue. Text via argument, --body, or --file. Use --stdin for batch (not with -f -).",
	Example: `  ajira issue comment add PROJ-123 "Comment text"   # Inline comment
  ajira issue comment add PROJ-123 -f comment.md    # From file
  echo "text" | ajira issue comment add PROJ-123 -f - # From stdin
  echo -e "PROJ-1\nPROJ-2" | ajira issue comment add --stdin "Comment for all"  # Batch`,
	Args: func(cmd *cobra.Command, args []string) error {
		if commentStdin {
			// With --stdin, comment text must be provided via arg or --body (not --file -)
			if commentFile == "-" {
				return fmt.Errorf("cannot use --stdin with --file - (both read from stdin)")
			}
			// Need at least comment text as argument or via --body/--file
			if len(args) == 0 && commentBody == "" && commentFile == "" {
				return fmt.Errorf("with --stdin, comment text must be provided via argument, --body, or --file")
			}
		} else {
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("requires 1 or 2 arguments: <issue-key> [text]")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueCommentAdd,
}

func init() {
	issueCommentAddCmd.Flags().StringVarP(&commentBody, "body", "b", "", "Comment text in Markdown")
	issueCommentAddCmd.Flags().StringVarP(&commentFile, "file", "f", "", "Read comment from file (use - for stdin)")
	issueCommentAddCmd.Flags().BoolVar(&commentStdin, "stdin", false, "Read issue keys from stdin (one per line)")

	issueCommentEditCmd.Flags().StringVarP(&commentBody, "body", "b", "", "Comment text in Markdown")
	issueCommentEditCmd.Flags().StringVarP(&commentFile, "file", "f", "", "Read comment from file (use - for stdin)")

	issueCommentCmd.AddCommand(issueCommentAddCmd)
	issueCommentCmd.AddCommand(issueCommentEditCmd)
	issueCmd.AddCommand(issueCommentCmd)
}

func runIssueCommentAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	var issueKeys []string
	var commentText string

	if commentStdin {
		// Read keys from stdin, get comment from args/flags
		issueKeys, err = ReadKeysFromStdin()
		if err != nil {
			return err
		}
		if len(issueKeys) == 0 {
			return fmt.Errorf("no issue keys provided via stdin")
		}
		commentText, err = getCommentTextForBatch(args)
		if err != nil {
			return err
		}
	} else {
		issueKeys = []string{args[0]}
		commentText, err = getCommentText(args)
		if err != nil {
			return fmt.Errorf("failed to read comment: %w", err)
		}
	}

	if commentText == "" {
		return fmt.Errorf("comment text is required (provide as argument, --body, or --file)")
	}

	// Dry-run mode
	if DryRun() {
		preview := commentText
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		if len(issueKeys) == 1 {
			PrintDryRun(fmt.Sprintf("add comment to %s: %q", issueKeys[0], preview))
		} else {
			PrintDryRunBatch(issueKeys, fmt.Sprintf("add comment: %q", preview))
		}
		return nil
	}

	// Single comment
	if len(issueKeys) == 1 {
		result, err := addComment(ctx, client, issueKeys[0], commentText)
		if err != nil {
			return err
		}

		if JSONOutput() {
			PrintSuccessJSON(result)
		} else {
			PrintSuccess(IssueURL(cfg.BaseURL, issueKeys[0]))
		}
		return nil
	}

	// Batch comments
	var results []BatchResult
	for _, key := range issueKeys {
		_, err := addComment(ctx, client, key, commentText)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
		} else {
			results = append(results, BatchResult{Key: key, Success: true})
		}
	}

	return PrintBatchResults(results)
}

func getCommentText(args []string) (string, error) {
	// Priority: file > body > positional arg
	if commentFile != "" {
		if commentFile == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
		data, err := os.ReadFile(commentFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if commentBody != "" {
		return commentBody, nil
	}

	if len(args) > 1 {
		return args[1], nil
	}

	return "", nil
}

func getCommentTextForBatch(args []string) (string, error) {
	// For batch mode, stdin is used for keys, so file must be actual file (not -)
	if commentFile != "" && commentFile != "-" {
		data, err := os.ReadFile(commentFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if commentBody != "" {
		return commentBody, nil
	}

	if len(args) > 0 {
		return args[0], nil
	}

	return "", nil
}

func addComment(ctx context.Context, client *api.Client, issueKey, text string) (*CommentResult, error) {
	adf, err := converter.MarkdownToADF(text)
	if err != nil {
		return nil, fmt.Errorf("failed to convert comment: %w", err)
	}

	req := commentAddRequest{Body: adf}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/comment", issueKey)
	respBody, err := client.Post(ctx, path, body)
	if err != nil {
		return nil, err
	}

	var result CommentResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

func runIssueCommentEdit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]
	commentID := args[1]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	commentText, err := getCommentTextForEdit(args)
	if err != nil {
		return fmt.Errorf("failed to read comment: %w", err)
	}

	if commentText == "" {
		return fmt.Errorf("comment text is required (provide as argument, --body, or --file)")
	}

	if DryRun() {
		preview := commentText
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		PrintDryRun(fmt.Sprintf("edit comment %s on %s: %q", commentID, issueKey, preview))
		return nil
	}

	result, err := editComment(ctx, client, issueKey, commentID, commentText)
	if err != nil {
		return err
	}

	if JSONOutput() {
		PrintSuccessJSON(result)
	} else {
		PrintSuccess(IssueURL(cfg.BaseURL, issueKey))
	}
	return nil
}

func getCommentTextForEdit(args []string) (string, error) {
	// Priority: file > body > positional arg
	if commentFile != "" {
		if commentFile == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
		data, err := os.ReadFile(commentFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if commentBody != "" {
		return commentBody, nil
	}

	if len(args) > 2 {
		return args[2], nil
	}

	return "", nil
}

func editComment(ctx context.Context, client *api.Client, issueKey, commentID, text string) (*CommentResult, error) {
	adf, err := converter.MarkdownToADF(text)
	if err != nil {
		return nil, fmt.Errorf("failed to convert comment: %w", err)
	}

	req := commentAddRequest{Body: adf}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/comment/%s", issueKey, commentID)
	respBody, err := client.Put(ctx, path, body)
	if err != nil {
		return nil, err
	}

	var result CommentResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
