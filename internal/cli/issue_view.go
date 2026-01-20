package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/converter"
	"github.com/gcarthew/ajira/internal/width"
	"github.com/spf13/cobra"
)

// IssueDetail represents a full Jira issue for output.
type IssueDetail struct {
	Key         string        `json:"key"`
	Summary     string        `json:"summary"`
	Status      string        `json:"status"`
	Type        string        `json:"type"`
	Priority    string        `json:"priority"`
	Assignee    string        `json:"assignee"`
	Reporter    string        `json:"reporter"`
	Created     string        `json:"created"`
	Updated     string        `json:"updated"`
	Description string        `json:"description"`
	Labels      []string      `json:"labels"`
	Project     string        `json:"project"`
	Links       []LinkInfo    `json:"links,omitempty"`
	Comments    []CommentInfo `json:"comments,omitempty"`
}

// LinkInfo represents a linked issue for display.
type LinkInfo struct {
	Direction string `json:"direction"`
	Key       string `json:"key"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
}

// CommentInfo represents a comment on an issue.
type CommentInfo struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Created string `json:"created"`
	Body    string `json:"body"`
}

// commentsResponse matches the Jira comments API response.
type commentsResponse struct {
	Comments   []commentValue `json:"comments"`
	Total      int            `json:"total"`
	MaxResults int            `json:"maxResults"`
	StartAt    int            `json:"startAt"`
}

type commentValue struct {
	ID      string          `json:"id"`
	Author  *userField      `json:"author"`
	Created string          `json:"created"`
	Body    json.RawMessage `json:"body"`
}

// issueDetailResponse matches the Jira issue API response.
type issueDetailResponse struct {
	Key    string            `json:"key"`
	Fields issueDetailFields `json:"fields"`
}

type issueDetailFields struct {
	Summary     string            `json:"summary"`
	Status      *statusField      `json:"status"`
	IssueType   *issueType        `json:"issuetype"`
	Priority    *priorityField    `json:"priority"`
	Assignee    *userField        `json:"assignee"`
	Reporter    *userField        `json:"reporter"`
	Created     string            `json:"created"`
	Updated     string            `json:"updated"`
	Description json.RawMessage   `json:"description"`
	Labels      []string          `json:"labels"`
	Project     *projectField     `json:"project"`
	IssueLinks  []issueLinkDetail `json:"issuelinks"`
}

type issueLinkDetail struct {
	ID           string              `json:"id"`
	Type         issueLinkTypeDetail `json:"type"`
	InwardIssue  *linkedIssueDetail  `json:"inwardIssue,omitempty"`
	OutwardIssue *linkedIssueDetail  `json:"outwardIssue,omitempty"`
}

type issueLinkTypeDetail struct {
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

type linkedIssueDetail struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string       `json:"summary"`
		Status  *statusField `json:"status"`
	} `json:"fields"`
}

type projectField struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

var (
	viewCommentCount int
)

var issueViewCmd = &cobra.Command{
	Use:   "view <issue-key>",
	Short: "View issue",
	Long:  "Display issue details. Use -c to include comments.",
	Example: `  ajira issue view PROJ-123           # View issue details
  ajira issue view PROJ-123 -c 5      # Include 5 recent comments
  ajira issue view PROJ-123 --json    # JSON output`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueView,
}

func init() {
	issueViewCmd.Flags().IntVarP(&viewCommentCount, "comments", "c", 0, "Number of recent comments to show")

	issueCmd.AddCommand(issueViewCmd)
}

