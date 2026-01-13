package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

var deleteCascade bool

var issueDeleteCmd = &cobra.Command{
	Use:          "delete <issue-key>",
	Short:        "Delete an issue",
	Long:         "Permanently delete a Jira issue. This action cannot be undone.",
	Example: `  ajira issue delete PROJ-123             # Delete issue permanently
  ajira issue delete PROJ-123 --cascade   # Delete issue and all subtasks`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueDelete,
}

func init() {
	issueDeleteCmd.Flags().BoolVar(&deleteCascade, "cascade", false, "Delete issue with all subtasks")

	issueCmd.AddCommand(issueDeleteCmd)
}

func runIssueDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	err = deleteIssue(ctx, client, issueKey, deleteCascade)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("Failed to delete issue: %v", err)
	}

	if JSONOutput() {
		result := map[string]string{"key": issueKey, "status": "deleted"}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("%s deleted\n", issueKey)
	}

	return nil
}

func deleteIssue(ctx context.Context, client *api.Client, key string, cascade bool) error {
	path := fmt.Sprintf("/issue/%s", key)
	if cascade {
		path += "?deleteSubtasks=true"
	}
	_, err := client.Delete(ctx, path)
	return err
}
