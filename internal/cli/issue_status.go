package cli

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

var issueStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"statuses"},
	Short:   "List available statuses",
	Long:    "List statuses available for the current project.",
	Example: `  ajira issue status           # List statuses for default project
  ajira issue status -p PROJ   # List statuses for specific project`,
	SilenceUsage: true,
	RunE:         runIssueStatus,
}

func init() {
	issueCmd.AddCommand(issueStatusCmd)
}

func runIssueStatus(cmd *cobra.Command, args []string) error {
	projectKey := Project()
	if projectKey == "" {
		return Errorf("project is required (use -p flag or set JIRA_PROJECT)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	statuses, err := jira.GetStatuses(client, projectKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch statuses: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(statuses, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printStatuses(statuses)
	}

	return nil
}

func printStatuses(statuses []jira.Status) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths
	nameWidth := 4 // "NAME"
	for _, s := range statuses {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}

	// Print header
	fmt.Printf("%s  %s\n",
		header(padRight("NAME", nameWidth)),
		header("CATEGORY"))

	// Print rows
	for _, s := range statuses {
		fmt.Printf("%s  %s\n", bold(padRight(s.Name, nameWidth)), s.Category)
	}
}
