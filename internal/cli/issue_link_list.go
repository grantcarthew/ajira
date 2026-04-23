package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/grantcarthew/ajira/internal/api"
	"github.com/grantcarthew/ajira/internal/config"
	"github.com/grantcarthew/ajira/internal/width"
	"github.com/spf13/cobra"
)

var issueLinkListCmd = &cobra.Command{
	Use:   "list <issue-key>",
	Short: "List links",
	Long:  "List all issue links with direction, key, status, and summary.",
	Example: `  ajira issue link list PROJ-123          # List all links
  ajira issue link list PROJ-123 --json  # JSON output`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueLinkList,
}

func init() {
	issueLinkCmd.AddCommand(issueLinkListCmd)
}

func runIssueLinkList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	links, err := getIssueLinks(ctx, client, issueKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch links: %w", err)
	}

	infos := linksToLinkInfos(links)

	if JSONOutput() {
		output, err := json.MarshalIndent(infos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		printLinkList(issueKey, infos)
	}

	return nil
}

func printLinkList(issueKey string, links []LinkInfo) {
	if len(links) == 0 {
		fmt.Printf("No links for %s\n", issueKey)
		return
	}

	fmt.Printf("Links for %s (%d):\n\n", issueKey, len(links))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Direction\tKey\tStatus\tSummary")
	for _, l := range links {
		summary := width.Truncate(l.Summary, 50, "...")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			l.Direction,
			l.Key,
			l.Status,
			summary,
		)
	}
	w.Flush()
}
