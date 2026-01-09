package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

// LinkResult represents the result of creating a link.
type LinkResult struct {
	OutwardIssue string `json:"outwardIssue"`
	InwardIssue  string `json:"inwardIssue"`
	Type         string `json:"type"`
}

// issueLinkRequest represents the request body for creating an issue link.
type issueLinkRequest struct {
	OutwardIssue issueRef    `json:"outwardIssue"`
	InwardIssue  issueRef    `json:"inwardIssue"`
	Type         linkTypeRef `json:"type"`
}

type issueRef struct {
	Key string `json:"key"`
}

type linkTypeRef struct {
	Name string `json:"name"`
}

var issueLinkAddCmd = &cobra.Command{
	Use:   "add <outward-key> <link-type> <inward-key>",
	Short: "Create a link between two issues",
	Long: `Create a directional link between two issues.

The command reads as a sentence: "KEY1 blocks KEY2" means KEY1 is the
outward issue that blocks KEY2 (the inward issue).`,
	Example: `  ajira issue link add GCP-123 Blocks GCP-456      # GCP-123 blocks GCP-456
  ajira issue link add GCP-100 Duplicate GCP-200   # GCP-100 duplicates GCP-200
  ajira issue link add GCP-50 "Related Issues" GCP-60`,
	Args:         cobra.ExactArgs(3),
	SilenceUsage: true,
	RunE:         runIssueLinkAdd,
}

func init() {
	issueLinkCmd.AddCommand(issueLinkAddCmd)
}

func runIssueLinkAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	outwardKey := args[0]
	linkTypeName := args[1]
	inwardKey := args[2]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Validate link type
	linkTypes, err := jira.GetLinkTypes(ctx, client)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("Failed to fetch link types: %v", err)
	}

	validType := findLinkType(linkTypes, linkTypeName)
	if validType == nil {
		var available []string
		for _, lt := range linkTypes {
			available = append(available, lt.Name)
		}
		return fmt.Errorf("Link type not found: %s (available: %s)", linkTypeName, strings.Join(available, ", "))
	}

	// Create the link
	err = createIssueLink(ctx, client, outwardKey, inwardKey, validType.Name)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("Failed to create link: %v", err)
	}

	if JSONOutput() {
		result := LinkResult{
			OutwardIssue: outwardKey,
			InwardIssue:  inwardKey,
			Type:         validType.Name,
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, outwardKey))
	}

	return nil
}

// findLinkType finds a link type by name (case-insensitive).
func findLinkType(linkTypes []jira.LinkType, name string) *jira.LinkType {
	for _, lt := range linkTypes {
		if strings.EqualFold(lt.Name, name) {
			return &lt
		}
	}
	return nil
}

func createIssueLink(ctx context.Context, client *api.Client, outwardKey, inwardKey, linkTypeName string) error {
	// Jira API naming is counterintuitive:
	// - inwardIssue = the issue that "does" the action (shows outward text, e.g., "blocks")
	// - outwardIssue = the issue that "receives" the action (shows inward text, e.g., "is blocked by")
	// So we swap them to match our command syntax: KEY1 TYPE KEY2 = "KEY1 blocks KEY2"
	req := issueLinkRequest{
		OutwardIssue: issueRef{Key: inwardKey},
		InwardIssue:  issueRef{Key: outwardKey},
		Type:         linkTypeRef{Name: linkTypeName},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Failed to marshal request: %w", err)
	}

	_, err = client.Post(ctx, "/issueLink", body)
	return err
}
