package cli

import (
	"github.com/spf13/cobra"
)

var epicCmd = &cobra.Command{
	Use:   "epic",
	Short: "Manage Jira epics",
	Long:  "Commands for managing Jira epics: list, create, add, remove.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(epicCmd)
}
