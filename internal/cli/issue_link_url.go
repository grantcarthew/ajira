package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// RemoteLinkResult represents the result of adding a remote link.
type RemoteLinkResult struct {
	ID    int    `json:"id"`
	Self  string `json:"self"`
	Issue string `json:"issue"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

// remoteLinkRequest represents the request body for creating a remote link.
type remoteLinkRequest struct {
	Object remoteLinkObject `json:"object"`
}

type remoteLinkObject struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// remoteLinkResponse represents the API response for creating a remote link.
type remoteLinkResponse struct {
	ID   int    `json:"id"`
	Self string `json:"self"`
}

var issueLinkURLCmd = &cobra.Command{
	Use:     "url <issue-key> <url> [title]",
	Aliases: []string{"web"},
	Short:   "Add a web URL to an issue",
	Long:    "Add an external web URL as a remote link to an issue.",
	Example: `  ajira issue link url GCP-123 https://github.com/org/repo/pull/42
  ajira issue link url GCP-123 https://docs.example.com "API Documentation"
  ajira issue link web GCP-123 https://example.com`,
	Args:         cobra.RangeArgs(2, 3),
	SilenceUsage: true,
	RunE:         runIssueLinkURL,
}

func init() {
	issueLinkCmd.AddCommand(issueLinkURLCmd)
}

func runIssueLinkURL(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]
	url := args[1]
	title := url // Default title is the URL
	if len(args) > 2 {
		title = args[2]
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	result, err := createRemoteLink(ctx, client, issueKey, url, title)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to create remote link: %v", err)
	}

	result.Issue = issueKey
	result.URL = url
	result.Title = title

	if JSONOutput() {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

func createRemoteLink(ctx context.Context, client *api.Client, issueKey, url, title string) (*RemoteLinkResult, error) {
	req := remoteLinkRequest{
		Object: remoteLinkObject{
			URL:   url,
			Title: title,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/remotelink", issueKey)
	respBody, err := client.Post(ctx, path, body)
	if err != nil {
		return nil, err
	}

	var resp remoteLinkResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &RemoteLinkResult{
		ID:   resp.ID,
		Self: resp.Self,
	}, nil
}
