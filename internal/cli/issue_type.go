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

var issueTypeCmd = &cobra.Command{
	Use:   "type",
	Short: "List available issue types",
	Long:  "List issue types available for the current project.",
	Example: `  ajira issue type           # List types for default project
  ajira issue type -p PROJ   # List types for specific project`,
	SilenceUsage: true,
	RunE:         runIssueType,
}

func init() {
	issueCmd.AddCommand(issueTypeCmd)
}

func runIssueType(cmd *cobra.Command, args []string) error {
	projectKey := Project()
	if projectKey == "" {
		return Errorf("project is required (use -p flag or set JIRA_PROJECT)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	types, err := jira.GetIssueTypes(client, projectKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch issue types: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(types, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printIssueTypes(types)
	}

	return nil
}

func printIssueTypes(types []jira.IssueType) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths
	nameWidth := 4 // "NAME"
	for _, t := range types {
		if len(t.Name) > nameWidth {
			nameWidth = len(t.Name)
		}
	}

	// Print header
	fmt.Printf("%s  %s\n",
		header(padRight("NAME", nameWidth)),
		header("DESCRIPTION"))

	// Print rows
	for _, t := range types {
		desc := t.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("%s  %s\n", bold(padRight(t.Name, nameWidth)), desc)
	}
}
