package cli

import (
	"github.com/spf13/cobra"
)

var issueLinkCmd = &cobra.Command{
	Use:     "link",
	Aliases: []string{"links"},
	Short:   "Manage links",
	Long:    "Commands for managing issue links and remote URLs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	issueCmd.AddCommand(issueLinkCmd)
}
