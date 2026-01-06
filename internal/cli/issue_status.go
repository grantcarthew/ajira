package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// StatusInfo represents a Jira status for output.
type StatusInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type statusResponse struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	StatusCategory statusCategoryField `json:"statusCategory"`
}

type statusCategoryField struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type projectStatusesResponse struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Statuses []statusResponse `json:"statuses"`
}

var issueStatusCmd = &cobra.Command{
	Use:           "status",
	Short:         "List available statuses",
	Long:          "List statuses available for the current project.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runIssueStatus,
}

func init() {
	issueCmd.AddCommand(issueStatusCmd)
}

func runIssueStatus(cmd *cobra.Command, args []string) error {
	projectKey := Project()
	if projectKey == "" {
		return Errorf("project is required (use -p flag or set JIRA_PROJECT)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	statuses, err := getStatuses(client, projectKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 {
				return Errorf("authentication failed (401)")
			}
			if apiErr.StatusCode == 404 {
				return Errorf("project not found: %s", projectKey)
			}
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch statuses: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(statuses, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printStatuses(statuses)
	}

	return nil
}

func getStatuses(client *api.Client, projectKey string) ([]StatusInfo, error) {
	path := fmt.Sprintf("/project/%s/statuses", projectKey)

	body, err := client.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var resp []projectStatusesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Deduplicate statuses across issue types
	seen := make(map[string]bool)
	var statuses []StatusInfo
	for _, issueType := range resp {
		for _, s := range issueType.Statuses {
			if seen[s.ID] {
				continue
			}
			seen[s.ID] = true
			statuses = append(statuses, StatusInfo{
				ID:       s.ID,
				Name:     s.Name,
				Category: s.StatusCategory.Name,
			})
		}
	}

	return statuses, nil
}

func printStatuses(statuses []StatusInfo) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths
	nameWidth := 4 // "NAME"
	for _, s := range statuses {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}

	// Print header
	fmt.Printf("%s  %s\n",
		header(padRight("NAME", nameWidth)),
		header("CATEGORY"))

	// Print rows
	for _, s := range statuses {
		fmt.Printf("%s  %s\n", bold(padRight(s.Name, nameWidth)), s.Category)
	}
}
