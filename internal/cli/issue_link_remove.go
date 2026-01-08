package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// issueLinksResponse represents the response for fetching issue links.
type issueLinksResponse struct {
	Fields struct {
		IssueLinks []issueLink `json:"issuelinks"`
	} `json:"fields"`
}

type issueLink struct {
	ID           string        `json:"id"`
	Type         issueLinkType `json:"type"`
	InwardIssue  *linkedIssue  `json:"inwardIssue,omitempty"`
	OutwardIssue *linkedIssue  `json:"outwardIssue,omitempty"`
}

type issueLinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

type linkedIssue struct {
	Key    string            `json:"key"`
	Fields linkedIssueFields `json:"fields"`
}

type linkedIssueFields struct {
	Summary   string         `json:"summary"`
	Status    *statusField   `json:"status"`
	IssueType *issueType     `json:"issuetype"`
	Priority  *priorityField `json:"priority"`
}

// RemoveResult represents the result of removing links.
type RemoveResult struct {
	Issue1       string `json:"issue1"`
	Issue2       string `json:"issue2"`
	LinksRemoved int    `json:"linksRemoved"`
}

var issueLinkRemoveCmd = &cobra.Command{
	Use:   "remove <key1> <key2>",
	Short: "Remove all links between two issues",
	Long:  "Remove all links between two issues regardless of link type.",
	Example: `  ajira issue link remove GCP-123 GCP-456    # Remove all links between issues
  ajira issue link remove GCP-100 GCP-200`,
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE:         runIssueLinkRemove,
}

func init() {
	issueLinkCmd.AddCommand(issueLinkRemoveCmd)
}

func runIssueLinkRemove(cmd *cobra.Command, args []string) error {
	key1 := args[0]
	key2 := args[1]

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Fetch links from key1
	links, err := getIssueLinks(client, key1)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch issue links: %v", err)
	}

	// Find links connecting to key2
	var linksToRemove []string
	for _, link := range links {
		if link.InwardIssue != nil && link.InwardIssue.Key == key2 {
			linksToRemove = append(linksToRemove, link.ID)
		}
		if link.OutwardIssue != nil && link.OutwardIssue.Key == key2 {
			linksToRemove = append(linksToRemove, link.ID)
		}
	}

	if len(linksToRemove) == 0 {
		return Errorf("no links found between %s and %s", key1, key2)
	}

	// Delete each link
	for _, linkID := range linksToRemove {
		err := deleteIssueLink(client, linkID)
		if err != nil {
			if apiErr, ok := err.(*api.APIError); ok {
				return Errorf("API error - %v", apiErr)
			}
			return Errorf("failed to delete link %s: %v", linkID, err)
		}
	}

	if JSONOutput() {
		result := RemoveResult{
			Issue1:       key1,
			Issue2:       key2,
			LinksRemoved: len(linksToRemove),
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, key1))
	}

	return nil
}

func getIssueLinks(client *api.Client, key string) ([]issueLink, error) {
	path := fmt.Sprintf("/issue/%s?fields=issuelinks", key)

	body, err := client.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var resp issueLinksResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.Fields.IssueLinks, nil
}

func deleteIssueLink(client *api.Client, linkID string) error {
	path := fmt.Sprintf("/issueLink/%s", linkID)
	_, err := client.Delete(context.Background(), path)
	return err
}
