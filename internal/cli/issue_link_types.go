package cli

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
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
	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	linkTypes, err := jira.GetLinkTypes(client)
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

	// Calculate column widths
	nameWidth := 4    // "NAME"
	outwardWidth := 7 // "OUTWARD"
	for _, lt := range linkTypes {
		if len(lt.Name) > nameWidth {
			nameWidth = len(lt.Name)
		}
		if len(lt.Outward) > outwardWidth {
			outwardWidth = len(lt.Outward)
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
