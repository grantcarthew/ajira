package cli

import (
	"github.com/spf13/cobra"
)

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
