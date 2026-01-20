package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/width"
	"github.com/spf13/cobra"
)

// BoardInfo represents a Jira board for output.
type BoardInfo struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Project string `json:"project,omitempty"`
}

// boardListResponse matches the Jira Agile board list API response.
type boardListResponse struct {
	MaxResults int          `json:"maxResults"`
	StartAt    int          `json:"startAt"`
	Total      int          `json:"total"`
	IsLast     bool         `json:"isLast"`
	Values     []boardValue `json:"values"`
}

type boardValue struct {
	ID       int            `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Location *boardLocation `json:"location,omitempty"`
}

type boardLocation struct {
	ProjectKey string `json:"projectKey"`
}

var boardListLimit int

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Manage boards",
	Long:  "Commands for managing Jira boards.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List boards",
	Long:  "List Jira boards. Use -p to filter by project.",
	Example: `  ajira board list                    # List boards in default project
  ajira board list -p GCP             # List boards in specific project
  ajira board list -l 10              # Limit results`,
	SilenceUsage: true,
	RunE:         runBoardList,
}

func init() {
	boardListCmd.Flags().IntVarP(&boardListLimit, "limit", "l", 0, "Maximum boards to return (0 = all)")

	boardCmd.AddCommand(boardListCmd)
	rootCmd.AddCommand(boardCmd)
}

func runBoardList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	boards, err := listBoards(ctx, client, Project(), boardListLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to list boards: %w", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(boards, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(boards) == 0 {
			fmt.Println("No boards found.")
			return nil
		}

		bold := color.New(color.Bold).SprintFunc()
		header := color.New(color.FgCyan, color.Bold).SprintFunc()

		// Calculate column widths
		idWidth, nameWidth, typeWidth, projectWidth := 4, 4, 4, 7
		for _, b := range boards {
			idStr := fmt.Sprintf("%d", b.ID)
			if w := width.StringWidth(idStr); w > idWidth {
				idWidth = w
			}
			if w := width.StringWidth(b.Name); w > nameWidth {
				nameWidth = w
			}
			if w := width.StringWidth(b.Type); w > typeWidth {
				typeWidth = w
			}
			if w := width.StringWidth(b.Project); w > projectWidth {
				projectWidth = w
			}
		}

		// Cap name width for display
		if nameWidth > 40 {
			nameWidth = 40
		}

		// Print header
		fmt.Printf("%s  %s  %s  %s\n",
			header(padRight("ID", idWidth)),
			header(padRight("NAME", nameWidth)),
			header(padRight("TYPE", typeWidth)),
			header(padRight("PROJECT", projectWidth)))

		// Print rows
		for _, b := range boards {
			idStr := fmt.Sprintf("%d", b.ID)
			name := width.Truncate(b.Name, nameWidth, "...")

			fmt.Printf("%s  %s  %s  %s\n",
				bold(padRight(idStr, idWidth)),
				padRight(name, nameWidth),
				padRight(b.Type, typeWidth),
				padRight(b.Project, projectWidth))
		}
	}

	return nil
}

func listBoards(ctx context.Context, client *api.Client, projectKey string, limit int) ([]BoardInfo, error) {
	var allBoards []BoardInfo
	maxResults := 50
	if limit > 0 && limit < maxResults {
		maxResults = limit
	}

	startAt := 0
	const maxPages = 100

	for range maxPages {
		path := fmt.Sprintf("/board?maxResults=%d&startAt=%d", maxResults, startAt)
		if projectKey != "" {
			path += "&projectKeyOrId=" + url.QueryEscape(projectKey)
		}

		body, err := client.AgileGet(ctx, path)
		if err != nil {
			return nil, err
		}

		var resp boardListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, b := range resp.Values {
			info := BoardInfo{
				ID:   b.ID,
				Name: b.Name,
				Type: b.Type,
			}
			if b.Location != nil {
				info.Project = b.Location.ProjectKey
			}

			allBoards = append(allBoards, info)

			if limit > 0 && len(allBoards) >= limit {
				return allBoards[:limit], nil
			}
		}

		if resp.IsLast || len(resp.Values) == 0 {
			break
		}

		startAt += len(resp.Values)
	}

	return allBoards, nil
}
