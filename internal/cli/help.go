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

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Help for commands and topics",
	Long:  "Help provides help for commands and topics (agents, markdown, schemas).",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = rootCmd.Help()
			return
		}
		// Find the command and show its help
		target, _, err := rootCmd.Find(args)
		if err != nil || target == nil {
			fmt.Printf("Unknown help topic: %s\n", args[0])
			return
		}
		_ = target.Help()
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

func init() {
	rootCmd.SetHelpCommand(helpCmd)
	helpCmd.AddCommand(helpAgentsCmd)
	helpCmd.AddCommand(helpSchemasCmd)
	helpCmd.AddCommand(helpMarkdownCmd)
}
