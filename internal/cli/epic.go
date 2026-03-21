package cli

import (
	"github.com/spf13/cobra"
)

var epicCmd = &cobra.Command{
	Use:     "epic",
	Aliases: []string{"epics"},
	Short:   "Manage epics",
	Long:    "Commands for managing Jira epics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(epicCmd)
}
