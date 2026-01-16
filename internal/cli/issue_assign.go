package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// assigneeRequest represents the request body for assigning an issue.
type assigneeRequest struct {
	AccountID *string `json:"accountId"`
}

// userSearchResponse matches the Jira user search API response.
type userSearchResponse []userSearchResult

type userSearchResult struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
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
	var accountID *string
	if strings.EqualFold(userArg, "unassigned") {
		accountID = nil
	} else if strings.EqualFold(userArg, "me") {
		resolved, err := resolveUser(ctx, client, cfg.Email)
		if err != nil {
			return fmt.Errorf("failed to resolve current user: %w", err)
		}
		accountID = &resolved
	} else {
		resolved, err := resolveUser(ctx, client, userArg)
		if err != nil {
			return err
		}
		if resolved == "" {
			return fmt.Errorf("user not found: %s", userArg)
		}
		accountID = &resolved
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

// resolveUser resolves a user identifier to an accountId.
// Accepts email address or accountId directly.
func resolveUser(ctx context.Context, client *api.Client, user string) (string, error) {
	// If it looks like an accountId (no @), try using it directly
	// Jira accountIds are typically alphanumeric strings
	if !strings.Contains(user, "@") && len(user) > 20 {
		// Assume it's an accountId
		return user, nil
	}

	// Search by email or display name
	path := fmt.Sprintf("/user/search?query=%s&maxResults=1", url.QueryEscape(user))

	body, err := client.Get(ctx, path)
	if err != nil {
		return "", err
	}

	var users userSearchResponse
	if err := json.Unmarshal(body, &users); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(users) == 0 {
		return "", nil
	}

	return users[0].AccountID, nil
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
