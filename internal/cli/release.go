package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// ReleaseInfo represents a project version/release.
type ReleaseInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
}

// releaseListResponse matches the Jira project version API response.
type releaseListResponse struct {
	Values     []releaseValue `json:"values"`
	StartAt    int            `json:"startAt"`
	MaxResults int            `json:"maxResults"`
	Total      int            `json:"total"`
	IsLast     bool           `json:"isLast"`
}

type releaseValue struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	ReleaseDate string `json:"releaseDate"`
	StartDate   string `json:"startDate"`
}

var (
	releaseStatus string
	releaseLimit  int
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Manage project releases/versions",
	Long:  "Commands for listing project releases and versions.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var releaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List project releases/versions",
	Long:  "List all versions/releases for a project. Use -p to specify the project or set JIRA_PROJECT.",
	Example: `  ajira release list                      # List all releases
  ajira release list -p PROJ              # List releases for project
  ajira release list --status released    # Only released versions
  ajira release list --status unreleased  # Only unreleased versions
  ajira release list -l 10                # Limit results`,
	SilenceUsage: true,
	RunE:         runReleaseList,
}

func init() {
	releaseListCmd.Flags().StringVar(&releaseStatus, "status", "", "Filter by status: released, unreleased")
	releaseListCmd.Flags().IntVarP(&releaseLimit, "limit", "l", 0, "Maximum releases to return (0 = all)")

	releaseCmd.AddCommand(releaseListCmd)
	rootCmd.AddCommand(releaseCmd)
}

func runReleaseList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	projectKey := Project()
	if projectKey == "" {
		return fmt.Errorf("project is required: use -p flag or set JIRA_PROJECT environment variable")
	}

	client := api.NewClient(cfg)

	releases, err := fetchAllReleases(ctx, client, projectKey, releaseStatus, releaseLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch releases: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(releases, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(releases) == 0 {
			fmt.Println("No releases found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tRELEASE DATE\tDESCRIPTION")
		for _, r := range releases {
			status := "Unreleased"
			if r.Released {
				status = "Released"
			}
			if r.Archived {
				status = "Archived"
			}
			desc := r.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, status, r.ReleaseDate, desc)
		}
		w.Flush()
	}

	return nil
}

func fetchAllReleases(ctx context.Context, client *api.Client, projectKey, status string, limit int) ([]ReleaseInfo, error) {
	var allReleases []ReleaseInfo
	startAt := 0
	maxResults := 50

	for {
		path := fmt.Sprintf("/project/%s/version?startAt=%d&maxResults=%d&orderBy=name", projectKey, startAt, maxResults)
		if status != "" {
			path += "&status=" + status
		}

		body, err := client.Get(ctx, path)
		if err != nil {
			return nil, err
		}

		var resp releaseListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, v := range resp.Values {
			allReleases = append(allReleases, ReleaseInfo(v))

			if limit > 0 && len(allReleases) >= limit {
				return allReleases[:limit], nil
			}
		}

		if resp.IsLast {
			break
		}

		startAt += maxResults
	}

	return allReleases, nil
}
