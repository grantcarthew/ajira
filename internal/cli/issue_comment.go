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
	commentBody string
	commentFile string
)

var issueCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage issue comments",
	Long:  "Commands for managing comments on Jira issues.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var issueCommentAddCmd = &cobra.Command{
	Use:           "add <issue-key> [text]",
	Short:         "Add a comment to an issue",
	Long:          "Add a comment to a Jira issue. Comment text can be provided as an argument, via --body, --file, or stdin.",
	Args:          cobra.RangeArgs(1, 2),
	SilenceUsage:  true,
	RunE:          runIssueCommentAdd,
}

func init() {
	issueCommentAddCmd.Flags().StringVarP(&commentBody, "body", "b", "", "Comment text in Markdown")
	issueCommentAddCmd.Flags().StringVarP(&commentFile, "file", "f", "", "Read comment from file (use - for stdin)")

	issueCommentCmd.AddCommand(issueCommentAddCmd)
	issueCmd.AddCommand(issueCommentCmd)
}

func runIssueCommentAdd(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	// Get comment text from: positional arg > file > body flag
	commentText, err := getCommentText(args)
	if err != nil {
		return Errorf("failed to read comment: %v", err)
	}

	if commentText == "" {
		return Errorf("comment text is required (provide as argument, --body, or --file)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	result, err := addComment(client, issueKey, commentText)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to add comment: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
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

func addComment(client *api.Client, issueKey, text string) (*CommentResult, error) {
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
	respBody, err := client.Post(context.Background(), path, body)
	if err != nil {
		return nil, err
	}

	var result CommentResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
