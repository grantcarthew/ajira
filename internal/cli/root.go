package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

// Errorf returns a formatted error.
func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// IssueURL returns the browse URL for an issue key.
func IssueURL(baseURL, key string) string {
	return fmt.Sprintf("%s/browse/%s", baseURL, key)
}

// RenderMarkdown renders markdown with terminal styling.
// Falls back to plain text if rendering fails or output is not a TTY.
func RenderMarkdown(markdown string) string {
	if markdown == "" {
		return ""
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return markdown
	}

	// Get terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		width = 80
	}

	// Create renderer with auto style and terminal width
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return markdown
	}

	out, err := r.Render(markdown)
	if err != nil {
		return markdown
	}

	return out
}
