package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/width"
	"github.com/spf13/cobra"
)

// SprintInfo represents a Jira sprint for output.
type SprintInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Goal      string `json:"goal,omitempty"`
}

// sprintListResponse matches the Jira Agile sprint list API response.
type sprintListResponse struct {
	MaxResults int           `json:"maxResults"`
	StartAt    int           `json:"startAt"`
	IsLast     bool          `json:"isLast"`
	Values     []sprintValue `json:"values"`
}

type sprintValue struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
	Goal      string `json:"goal,omitempty"`
}

var (
	sprintListState   string
	sprintListCurrent bool
	sprintListLimit   int
)

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints",
	Long:  "List Jira sprints for a board.",
	Example: `  ajira sprint list --board 1342          # List sprints for board
  ajira sprint list --state active        # List active sprints only
  ajira sprint list --current             # Shorthand for --state active
  ajira sprint list --state closed -l 5   # Last 5 closed sprints`,
	SilenceUsage: true,
	RunE:         runSprintList,
}

func init() {
	sprintListCmd.Flags().StringVar(&sprintListState, "state", "", "Filter by state (active, future, closed)")
	sprintListCmd.Flags().BoolVar(&sprintListCurrent, "current", false, "Show current active sprints (shorthand for --state active)")
	sprintListCmd.Flags().IntVarP(&sprintListLimit, "limit", "l", 0, "Maximum sprints to return (0 = all)")

	sprintCmd.AddCommand(sprintListCmd)
}

func runSprintList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	boardID := Board()
	if boardID == "" {
		return fmt.Errorf("board ID required; use --board flag or set JIRA_BOARD")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Handle --current as shorthand for --state active
	state := sprintListState
	if sprintListCurrent && state == "" {
		state = "active"
	}

	sprints, err := listSprints(ctx, client, boardID, state, sprintListLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("failed to list sprints: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(sprints, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(sprints) == 0 {
			fmt.Println("No sprints found.")
			return nil
		}

		bold := color.New(color.Bold).SprintFunc()
		header := color.New(color.FgCyan, color.Bold).SprintFunc()

		// Calculate column widths
		idWidth, nameWidth, stateWidth, startWidth, endWidth := 4, 4, 6, 10, 10
		for _, s := range sprints {
			idStr := fmt.Sprintf("%d", s.ID)
			if w := width.StringWidth(idStr); w > idWidth {
				idWidth = w
			}
			if w := width.StringWidth(s.Name); w > nameWidth {
				nameWidth = w
			}
			if w := width.StringWidth(s.State); w > stateWidth {
				stateWidth = w
			}
		}

		// Cap name width for display
		if nameWidth > 30 {
			nameWidth = 30
		}

		// Print header
		fmt.Printf("%s  %s  %s  %s  %s  %s\n",
			header(padRight("ID", idWidth)),
			header(padRight("NAME", nameWidth)),
			header(padRight("STATE", stateWidth)),
			header(padRight("START", startWidth)),
			header(padRight("END", endWidth)),
			header("GOAL"))

		// Print rows
		for _, s := range sprints {
			idStr := fmt.Sprintf("%d", s.ID)
			name := width.Truncate(s.Name, nameWidth, "...")
			start := formatSprintDate(s.StartDate)
			end := formatSprintDate(s.EndDate)
			goal := width.Truncate(s.Goal, 40, "...")

			stateColored := colorSprintState(padRight(s.State, stateWidth), s.State)

			fmt.Printf("%s  %s  %s  %s  %s  %s\n",
				bold(padRight(idStr, idWidth)),
				padRight(name, nameWidth),
				stateColored,
				padRight(start, startWidth),
				padRight(end, endWidth),
				goal)
		}
	}

	return nil
}

func listSprints(ctx context.Context, client *api.Client, boardID, state string, limit int) ([]SprintInfo, error) {
	var allSprints []SprintInfo
	maxResults := 50
	if limit > 0 && limit < maxResults {
		maxResults = limit
	}

	startAt := 0
	const maxPages = 100

	for range maxPages {
		path := fmt.Sprintf("/board/%s/sprint?maxResults=%d&startAt=%d", url.PathEscape(boardID), maxResults, startAt)
		if state != "" {
			path += "&state=" + url.QueryEscape(state)
		}

		body, err := client.AgileGet(ctx, path)
		if err != nil {
			return nil, err
		}

		var resp sprintListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, s := range resp.Values {
			info := SprintInfo(s)

			allSprints = append(allSprints, info)

			if limit > 0 && len(allSprints) >= limit {
				return allSprints[:limit], nil
			}
		}

		if resp.IsLast || len(resp.Values) == 0 {
			break
		}

		startAt += len(resp.Values)
	}

	return allSprints, nil
}

// formatSprintDate formats a sprint date string to just the date portion.
func formatSprintDate(dateStr string) string {
	if dateStr == "" {
		return "-"
	}
	// Sprint dates are in ISO format like "2026-01-06T00:00:00.000Z"
	// Return just the date portion
	if len(dateStr) >= 10 {
		return dateStr[:10]
	}
	return dateStr
}

// colorSprintState returns a colored sprint state string.
func colorSprintState(state, rawState string) string {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()

	switch strings.ToLower(rawState) {
	case "active":
		return green(state)
	case "future":
		return blue(state)
	case "closed":
		return faint(state)
	default:
		return state
	}
}
