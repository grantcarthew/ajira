package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	Use:   "ajira <command> [flags]",
	Short: "Atlassian Jira CLI for AI agents and automation",
	Long: `Atlassian Jira CLI designed for AI agents and automation.
Non-interactive, environment-configured, with Markdown input/output and JSON support.

Environment Variables:
  JIRA_BASE_URL    Jira instance URL (required)
  JIRA_EMAIL       User email for authentication (required)
  JIRA_API_TOKEN   API token for authentication (required)
  JIRA_PROJECT     Default project key (optional)

Quick Start:
  ajira me                          Verify authentication
  ajira project list                List accessible projects
  ajira issue list -p PROJECT       List issues in project
  ajira issue view KEY              View issue details
  ajira issue create -p PROJECT -s "Summary" -t Task    Create issue

AI Agents: Run "ajira help agents" for a token-efficient reference.`,
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
	rootCmd.SetVersionTemplate(`ajira version {{.Version}}
Repository: https://github.com/grantcarthew/ajira
Report issues: https://github.com/grantcarthew/ajira/issues/new
`)

	// Custom usage template to avoid duplicate usage lines
	rootCmd.SetUsageTemplate(`Usage:
  {{.UseLine}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} <command> --help" for more information about a command.{{end}}
`)
}

func Execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
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