func runIssueView(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	issue, err := getIssue(ctx, client, issueKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch issue: %w", err)
	}

	// Fetch comments if requested
	if viewCommentCount > 0 {
		comments, err := getComments(ctx, client, issueKey, viewCommentCount)
		if err != nil {
			// Non-fatal: skip comments but warn in verbose mode
			if Verbose() {
				fmt.Fprintf(os.Stderr, "warning: failed to fetch comments: %v\n", err)
			}
			comments = nil
		}
		issue.Comments = comments
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(issue, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		printIssueDetail(issue)
	}

	return nil
}

func getIssue(ctx context.Context, client *api.Client, key string) (*IssueDetail, error) {
	path := fmt.Sprintf("/issue/%s", key)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp issueDetailResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	detail := &IssueDetail{
		Key:     resp.Key,
		Summary: resp.Fields.Summary,
		Created: resp.Fields.Created,
		Updated: resp.Fields.Updated,
		Labels:  resp.Fields.Labels,
	}

	if resp.Fields.Status != nil {
		detail.Status = resp.Fields.Status.Name
	}
	if resp.Fields.IssueType != nil {
		detail.Type = resp.Fields.IssueType.Name
	}
	if resp.Fields.Priority != nil {
		detail.Priority = resp.Fields.Priority.Name
	}
	if resp.Fields.Assignee != nil {
		detail.Assignee = resp.Fields.Assignee.DisplayName
	}
	if resp.Fields.Reporter != nil {
		detail.Reporter = resp.Fields.Reporter.DisplayName
	}
	if resp.Fields.Project != nil {
		detail.Project = resp.Fields.Project.Key
	}

	// Convert description from ADF to Markdown
	if len(resp.Fields.Description) > 0 && string(resp.Fields.Description) != "null" {
		detail.Description = converter.ADFToMarkdown(resp.Fields.Description)
	}

	// Parse issue links
	for _, link := range resp.Fields.IssueLinks {
		var info LinkInfo
		if link.OutwardIssue != nil {
			info.Direction = link.Type.Outward
			info.Key = link.OutwardIssue.Key
			info.Summary = link.OutwardIssue.Fields.Summary
			if link.OutwardIssue.Fields.Status != nil {
				info.Status = link.OutwardIssue.Fields.Status.Name
			}
		} else if link.InwardIssue != nil {
			info.Direction = link.Type.Inward
			info.Key = link.InwardIssue.Key
			info.Summary = link.InwardIssue.Fields.Summary
			if link.InwardIssue.Fields.Status != nil {
				info.Status = link.InwardIssue.Fields.Status.Name
			}
		} else {
			continue
		}
		detail.Links = append(detail.Links, info)
	}

	return detail, nil
}

func printIssueDetail(issue *IssueDetail) {
	fmt.Printf("%s: %s\n", issue.Key, issue.Summary)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Project:  %s\n", issue.Project)
	fmt.Printf("Type:     %s\n", issue.Type)
	fmt.Printf("Status:   %s\n", issue.Status)
	fmt.Printf("Priority: %s\n", issue.Priority)

	assignee := issue.Assignee
	if assignee == "" {
		assignee = "Unassigned"
	}
	fmt.Printf("Assignee: %s\n", assignee)
	fmt.Printf("Reporter: %s\n", issue.Reporter)

	if len(issue.Labels) > 0 {
		fmt.Printf("Labels:   %s\n", strings.Join(issue.Labels, ", "))
	}

	fmt.Printf("Created:  %s\n", formatDateTime(issue.Created))
	fmt.Printf("Updated:  %s\n", formatDateTime(issue.Updated))

	if issue.Description != "" {
		fmt.Println()
		fmt.Println("Description:")
		fmt.Print(RenderMarkdown(issue.Description))
	}

	if len(issue.Links) > 0 {
		fmt.Println()
		fmt.Println("Links:")
		for _, link := range issue.Links {
			summary := width.Truncate(link.Summary, 50, "...")
			fmt.Printf("  %s %s (%s) - %s\n", link.Direction, link.Key, link.Status, summary)
		}
	}

	if len(issue.Comments) > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("Comments (%d):\n", len(issue.Comments))
		for _, c := range issue.Comments {
			fmt.Println()
			fmt.Printf("[%s] [%s] %s:\n", formatDateTime(c.Created), c.ID, c.Author)
			fmt.Print(RenderMarkdown(c.Body))
		}
	}
}

// formatDateTime formats an ISO datetime string for display.
func formatDateTime(iso string) string {
	// Jira returns ISO 8601 format like "2024-01-16T14:30:00.000+0000"
	// Truncate to just date and time for readability
	if len(iso) >= 16 {
		return iso[:10] + " " + iso[11:16]
	}
	return iso
}

// getComments fetches recent comments for an issue.
func getComments(ctx context.Context, client *api.Client, key string, limit int) ([]CommentInfo, error) {
	path := fmt.Sprintf("/issue/%s/comment?maxResults=%d&orderBy=-created", key, limit)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp commentsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var comments []CommentInfo
	for _, c := range resp.Comments {
		info := CommentInfo{
			ID:      c.ID,
			Created: c.Created,
		}
		if c.Author != nil {
			info.Author = c.Author.DisplayName
		}
		// Convert comment body from ADF to Markdown
		if len(c.Body) > 0 && string(c.Body) != "null" {
			info.Body = converter.ADFToMarkdown(c.Body)
		}
		comments = append(comments, info)
	}

	return comments, nil
}
