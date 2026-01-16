package cli

import (
	"context"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

var (
	deleteCascade bool
	deleteStdin   bool
)

var issueDeleteCmd = &cobra.Command{
	Use:   "delete <issue-key>",
	Short: "Delete issue",
	Long:  "Permanently delete an issue. Use --cascade for subtasks, --stdin for batch.",
	Example: `  ajira issue delete PROJ-123             # Delete issue permanently
  ajira issue delete PROJ-123 --cascade   # Delete issue and all subtasks
  echo -e "PROJ-1\nPROJ-2" | ajira issue delete --stdin  # Batch delete`,
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteStdin {
			if len(args) != 0 {
				return fmt.Errorf("with --stdin, no arguments should be provided")
			}
		} else {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 argument: <issue-key>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueDelete,
}

func init() {
	issueDeleteCmd.Flags().BoolVar(&deleteCascade, "cascade", false, "Delete issue with all subtasks")
	issueDeleteCmd.Flags().BoolVar(&deleteStdin, "stdin", false, "Read issue keys from stdin (one per line)")

	issueCmd.AddCommand(issueDeleteCmd)
}

func runIssueDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	var issueKeys []string
	if deleteStdin {
		issueKeys, err = ReadKeysFromStdin()
		if err != nil {
			return err
		}
		if len(issueKeys) == 0 {
			return fmt.Errorf("no issue keys provided via stdin")
		}
	} else {
		issueKeys = []string{args[0]}
	}

	// Dry-run mode
	if DryRun() {
		action := "delete"
		if deleteCascade {
			action = "delete with subtasks"
		}
		if len(issueKeys) == 1 {
			PrintDryRun(fmt.Sprintf("%s %s", action, issueKeys[0]))
		} else {
			PrintDryRunBatch(issueKeys, action)
		}
		return nil
	}

	// Single delete
	if len(issueKeys) == 1 {
		err = deleteIssue(ctx, client, issueKeys[0], deleteCascade)
		if err != nil {
			return err
		}

		if JSONOutput() {
			PrintSuccessJSON(map[string]string{"key": issueKeys[0], "status": "deleted"})
		} else {
			PrintSuccess(fmt.Sprintf("%s deleted", issueKeys[0]))
		}
		return nil
	}

	// Batch delete
	var results []BatchResult
	for _, key := range issueKeys {
		err := deleteIssue(ctx, client, key, deleteCascade)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
		} else {
			results = append(results, BatchResult{Key: key, Success: true})
		}
	}

	return PrintBatchResults(results)
}

func deleteIssue(ctx context.Context, client *api.Client, key string, cascade bool) error {
	path := fmt.Sprintf("/issue/%s", key)
	if cascade {
		path += "?deleteSubtasks=true"
	}
	_, err := client.Delete(ctx, path)
	return err
}
