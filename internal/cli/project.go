package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// ProjectInfo represents a Jira project.
type ProjectInfo struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Lead string `json:"lead"`
	Style string `json:"style"`
}

// projectSearchResponse matches the Jira project search API response.
type projectSearchResponse struct {
	Values     []projectValue `json:"values"`
	StartAt    int            `json:"startAt"`
	MaxResults int            `json:"maxResults"`
	Total      int            `json:"total"`
	IsLast     bool           `json:"isLast"`
}

type projectValue struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
	Lead struct {
		DisplayName string `json:"displayName"`
	} `json:"lead"`
	Style string `json:"style"`
}

var (
	projectQuery string
	projectLimit int
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage Jira projects",
	Long:  "Commands for managing and listing Jira projects.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var projectListCmd = &cobra.Command{
	Use:           "list",
	Short:         "List accessible projects",
	Long:          "List all Jira projects accessible to the current user.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runProjectList,
}

func init() {
	projectListCmd.Flags().StringVarP(&projectQuery, "query", "q", "", "Filter by project name/key")
	projectListCmd.Flags().IntVarP(&projectLimit, "limit", "l", 0, "Maximum projects to return (0 = all)")

	projectCmd.AddCommand(projectListCmd)
	rootCmd.AddCommand(projectCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	projects, err := fetchAllProjects(client, projectQuery, projectLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 {
				return Errorf("authentication failed (401)")
			}
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to fetch projects: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(projects, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tNAME\tLEAD\tSTYLE")
		for _, p := range projects {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Key, p.Name, p.Lead, p.Style)
		}
		w.Flush()
	}

	return nil
}

func fetchAllProjects(client *api.Client, query string, limit int) ([]ProjectInfo, error) {
	var allProjects []ProjectInfo
	startAt := 0
	maxResults := 50

	for {
		path := fmt.Sprintf("/project/search?startAt=%d&maxResults=%d&expand=lead", startAt, maxResults)
		if query != "" {
			path += "&query=" + query
		}

		body, err := client.Get(context.Background(), path)
		if err != nil {
			return nil, err
		}

		var resp projectSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, v := range resp.Values {
			allProjects = append(allProjects, ProjectInfo{
				ID:    v.ID,
				Key:   v.Key,
				Name:  v.Name,
				Lead:  v.Lead.DisplayName,
				Style: v.Style,
			})

			if limit > 0 && len(allProjects) >= limit {
				return allProjects[:limit], nil
			}
		}

		if resp.IsLast {
			break
		}

		startAt += maxResults
	}

	return allProjects, nil
}
