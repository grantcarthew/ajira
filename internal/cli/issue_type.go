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

// TypeInfo represents a Jira issue type for output.
type TypeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

type issueTypeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

var issueTypeCmd = &cobra.Command{
	Use:           "type",
	Short:         "List available issue types",
	Long:          "List issue types available for the current project.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runIssueType,
}

func init() {
	issueCmd.AddCommand(issueTypeCmd)
}

func runIssueType(cmd *cobra.Command, args []string) error {
	projectKey := Project()
	if projectKey == "" {
		return Errorf("project is required (use -p flag or set JIRA_PROJECT)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	types, err := getIssueTypes(client, projectKey)
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
		return Errorf("failed to fetch issue types: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(types, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		printIssueTypes(types)
	}

	return nil
}

func getIssueTypes(client *api.Client, projectKey string) ([]TypeInfo, error) {
	// Get issue types for project via createmeta endpoint
	path := fmt.Sprintf("/issue/createmeta/%s/issuetypes", projectKey)

	body, err := client.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var resp struct {
		IssueTypes []issueTypeResponse `json:"issueTypes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		// Try parsing as direct array (different API versions)
		var types []issueTypeResponse
		if err2 := json.Unmarshal(body, &types); err2 != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		resp.IssueTypes = types
	}

	var issueTypes []TypeInfo
	for _, t := range resp.IssueTypes {
		issueTypes = append(issueTypes, TypeInfo{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			Subtask:     t.Subtask,
		})
	}

	return issueTypes, nil
}

func printIssueTypes(types []TypeInfo) {
	bold := color.New(color.Bold).SprintFunc()
	header := color.New(color.FgCyan, color.Bold).SprintFunc()

	// Calculate column widths
	nameWidth := 4 // "NAME"
	for _, t := range types {
		if len(t.Name) > nameWidth {
			nameWidth = len(t.Name)
		}
	}

	// Print header
	fmt.Printf("%s  %s\n",
		header(padRight("NAME", nameWidth)),
		header("DESCRIPTION"))

	// Print rows
	for _, t := range types {
		desc := t.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("%s  %s\n", bold(padRight(t.Name, nameWidth)), desc)
	}
}
