package cli

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed help/agents.md
var agentsHelp string

//go:embed help/schemas.md
var schemasHelp string

//go:embed help/markdown.md
var markdownHelp string

//go:embed help/agile.md
var agileHelp string

var helpCmd = &cobra.Command{
	Use:          "help [command]",
	Short:        "Help for commands and topics",
	Long:         "Help provides help for commands and topics (agents, agile, markdown, schemas).",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return rootCmd.Help()
		}
		target, _, err := rootCmd.Find(args)
		if err != nil || target == nil {
			return NewExitError(ExitUserError, fmt.Errorf("unknown help topic: %s", args[0]))
		}
		return target.Help()
	},
}

var helpAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "AI agent reference",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(RenderMarkdown(agentsHelp))
	},
}

var helpSchemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "JSON output schemas",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(RenderMarkdown(schemasHelp))
	},
}

var helpMarkdownCmd = &cobra.Command{
	Use:   "markdown",
	Short: "Markdown formatting reference",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(RenderMarkdown(markdownHelp))
	},
}

var helpAgileCmd = &cobra.Command{
	Use:   "agile",
	Short: "Agile commands reference (epics, sprints, boards)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(RenderMarkdown(agileHelp))
	},
}

func init() {
	rootCmd.SetHelpCommand(helpCmd)
	helpCmd.AddCommand(helpAgentsCmd)
	helpCmd.AddCommand(helpSchemasCmd)
	helpCmd.AddCommand(helpMarkdownCmd)
	helpCmd.AddCommand(helpAgileCmd)
}
