package cli

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/gcarthew/ajira/internal/width"
	"github.com/spf13/cobra"
)

var issueLinkTypesCmd = &cobra.Command{
	Use:     "types",
	Aliases: []string{"type"},
	Short:   "List available link types",
	Long:    "List all issue link types available in the Jira instance.",
	Example: `  ajira issue link types         # List all link types
  ajira issue link types --json  # JSON output`,
	SilenceUsage: true,
	RunE:         runIssueLinkTypes,
}

func init() {
	issueLinkCmd.AddCommand(issueLinkTypesCmd)
}

func runIssueLinkTypes(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	linkTypes, err := jira.GetLinkTypes(ctx, client)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch link types: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(linkTypes, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printLinkTypes(linkTypes)
	}

	return nil
}

func printLinkTypes(linkTypes []jira.LinkType) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths using display width for Unicode support
	nameWidth := 4    // "NAME"
	outwardWidth := 7 // "OUTWARD"
	for _, lt := range linkTypes {
		if w := width.StringWidth(lt.Name); w > nameWidth {
			nameWidth = w
		}
		if w := width.StringWidth(lt.Outward); w > outwardWidth {
			outwardWidth = w
		}
	}

	// Print header
	fmt.Printf("%s  %s  %s\n",
		header(padRight("NAME", nameWidth)),
		header(padRight("OUTWARD", outwardWidth)),
		header("INWARD"))

	// Print rows
	for _, lt := range linkTypes {
		fmt.Printf("%s  %s  %s\n",
			bold(padRight(lt.Name, nameWidth)),
			padRight(lt.Outward, outwardWidth),
			lt.Inward)
	}
}
