package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// epicRemoveRequest matches the Jira Agile API request for removing issues from epics.
type epicRemoveRequest struct {
	Issues []string `json:"issues"`
}

var epicRemoveCmd = &cobra.Command{
	Use:   "remove <issue-keys...>",
	Short: "Remove issues from epic",
	Long:  "Remove issues from their current epic. No epic key is needed as this clears the epic link.",
	Example: `  ajira epic remove GCP-101 GCP-102
  ajira epic remove GCP-100`,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE:         runEpicRemove,
}

func init() {
	epicCmd.AddCommand(epicRemoveCmd)
}

func runEpicRemove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	issueKeys := args

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	err = removeIssuesFromEpic(ctx, client, issueKeys)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("failed to remove issues from epic: %v", err)
	}

	if JSONOutput() {
		result := map[string]any{
			"issues": issueKeys,
			"count":  len(issueKeys),
		}
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(issueKeys) == 1 {
			fmt.Println("Removed 1 issue from epic")
		} else {
			fmt.Printf("Removed %d issues from epic\n", len(issueKeys))
		}
	}

	return nil
}

func removeIssuesFromEpic(ctx context.Context, client *api.Client, issueKeys []string) error {
	req := epicRemoveRequest{
		Issues: issueKeys,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// POST to /epic/none/issue removes issues from their current epic
	_, err = client.AgilePost(ctx, "/epic/none/issue", body)
	return err
}
