package cli

import (
	"github.com/spf13/cobra"
)

var issueLinkCmd = &cobra.Command{
	Use:     "link",
	Aliases: []string{"links"},
	Short:   "Manage issue links",
	Long:    "Commands for managing links between Jira issues and external URLs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	issueCmd.AddCommand(issueLinkCmd)
}
