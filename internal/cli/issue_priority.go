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


var issuePriorityCmd = &cobra.Command{
	Use:           "priority",
	Short:         "List available priorities",
	Long:          "List all priorities available in the Jira instance.",
	Example: `  ajira issue priority         # List all priorities
  ajira issue priority --json  # JSON output`,
	SilenceUsage:  true,
	RunE:          runIssuePriority,
}

func init() {
	issueCmd.AddCommand(issuePriorityCmd)
}

func runIssuePriority(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	priorities, err := jira.GetPriorities(client)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch priorities: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(priorities, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printPriorities(priorities)
	}

	return nil
}

func printPriorities(priorities []jira.Priority) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths
	nameWidth := 4 // "NAME"
	for _, p := range priorities {
		if len(p.Name) > nameWidth {
			nameWidth = len(p.Name)
		}
	}

	// Print header
	fmt.Printf("%s  %s\n",
		header(padRight("NAME", nameWidth)),
		header("DESCRIPTION"))

	// Print rows
	for _, p := range priorities {
		desc := p.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("%s  %s\n", bold(padRight(p.Name, nameWidth)), desc)
	}
}
