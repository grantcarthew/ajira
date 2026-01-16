package cli

import (
	"github.com/spf13/cobra"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage sprints",
	Long:  "Commands for managing Jira sprints.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(sprintCmd)
}
