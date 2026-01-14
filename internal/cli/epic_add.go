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

var epicAddStdin bool

var epicAddCmd = &cobra.Command{
	Use:   "add <epic-key> <issue-keys...>",
	Short: "Add issues to an epic",
	Long: `Move issues to an epic. Issues can only belong to one epic at a time.

With --stdin, reads issue keys from stdin (one per line).`,
	Example: `  ajira epic add GCP-50 GCP-101 GCP-102 GCP-103
  ajira epic add GCP-50 GCP-100
  echo -e "GCP-1\nGCP-2" | ajira epic add GCP-50 --stdin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if epicAddStdin {
			if len(args) != 1 {
				return fmt.Errorf("with --stdin, requires exactly 1 argument: <epic-key>")
			}
		} else {
			if len(args) < 2 {
				return fmt.Errorf("requires at least 2 arguments: <epic-key> <issue-keys...>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runEpicAdd,
}

func init() {
	epicAddCmd.Flags().BoolVar(&epicAddStdin, "stdin", false, "Read issue keys from stdin (one per line)")
	epicCmd.AddCommand(epicAddCmd)
}

func runEpicAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	epicKey := args[0]

	var issueKeys []string
	var err error

	if epicAddStdin {
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
		PrintDryRunBatch(issueKeys, fmt.Sprintf("add to epic %s", epicKey))
		return nil
	}

	err = addIssuesToEpic(ctx, client, epicKey, issueKeys)
	if err != nil {
		return err
	}

	if JSONOutput() {
		PrintSuccessJSON(map[string]any{
			"epicKey": epicKey,
			"issues":  issueKeys,
			"count":   len(issueKeys),
		})
	} else {
		if len(issueKeys) == 1 {
			PrintSuccess(fmt.Sprintf("Added 1 issue to epic %s", epicKey))
		} else {
			PrintSuccess(fmt.Sprintf("Added %d issues to epic %s", len(issueKeys), epicKey))
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
