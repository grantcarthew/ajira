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

var sprintAddStdin bool

var sprintAddCmd = &cobra.Command{
	Use:   "add <sprint-id> <issue-keys...>",
	Short: "Add to sprint",
	Long:  "Move issues to a sprint. Sprint must be open or active. Use --stdin for batch.",
	Example: `  ajira sprint add 42 GCP-123 GCP-124 GCP-125
  ajira sprint add 42 GCP-100
  echo -e "GCP-1\nGCP-2" | ajira sprint add 42 --stdin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if sprintAddStdin {
			if len(args) != 1 {
				return fmt.Errorf("with --stdin, requires exactly 1 argument: <sprint-id>")
			}
		} else {
			if len(args) < 2 {
				return fmt.Errorf("requires at least 2 arguments: <sprint-id> <issue-keys...>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runSprintAdd,
}

func init() {
	sprintAddCmd.Flags().BoolVar(&sprintAddStdin, "stdin", false, "Read issue keys from stdin (one per line)")
	sprintCmd.AddCommand(sprintAddCmd)
}

func runSprintAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	sprintID := args[0]

	var issueKeys []string
	var err error

	if sprintAddStdin {
		issueKeys, err = ReadKeysFromStdin()
		if err != nil {
			return err
		}
		if len(issueKeys) == 0 {
			return fmt.Errorf("no issue keys provided via stdin")
		}
	} else {
		issueKeys = args[1:]
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Dry-run mode
	if DryRun() {
		PrintDryRunBatch(issueKeys, fmt.Sprintf("add to sprint %s", sprintID))
		return nil
	}

	err = addIssuesToSprint(ctx, client, sprintID, issueKeys)
	if err != nil {
		return err
	}

	if JSONOutput() {
		PrintSuccessJSON(map[string]any{
			"sprintId": sprintID,
			"issues":   issueKeys,
			"count":    len(issueKeys),
		})
	} else {
		if len(issueKeys) == 1 {
			PrintSuccess(fmt.Sprintf("Added 1 issue to sprint %s", sprintID))
		} else {
			PrintSuccess(fmt.Sprintf("Added %d issues to sprint %s", len(issueKeys), sprintID))
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
