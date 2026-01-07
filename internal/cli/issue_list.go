package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// IssueInfo represents a Jira issue for output.
type IssueInfo struct {
	Key            string `json:"key"`
	Summary        string `json:"summary"`
	Status         string `json:"status"`
	StatusCategory string `json:"statusCategory"`
	Type           string `json:"type"`
	Priority       string `json:"priority"`
	Assignee       string `json:"assignee"`
}

// issueSearchResponse matches the Jira issue search API response.
type issueSearchResponse struct {
	Issues        []issueValue `json:"issues"`
	NextPageToken string       `json:"nextPageToken"`
	IsLast        bool         `json:"isLast"`
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
	Name           string          `json:"name"`
	StatusCategory *statusCategory `json:"statusCategory"`
}

type statusCategory struct {
	Key string `json:"key"`
}

type issueType struct {
	Name string `json:"name"`
}

type priorityField struct {
	Name string `json:"name"`
}

type userField struct {
	DisplayName  string `json:"displayName"`
	AccountID    string `json:"accountId"`
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
	Use:   "list",
	Short: "List and search issues",
	Long:  "List Jira issues using JQL or convenience filters.",
	Example: `  ajira issue list                        # List issues in default project
  ajira issue list -s "In Progress"       # Filter by status
  ajira issue list -a me -t Bug           # My bugs
  ajira issue list -q "updated >= -7d"    # JQL query`,
	SilenceUsage: true,
	RunE:         runIssueList,
}

func init() {
	issueListCmd.Flags().StringVarP(&issueListQuery, "query", "q", "", "JQL query (overrides other filters)")
	issueListCmd.Flags().StringVarP(&issueListStatus, "status", "s", "", "Filter by status")
	issueListCmd.Flags().StringVarP(&issueListType, "type", "t", "", "Filter by issue type")
	issueListCmd.Flags().StringVarP(&issueListAssignee, "assignee", "a", "", "Filter by assignee (email, accountId, 'me', or 'unassigned')")
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

		bold := color.New(color.Bold).SprintFunc()
		faint := color.New(color.Faint).SprintFunc()
		header := color.New(color.FgCyan, color.Bold).SprintFunc()

		// Calculate column widths
		keyWidth, statusWidth, typeWidth, assigneeWidth := 8, 11, 4, 8
		for _, issue := range issues {
			if len(issue.Key) > keyWidth {
				keyWidth = len(issue.Key)
			}
			if len(issue.Status) > statusWidth {
				statusWidth = len(issue.Status)
			}
			if len(issue.Type) > typeWidth {
				typeWidth = len(issue.Type)
			}
			assignee := issue.Assignee
			if assignee == "" {
				assignee = "-"
			}
			if len(assignee) > assigneeWidth {
				assigneeWidth = len(assignee)
			}
		}

		// Print header
		fmt.Printf("%s  %s  %s  %s  %s\n",
			header(padRight("KEY", keyWidth)),
			header(padRight("STATUS", statusWidth)),
			header(padRight("TYPE", typeWidth)),
			header(padRight("ASSIGNEE", assigneeWidth)),
			header("SUMMARY"))

		// Print rows
		for _, issue := range issues {
			key := bold(padRight(issue.Key, keyWidth))
			status := colorStatus(padRight(issue.Status, statusWidth), issue.StatusCategory)

			assignee := issue.Assignee
			if assignee == "" {
				assignee = faint(padRight("-", assigneeWidth))
			} else {
				assignee = padRight(assignee, assigneeWidth)
			}

			// Truncate summary for display
			summary := issue.Summary
			if len(summary) > 60 {
				summary = summary[:57] + "..."
			}

			fmt.Printf("%s  %s  %s  %s  %s\n", key, status, padRight(issue.Type, typeWidth), assignee, summary)
		}
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
		switch strings.ToLower(issueListAssignee) {
		case "unassigned":
			conditions = append(conditions, "assignee IS EMPTY")
		case "me":
			conditions = append(conditions, "assignee = currentUser()")
		default:
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
	maxResults := 50
	if limit > 0 && limit < maxResults {
		maxResults = limit
	}

	nextPageToken := ""
	const maxPages = 100 // Safety guard against infinite pagination loops

	for page := 0; page < maxPages; page++ {
		path := fmt.Sprintf("/search/jql?jql=%s&maxResults=%d&fields=summary,status,issuetype,priority,assignee",
			url.QueryEscape(jql), maxResults)
		if nextPageToken != "" {
			path += "&nextPageToken=" + url.QueryEscape(nextPageToken)
		}

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
				if issue.Fields.Status.StatusCategory != nil {
					info.StatusCategory = issue.Fields.Status.StatusCategory.Key
				}
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

		if resp.IsLast || resp.NextPageToken == "" {
			break
		}

		nextPageToken = resp.NextPageToken
	}

	return allIssues, nil
}

// colorStatus returns a colored status string based on status category.
func colorStatus(status, category string) string {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()

	// Override for specific status names
	switch strings.ToLower(status) {
	case "blocked", "on hold", "on-hold", "awaiting":
		return yellow(status)
	}

	// Fall back to category-based coloring
	switch category {
	case "done":
		return green(status)
	case "indeterminate":
		return color.BlueString(status)
	case "new":
		return faint(status)
	default:
		return status
	}
}

// colorPriority returns a colored priority string.
func colorPriority(priority string) string {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	switch strings.ToLower(priority) {
	case "highest", "critical", "blocker":
		return red(priority)
	case "high":
		return red(priority)
	case "medium":
		return yellow(priority)
	default:
		return priority
	}
}

// padRight pads a string to the specified width with spaces.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
