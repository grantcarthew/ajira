package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"

	// Global flags
	jsonOutput bool
	project    string
)

var rootCmd = &cobra.Command{
	Use:   "ajira <command>",
	Short: "Atlassian Jira CLI for AI agents and automation",
	Long:  "Atlassian Jira CLI designed for AI agents and automation. Non-interactive, environment-configured, with Markdown input/output and JSON support.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set project from env if not specified via flag
		if project == "" {
			project = os.Getenv("JIRA_PROJECT")
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Default project key (or set JIRA_PROJECT)")
	rootCmd.Version = Version
}

func Execute() error {
	return rootCmd.Execute()
}

// JSONOutput returns true if JSON output is requested.
func JSONOutput() bool {
	return jsonOutput
}

// Project returns the current project key.
func Project() string {
	return project
}

// Errorf prints an error message to stderr and returns an error.
func Errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, "ajira:", msg)
	return fmt.Errorf("%s", msg)
}
