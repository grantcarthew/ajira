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

var epicRemoveStdin bool

var epicRemoveCmd = &cobra.Command{
	Use:   "remove <issue-keys...>",
	Short: "Remove from epic",
	Long:  "Remove issues from their current epic. Use --stdin for batch.",
	Example: `  ajira epic remove GCP-101 GCP-102
  ajira epic remove GCP-100
  echo -e "GCP-1\nGCP-2" | ajira epic remove --stdin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if epicRemoveStdin {
			if len(args) != 0 {
				return fmt.Errorf("with --stdin, no arguments should be provided")
			}
		} else {
			if len(args) < 1 {
				return fmt.Errorf("requires at least 1 argument: <issue-keys...>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runEpicRemove,
}

func init() {
	epicRemoveCmd.Flags().BoolVar(&epicRemoveStdin, "stdin", false, "Read issue keys from stdin (one per line)")
	epicCmd.AddCommand(epicRemoveCmd)
}

func runEpicRemove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	var issueKeys []string
	var err error

	if epicRemoveStdin {
		issueKeys, err = ReadKeysFromStdin()
		if err != nil {
			return err
		}
		if len(issueKeys) == 0 {
			return fmt.Errorf("no issue keys provided via stdin")
		}
	} else {
		issueKeys = args
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Dry-run mode
	if DryRun() {
		PrintDryRunBatch(issueKeys, "remove from epic")
		return nil
	}

	err = removeIssuesFromEpic(ctx, client, issueKeys)
	if err != nil {
		return err
	}

	if JSONOutput() {
		PrintSuccessJSON(map[string]any{
			"issues": issueKeys,
			"count":  len(issueKeys),
		})
	} else {
		if len(issueKeys) == 1 {
			PrintSuccess("Removed 1 issue from epic")
		} else {
			PrintSuccess(fmt.Sprintf("Removed %d issues from epic", len(issueKeys)))
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
