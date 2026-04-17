package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// assigneeRequest represents the request body for assigning an issue.
type assigneeRequest struct {
	AccountID *string `json:"accountId"`
}

var assignStdin bool

var issueAssignCmd = &cobra.Command{
	Use:   "assign <issue-key> <user>",
	Short: "Assign issue",
	Long:  "Assign an issue to a user. Use 'me', 'unassigned', or --stdin for batch.",
	Example: `  ajira issue assign PROJ-123 me                   # Assign to yourself
  ajira issue assign PROJ-123 user@example.com     # Assign by email
  ajira issue assign PROJ-123 unassigned           # Remove assignee
  echo -e "PROJ-1\nPROJ-2" | ajira issue assign --stdin me  # Batch assign`,
	Args: func(cmd *cobra.Command, args []string) error {
		if assignStdin {
			if len(args) != 1 {
				return fmt.Errorf("with --stdin, requires exactly 1 argument: <user>")
			}
		} else {
			if len(args) != 2 {
				return fmt.Errorf("requires exactly 2 arguments: <issue-key> <user>")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueAssign,
}

func init() {
	issueAssignCmd.Flags().BoolVar(&assignStdin, "stdin", false, "Read issue keys from stdin (one per line)")
	issueCmd.AddCommand(issueAssignCmd)
}

func runIssueAssign(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Determine user argument position based on --stdin
	var userArg string
	var issueKeys []string

	if assignStdin {
		userArg = args[0]
		issueKeys, err = ReadKeysFromStdin()
		if err != nil {
			return err
		}
		if len(issueKeys) == 0 {
			return fmt.Errorf("no issue keys provided via stdin")
		}
	} else {
		issueKeys = []string{args[0]}
		userArg = args[1]
	}

	// Resolve user to accountId
	accountID, err := resolveAssigneeInput(ctx, client, cfg.Email, userArg)
	if err != nil {
		return fmt.Errorf("failed to resolve assignee: %w", err)
	}

	// Dry-run mode
	if DryRun() {
		assignee := userArg
		if accountID == nil {
			assignee = "unassigned"
		}
		if len(issueKeys) == 1 {
			PrintDryRun(fmt.Sprintf("assign %s to %s", issueKeys[0], assignee))
		} else {
			PrintDryRunBatch(issueKeys, fmt.Sprintf("assign to %s", assignee))
		}
		return nil
	}

	// Single issue assignment
	if len(issueKeys) == 1 {
		err = assignIssue(ctx, client, issueKeys[0], accountID)
		if err != nil {
			return err
		}

		assignee := userArg
		if accountID == nil {
			assignee = "unassigned"
		}

		if JSONOutput() {
			PrintSuccessJSON(map[string]string{"key": issueKeys[0], "assignee": assignee})
		} else {
			PrintSuccess(IssueURL(cfg.BaseURL, issueKeys[0]))
		}
		return nil
	}

	// Batch assignment
	var results []BatchResult
	for _, key := range issueKeys {
		err := assignIssue(ctx, client, key, accountID)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
		} else {
			results = append(results, BatchResult{Key: key, Success: true})
		}
	}

	return PrintBatchResults(results)
}

func assignIssue(ctx context.Context, client *api.Client, key string, accountID *string) error {
	req := assigneeRequest{AccountID: accountID}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/assignee", key)
	_, err = client.Put(ctx, path, body)
	return err
}
