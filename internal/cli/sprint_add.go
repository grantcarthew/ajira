package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// sprintAddRequest matches the Jira Agile API request for moving issues to sprint.
type sprintAddRequest struct {
	Issues []string `json:"issues"`
}

var sprintAddCmd = &cobra.Command{
	Use:   "add <sprint-id> <issue-keys...>",
	Short: "Add issues to a sprint",
	Long:  "Move issues to a sprint. Issues can only be moved to open or active sprints.",
	Example: `  ajira sprint add 42 GCP-123 GCP-124 GCP-125
  ajira sprint add 42 GCP-100`,
	Args:         cobra.MinimumNArgs(2),
	SilenceUsage: true,
	RunE:         runSprintAdd,
}

func init() {
	sprintCmd.AddCommand(sprintAddCmd)
}

func runSprintAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	sprintID := args[0]
	issueKeys := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	err = addIssuesToSprint(ctx, client, sprintID, issueKeys)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("failed to add issues to sprint: %v", err)
	}

	if JSONOutput() {
		result := map[string]any{
			"sprintId": sprintID,
			"issues":   issueKeys,
			"count":    len(issueKeys),
		}
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(issueKeys) == 1 {
			fmt.Printf("Added 1 issue to sprint %s\n", sprintID)
		} else {
			fmt.Printf("Added %d issues to sprint %s\n", len(issueKeys), sprintID)
		}
	}

	return nil
}

func addIssuesToSprint(ctx context.Context, client *api.Client, sprintID string, issueKeys []string) error {
	req := sprintAddRequest{
		Issues: issueKeys,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/sprint/%s/issue", sprintID)
	_, err = client.AgilePost(ctx, path, body)
	return err
}
