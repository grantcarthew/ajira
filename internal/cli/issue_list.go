package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// IssueInfo represents a Jira issue for output.
type IssueInfo struct {
	Key      string `json:"key"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
	Type     string `json:"type"`
	Priority string `json:"priority"`
	Assignee string `json:"assignee"`
}

// issueSearchResponse matches the Jira issue search API response.
type issueSearchResponse struct {
	StartAt    int          `json:"startAt"`
	MaxResults int          `json:"maxResults"`
	Total      int          `json:"total"`
	Issues     []issueValue `json:"issues"`
}

type issueValue struct {
	Key    string      `json:"key"`
	Fields issueFields `json:"fields"`
}

type issueFields struct {
	Summary   string         `json:"summary"`
	Status    *statusField   `json:"status"`
	IssueType *issueType     `json:"issuetype"`
	Priority  *priorityField `json:"priority"`
	Assignee  *userField     `json:"assignee"`
}

type statusField struct {
	Name string `json:"name"`
}

type issueType struct {
	Name string `json:"name"`
}

type priorityField struct {
	Name string `json:"name"`
}

type userField struct {
	DisplayName string `json:"displayName"`
	AccountID   string `json:"accountId"`
	EmailAddress string `json:"emailAddress"`
}

var (
	issueListQuery    string
	issueListStatus   string
	issueListType     string
	issueListAssignee string
	issueListLimit    int
)

var issueListCmd = &cobra.Command{
	Use:           "list",
	Short:         "List and search issues",
	Long:          "List Jira issues using JQL or convenience filters.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runIssueList,
}

func init() {
	issueListCmd.Flags().StringVarP(&issueListQuery, "query", "q", "", "JQL query (overrides other filters)")
	issueListCmd.Flags().StringVarP(&issueListStatus, "status", "s", "", "Filter by status")
	issueListCmd.Flags().StringVarP(&issueListType, "type", "t", "", "Filter by issue type")
	issueListCmd.Flags().StringVarP(&issueListAssignee, "assignee", "a", "", "Filter by assignee (email, accountId, or 'unassigned')")
	issueListCmd.Flags().IntVarP(&issueListLimit, "limit", "l", 50, "Maximum issues to return")

	issueCmd.AddCommand(issueListCmd)
}

func runIssueList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	jql := buildJQL()
	if jql == "" {
		// Default: issues in current project if set
		if Project() != "" {
			jql = fmt.Sprintf("project = %s ORDER BY updated DESC", Project())
		} else {
			jql = "ORDER BY updated DESC"
		}
	}

	issues, err := searchIssues(client, jql, issueListLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 {
				return Errorf("authentication failed (401)")
			}
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to search issues: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(issues) == 0 {
			fmt.Println("No issues found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tSTATUS\tTYPE\tASSIGNEE\tSUMMARY")
		for _, issue := range issues {
			assignee := issue.Assignee
			if assignee == "" {
				assignee = "-"
			}
			// Truncate summary for display
			summary := issue.Summary
			if len(summary) > 60 {
				summary = summary[:57] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", issue.Key, issue.Status, issue.Type, assignee, summary)
		}
		w.Flush()
	}

	return nil
}

func buildJQL() string {
	// If raw query provided, use it directly
	if issueListQuery != "" {
		return issueListQuery
	}

	var conditions []string

	// Add project filter if set
	if Project() != "" {
		conditions = append(conditions, fmt.Sprintf("project = %s", Project()))
	}

	// Add convenience filters
	if issueListStatus != "" {
		conditions = append(conditions, fmt.Sprintf("status = \"%s\"", issueListStatus))
	}
	if issueListType != "" {
		conditions = append(conditions, fmt.Sprintf("issuetype = \"%s\"", issueListType))
	}
	if issueListAssignee != "" {
		if strings.ToLower(issueListAssignee) == "unassigned" {
			conditions = append(conditions, "assignee IS EMPTY")
		} else {
			conditions = append(conditions, fmt.Sprintf("assignee = \"%s\"", issueListAssignee))
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " AND ") + " ORDER BY updated DESC"
}

func searchIssues(client *api.Client, jql string, limit int) ([]IssueInfo, error) {
	var allIssues []IssueInfo
	startAt := 0
	maxResults := 50
	if limit > 0 && limit < maxResults {
		maxResults = limit
	}

	for {
		path := fmt.Sprintf("/search?jql=%s&startAt=%d&maxResults=%d&fields=summary,status,issuetype,priority,assignee",
			url.QueryEscape(jql), startAt, maxResults)

		body, err := client.Get(context.Background(), path)
		if err != nil {
			return nil, err
		}

		var resp issueSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, issue := range resp.Issues {
			info := IssueInfo{
				Key:     issue.Key,
				Summary: issue.Fields.Summary,
			}
			if issue.Fields.Status != nil {
				info.Status = issue.Fields.Status.Name
			}
			if issue.Fields.IssueType != nil {
				info.Type = issue.Fields.IssueType.Name
			}
			if issue.Fields.Priority != nil {
				info.Priority = issue.Fields.Priority.Name
			}
			if issue.Fields.Assignee != nil {
				info.Assignee = issue.Fields.Assignee.DisplayName
			}

			allIssues = append(allIssues, info)

			if limit > 0 && len(allIssues) >= limit {
				return allIssues[:limit], nil
			}
		}

		if startAt+len(resp.Issues) >= resp.Total {
			break
		}

		startAt += maxResults
	}

	return allIssues, nil
}
