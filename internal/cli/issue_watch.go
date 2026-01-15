package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

var watchStdin bool

var issueWatchCmd = &cobra.Command{
	Use:   "watch <issue-key>",
	Short: "Start watching an issue",
	Long:  "Add yourself as a watcher to a Jira issue to receive notifications about changes.",
	Example: `  ajira issue watch PROJ-123                      # Watch an issue
  echo -e "PROJ-1\nPROJ-2" | ajira issue watch --stdin  # Batch watch`,
	Args: func(cmd *cobra.Command, args []string) error {
		if watchStdin {
			if len(args) != 0 {
				return fmt.Errorf("with --stdin, no arguments are expected")
			}
		} else {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 argument: <issue-key>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueWatch,
}

var issueUnwatchCmd = &cobra.Command{
	Use:   "unwatch <issue-key>",
	Short: "Stop watching an issue",
	Long:  "Remove yourself as a watcher from a Jira issue to stop receiving notifications.",
	Example: `  ajira issue unwatch PROJ-123                      # Unwatch an issue
  echo -e "PROJ-1\nPROJ-2" | ajira issue unwatch --stdin  # Batch unwatch`,
	Args: func(cmd *cobra.Command, args []string) error {
		if unwatchStdin {
			if len(args) != 0 {
				return fmt.Errorf("with --stdin, no arguments are expected")
			}
		} else {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 argument: <issue-key>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueUnwatch,
}

var unwatchStdin bool

func init() {
	issueWatchCmd.Flags().BoolVar(&watchStdin, "stdin", false, "Read issue keys from stdin (one per line)")
	issueUnwatchCmd.Flags().BoolVar(&unwatchStdin, "stdin", false, "Read issue keys from stdin (one per line)")
	issueCmd.AddCommand(issueWatchCmd)
	issueCmd.AddCommand(issueUnwatchCmd)
}

func runIssueWatch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Get current user's account ID
	accountID, err := getCurrentUserAccountID(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	var issueKeys []string
	if watchStdin {
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
		if len(issueKeys) == 1 {
			PrintDryRun(fmt.Sprintf("watch %s", issueKeys[0]))
		} else {
			PrintDryRunBatch(issueKeys, "watch")
		}
		return nil
	}

	// Single issue watch
	if len(issueKeys) == 1 {
		err = addWatcher(ctx, client, issueKeys[0], accountID)
		if err != nil {
			return err
		}

		if JSONOutput() {
			PrintSuccessJSON(map[string]string{"key": issueKeys[0], "action": "watch"})
		} else {
			PrintSuccess(IssueURL(cfg.BaseURL, issueKeys[0]))
		}
		return nil
	}

	// Batch watch
	var results []BatchResult
	for _, key := range issueKeys {
		err := addWatcher(ctx, client, key, accountID)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
		} else {
			results = append(results, BatchResult{Key: key, Success: true})
		}
	}

	return PrintBatchResults(results)
}

func runIssueUnwatch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Get current user's account ID
	accountID, err := getCurrentUserAccountID(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	var issueKeys []string
	if unwatchStdin {
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
		if len(issueKeys) == 1 {
			PrintDryRun(fmt.Sprintf("unwatch %s", issueKeys[0]))
		} else {
			PrintDryRunBatch(issueKeys, "unwatch")
		}
		return nil
	}

	// Single issue unwatch
	if len(issueKeys) == 1 {
		err = removeWatcher(ctx, client, issueKeys[0], accountID)
		if err != nil {
			return err
		}

		if JSONOutput() {
			PrintSuccessJSON(map[string]string{"key": issueKeys[0], "action": "unwatch"})
		} else {
			PrintSuccess(IssueURL(cfg.BaseURL, issueKeys[0]))
		}
		return nil
	}

	// Batch unwatch
	var results []BatchResult
	for _, key := range issueKeys {
		err := removeWatcher(ctx, client, key, accountID)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
		} else {
			results = append(results, BatchResult{Key: key, Success: true})
		}
	}

	return PrintBatchResults(results)
}

// getCurrentUserAccountID fetches the current user's account ID.
func getCurrentUserAccountID(ctx context.Context, client *api.Client) (string, error) {
	body, err := client.Get(ctx, "/myself")
	if err != nil {
		return "", err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return "", fmt.Errorf("failed to parse user response: %w", err)
	}

	return user.AccountID, nil
}

// addWatcher adds a watcher to an issue.
func addWatcher(ctx context.Context, client *api.Client, issueKey, accountID string) error {
	path := fmt.Sprintf("/issue/%s/watchers", issueKey)

	// Jira API expects just the quoted account ID string as the body
	body, err := json.Marshal(accountID)
	if err != nil {
		return fmt.Errorf("failed to marshal account ID: %w", err)
	}

	_, err = client.Post(ctx, path, body)
	return err
}

// removeWatcher removes a watcher from an issue.
func removeWatcher(ctx context.Context, client *api.Client, issueKey, accountID string) error {
	path := fmt.Sprintf("/issue/%s/watchers?accountId=%s", issueKey, accountID)
	_, err := client.Delete(ctx, path)
	return err
}
