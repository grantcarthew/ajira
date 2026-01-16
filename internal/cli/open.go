package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var openCmd = &cobra.Command{
	Use:   "open [issue-key]",
	Short: "Open in browser",
	Long:  "Open the project or an issue in your default browser. Prints URL if not in TTY.",
	Example: `  ajira open                    # Open project in browser
  ajira open PROJ-123           # Open issue in browser
  ajira open -p PROJ            # Open specific project`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE:         runOpen,
}

var issueOpenCmd = &cobra.Command{
	Use:          "open <issue-key>",
	Short:        "Open issue",
	Long:         "Open an issue in your default browser. Prints URL if not in TTY.",
	Example:      `  ajira issue open PROJ-123     # Open issue in browser`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
	issueCmd.AddCommand(issueOpenCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var url string
	if len(args) == 1 {
		// Open specific issue
		url = IssueURL(cfg.BaseURL, args[0])
	} else {
		// Open project
		projectKey := Project()
		if projectKey == "" {
			return fmt.Errorf("project is required: use -p flag or set JIRA_PROJECT environment variable")
		}
		url = ProjectURL(cfg.BaseURL, projectKey)
	}

	return openURL(url)
}

func runIssueOpen(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	url := IssueURL(cfg.BaseURL, args[0])
	return openURL(url)
}

// ProjectURL returns the browse URL for a project key.
func ProjectURL(baseURL, projectKey string) string {
	return fmt.Sprintf("%s/browse/%s", baseURL, projectKey)
}

// openURL opens a URL in the default browser or prints it if not in a TTY.
func openURL(url string) error {
	// Check if we're in a TTY
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println(url)
		return nil
	}

	// Dry-run mode
	if DryRun() {
		PrintDryRun(fmt.Sprintf("open %s", url))
		return nil
	}

	// Try to open browser
	err := openBrowser(url)
	if err != nil {
		// Fall back to printing URL
		fmt.Println(url)
		return nil
	}

	if !Quiet() {
		fmt.Println(url)
	}
	return nil
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
