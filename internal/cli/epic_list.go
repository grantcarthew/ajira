package cli

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/gcarthew/ajira/internal/width"
	"github.com/spf13/cobra"
)

var (
	epicListStatus   string
	epicListAssignee string
	epicListPriority string
	epicListLimit    int
)

var epicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List epics",
	Long:  "List epics in the project. Requires -p or JIRA_PROJECT.",
	Example: `  ajira epic list                        # List epics in default project
  ajira epic list -p GCP                 # List epics in specific project
  ajira epic list --status "In Progress" # Filter by status
  ajira epic list -l 10                  # Limit results`,
	SilenceUsage: true,
	RunE:         runEpicList,
}

func init() {
	epicListCmd.Flags().StringVar(&epicListStatus, "status", "", "Filter by status")
	epicListCmd.Flags().StringVarP(&epicListAssignee, "assignee", "a", "", "Filter by assignee (email, accountId, 'me', or 'unassigned')")
	epicListCmd.Flags().StringVarP(&epicListPriority, "priority", "P", "", "Filter by priority")
	epicListCmd.Flags().IntVarP(&epicListLimit, "limit", "l", 50, "Maximum epics to return")

	epicCmd.AddCommand(epicListCmd)
}

func runEpicList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if Project() == "" {
		return fmt.Errorf("project required; use -p flag or set JIRA_PROJECT")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Validate filter values
	if err := jira.ValidatePriority(ctx, client, epicListPriority); err != nil {
		return fmt.Errorf("%v", err)
	}
	if err := jira.ValidateStatus(ctx, client, Project(), epicListStatus); err != nil {
		return fmt.Errorf("%v", err)
	}

	jql := buildEpicListJQL()

	issues, err := searchIssues(ctx, client, jql, epicListLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to search epics: %w", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(issues, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(issues) == 0 {
			fmt.Println("No epics found.")
			return nil
		}

		bold := color.New(color.Bold).SprintFunc()
		faint := color.New(color.Faint).SprintFunc()
		header := color.New(color.FgCyan, color.Bold).SprintFunc()

		// Calculate column widths
		keyWidth, statusWidth, priorityWidth, assigneeWidth := 8, 11, 8, 8
		for _, issue := range issues {
			if w := width.StringWidth(issue.Key); w > keyWidth {
				keyWidth = w
			}
			if w := width.StringWidth(issue.Status); w > statusWidth {
				statusWidth = w
			}
			if w := width.StringWidth(issue.Priority); w > priorityWidth {
				priorityWidth = w
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
			header(padRight("PRIORITY", priorityWidth)),
			header(padRight("ASSIGNEE", assigneeWidth)),
			header("SUMMARY"))

		// Print rows
		for _, issue := range issues {
			key := bold(padRight(issue.Key, keyWidth))
			status := colorStatus(padRight(issue.Status, statusWidth), issue.StatusCategory)
			priority := padRight(issue.Priority, priorityWidth)

			assignee := issue.Assignee
			if assignee == "" {
				assignee = faint(padRight("-", assigneeWidth))
			} else {
				assignee = padRight(assignee, assigneeWidth)
			}

			summary := width.Truncate(issue.Summary, 60, "...")

			fmt.Printf("%s  %s  %s  %s  %s\n", key, status, priority, assignee, summary)
		}
	}

	return nil
}

func buildEpicListJQL() string {
	conditions := []string{
		fmt.Sprintf("project = %s", Project()),
		"issuetype = Epic",
	}

	if epicListStatus != "" {
		conditions = append(conditions, fmt.Sprintf("status = \"%s\"", epicListStatus))
	}
	if epicListAssignee != "" {
		switch epicListAssignee {
		case "unassigned":
			conditions = append(conditions, "assignee IS EMPTY")
		case "me":
			conditions = append(conditions, "assignee = currentUser()")
		default:
			conditions = append(conditions, fmt.Sprintf("assignee = \"%s\"", epicListAssignee))
		}
	}
	if epicListPriority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = \"%s\"", epicListPriority))
	}

	jql := ""
	for i, cond := range conditions {
		if i > 0 {
			jql += " AND "
		}
		jql += cond
	}
	jql += " ORDER BY updated DESC"

	return jql
}
