package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/glamour"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"

	// Global flags
	jsonOutput bool
	project    string
	board      string

	// Automation flags
	dryRun  bool
	verbose bool
	quiet   bool
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:   "ajira <command> [flags]",
	Short: "Atlassian Jira CLI for AI agents and automation",
	Long: `Atlassian Jira CLI designed for AI agents and automation.
Non-interactive, environment-configured, with Markdown input/output and JSON support.

Environment Variables:
  ATLASSIAN_BASE_URL   Atlassian instance URL (shared with acon)
  ATLASSIAN_EMAIL      User email (shared with acon)
  ATLASSIAN_API_TOKEN  API token (shared with acon)
  JIRA_BASE_URL        Jira URL (overrides ATLASSIAN_BASE_URL)
  JIRA_EMAIL           User email (overrides ATLASSIAN_EMAIL)
  JIRA_API_TOKEN       API token (overrides ATLASSIAN_API_TOKEN)
  JIRA_PROJECT         Default project key (optional)
  JIRA_BOARD           Default board ID (optional)

Global Flags (work with most commands):
  --json       Output in JSON format for parsing
  --dry-run    Preview actions without executing
  --quiet      Suppress non-essential output
  --no-color   Disable coloured output
  --verbose    Show HTTP request/response details
  -p, --project   Override default project
  --board      Override default board ID

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
		// Set board from env if not specified via flag
		if board == "" {
			board = os.Getenv("JIRA_BOARD")
		}
		// Enable verbose HTTP logging if requested
		if verbose {
			api.SetVerboseOutput(os.Stderr)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Default project key (or set JIRA_PROJECT)")
	rootCmd.PersistentFlags().StringVar(&board, "board", "", "Default board ID for agile commands (or set JIRA_BOARD)")

	// Automation flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show planned actions without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show HTTP request/response details")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable coloured output")

	// Disable Cobra's verbose completion command, we'll add our own
	rootCmd.CompletionOptions.DisableDefaultCmd = true

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

// Board returns the current board ID.
func Board() string {
	return board
}

// DryRun returns true if dry-run mode is enabled.
func DryRun() bool {
	return dryRun
}

// Verbose returns true if verbose mode is enabled.
func Verbose() bool {
	return verbose
}

// Quiet returns true if quiet mode is enabled.
func Quiet() bool {
	return quiet
}

// NoColor returns true if colour output is disabled.
func NoColor() bool {
	return noColor
}

// IssueURL returns the browse URL for an issue key.
func IssueURL(baseURL, key string) string {
	return fmt.Sprintf("%s/browse/%s", baseURL, key)
}

// RenderMarkdown renders markdown with terminal styling.
// Falls back to plain text if rendering fails, output is not a TTY, or --no-color is set.
func RenderMarkdown(markdown string) string {
	if markdown == "" {
		return ""
	}

	// Check if colour is disabled or stdout is not a terminal
	if noColor || !term.IsTerminal(int(os.Stdout.Fd())) {
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
