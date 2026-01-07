package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// extractProjectKey extracts the project key from an issue key (e.g., "GCP-123" -> "GCP").
func extractProjectKey(issueKey string) string {
	if idx := strings.Index(issueKey, "-"); idx > 0 {
		return issueKey[:idx]
	}
	return issueKey
}

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage Jira issues",
	Long:  "Commands for managing Jira issues: list, view, create, edit, delete, assign, and move.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)
}
