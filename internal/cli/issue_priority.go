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

// PriorityInfo represents a Jira priority for output.
type PriorityInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type priorityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var issuePriorityCmd = &cobra.Command{
	Use:           "priority",
	Short:         "List available priorities",
	Long:          "List all priorities available in the Jira instance.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runIssuePriority,
}

func init() {
	issueCmd.AddCommand(issuePriorityCmd)
}

func runIssuePriority(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	priorities, err := getPriorities(client)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 {
				return Errorf("authentication failed (401)")
			}
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch priorities: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(priorities, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printPriorities(priorities)
	}

	return nil
}

func getPriorities(client *api.Client) ([]PriorityInfo, error) {
	body, err := client.Get(context.Background(), "/priority")
	if err != nil {
		return nil, err
	}

	var resp []priorityResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var priorities []PriorityInfo
	for _, p := range resp {
		priorities = append(priorities, PriorityInfo{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return priorities, nil
}

func printPriorities(priorities []PriorityInfo) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths
	nameWidth := 4 // "NAME"
	for _, p := range priorities {
		if len(p.Name) > nameWidth {
			nameWidth = len(p.Name)
		}
	}

	// Print header
	fmt.Printf("%s  %s\n",
		header(padRight("NAME", nameWidth)),
		header("DESCRIPTION"))

	// Print rows
	for _, p := range priorities {
		desc := p.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("%s  %s\n", bold(padRight(p.Name, nameWidth)), desc)
	}
}
