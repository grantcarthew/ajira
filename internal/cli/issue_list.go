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
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/gcarthew/ajira/internal/width"
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
	issueListReporter string
	issueListPriority string
	issueListLabels   []string
	issueListWatching bool
	issueListOrderBy  string
	issueListReverse  bool
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
	issueListCmd.Flags().StringVarP(&issueListReporter, "reporter", "r", "", "Filter by reporter (email, accountId, or 'me')")
	issueListCmd.Flags().StringVarP(&issueListPriority, "priority", "P", "", "Filter by priority")
	issueListCmd.Flags().StringSliceVarP(&issueListLabels, "labels", "L", nil, "Filter by labels (comma-separated)")
	issueListCmd.Flags().BoolVarP(&issueListWatching, "watching", "w", false, "Filter to issues you are watching")
	issueListCmd.Flags().StringVar(&issueListOrderBy, "order-by", "", "Sort field (created, updated, priority, key, rank)")
	issueListCmd.Flags().BoolVar(&issueListReverse, "reverse", false, "Reverse sort order (ASC instead of DESC)")
	issueListCmd.Flags().IntVarP(&issueListLimit, "limit", "l", 50, "Maximum issues to return")

	issueCmd.AddCommand(issueListCmd)
}

func runIssueList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Validate filter values before building JQL
	if err := jira.ValidatePriority(ctx, client, issueListPriority); err != nil {
		return fmt.Errorf("%v", err)
	}
	if issueListStatus != "" && Project() == "" {
		return fmt.Errorf("--status requires a project; use -p flag or set JIRA_PROJECT")
	}
	if err := jira.ValidateStatus(ctx, client, Project(), issueListStatus); err != nil {
		return fmt.Errorf("%v", err)
	}
	if issueListType != "" && Project() == "" {
		return fmt.Errorf("--type requires a project; use -p flag or set JIRA_PROJECT")
	}
	if err := jira.ValidateIssueType(ctx, client, Project(), issueListType); err != nil {
		return fmt.Errorf("%v", err)
	}

	jql := buildJQL()
	if jql == "" {
		// Default: issues in current project if set
		if Project() != "" {
			jql = fmt.Sprintf("project = %s ORDER BY updated DESC", Project())
		} else {
			jql = "ORDER BY updated DESC"
		}
	}

	issues, err := searchIssues(ctx, client, jql, issueListLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("Failed to search issues: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return fmt.Errorf("Failed to format JSON: %v", err)
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

		// Calculate column widths using display width for Unicode support
		keyWidth, statusWidth, typeWidth, assigneeWidth := 8, 11, 4, 8
		for _, issue := range issues {
			if w := width.StringWidth(issue.Key); w > keyWidth {
				keyWidth = w
			}
			if w := width.StringWidth(issue.Status); w > statusWidth {
				statusWidth = w
			}
			if w := width.StringWidth(issue.Type); w > typeWidth {
				typeWidth = w
			}
			assignee := issue.Assignee
			if assignee == "" {
				assignee = "-"
			}
			if w := width.StringWidth(assignee); w > assigneeWidth {
				assigneeWidth = w
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

			// Truncate summary for display using display width
			summary := width.Truncate(issue.Summary, 60, "...")

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
	if issueListReporter != "" {
		if strings.ToLower(issueListReporter) == "me" {
			conditions = append(conditions, "reporter = currentUser()")
		} else {
			conditions = append(conditions, fmt.Sprintf("reporter = \"%s\"", issueListReporter))
		}
	}
	if issueListPriority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = \"%s\"", issueListPriority))
	}
	if len(issueListLabels) > 0 {
		quoted := make([]string, len(issueListLabels))
		for i, label := range issueListLabels {
			quoted[i] = fmt.Sprintf("\"%s\"", label)
		}
		conditions = append(conditions, fmt.Sprintf("labels IN (%s)", strings.Join(quoted, ", ")))
	}
	if issueListWatching {
		conditions = append(conditions, "watcher = currentUser()")
	}

	if len(conditions) == 0 {
		return ""
	}

	// Build ORDER BY clause
	orderBy := buildOrderBy()

	return strings.Join(conditions, " AND ") + orderBy
}

// buildOrderBy constructs the ORDER BY clause based on flags.
func buildOrderBy() string {
	field := issueListOrderBy
	if field == "" {
		field = "updated"
	}

	direction := "DESC"
	if issueListReverse {
		direction = "ASC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", field, direction)
}

func searchIssues(ctx context.Context, client *api.Client, jql string, limit int) ([]IssueInfo, error) {
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

		body, err := client.Get(ctx, path)
		if err != nil {
			return nil, err
		}

		var resp issueSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("Failed to parse response: %w", err)
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

	// Override for specific status names (TrimSpace handles padded input)
	switch strings.ToLower(strings.TrimSpace(status)) {
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

// padRight pads a string to the specified display width with spaces.
// Uses display width calculation to handle Unicode characters correctly.
func padRight(s string, w int) string {
	sw := width.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}
