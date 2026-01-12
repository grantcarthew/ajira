package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// epicAddRequest matches the Jira Agile API request for moving issues to an epic.
type epicAddRequest struct {
	Issues []string `json:"issues"`
}

var epicAddCmd = &cobra.Command{
	Use:   "add <epic-key> <issue-keys...>",
	Short: "Add issues to an epic",
	Long:  "Move issues to an epic. Issues can only belong to one epic at a time.",
	Example: `  ajira epic add GCP-50 GCP-101 GCP-102 GCP-103
  ajira epic add GCP-50 GCP-100`,
	Args:         cobra.MinimumNArgs(2),
	SilenceUsage: true,
	RunE:         runEpicAdd,
}

func init() {
	epicCmd.AddCommand(epicAddCmd)
}

func runEpicAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	epicKey := args[0]
	issueKeys := args[1:]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	err = addIssuesToEpic(ctx, client, epicKey, issueKeys)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("failed to add issues to epic: %v", err)
	}

	if JSONOutput() {
		result := map[string]any{
			"epicKey": epicKey,
			"issues":  issueKeys,
			"count":   len(issueKeys),
		}
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		if len(issueKeys) == 1 {
			fmt.Printf("Added 1 issue to epic %s\n", epicKey)
		} else {
			fmt.Printf("Added %d issues to epic %s\n", len(issueKeys), epicKey)
		}
	}

	return nil
}

func addIssuesToEpic(ctx context.Context, client *api.Client, epicKey string, issueKeys []string) error {
	req := epicAddRequest{
		Issues: issueKeys,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/epic/%s/issue", epicKey)
	_, err = client.AgilePost(ctx, path, body)
	return err
}
