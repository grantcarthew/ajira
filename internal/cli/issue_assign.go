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

var issueAssignCmd = &cobra.Command{
	Use:           "assign <issue-key> <user>",
	Short:         "Assign an issue to a user",
	Long:          "Assign a Jira issue to a user. Use 'me' for yourself, or 'unassigned' to remove the assignee.",
	Args:          cobra.ExactArgs(2),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runIssueAssign,
}

func init() {
	issueCmd.AddCommand(issueAssignCmd)
}

func runIssueAssign(cmd *cobra.Command, args []string) error {
	issueKey := args[0]
	userArg := args[1]

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	var accountID *string

	if strings.EqualFold(userArg, "unassigned") {
		// null accountId removes assignee
		accountID = nil
	} else if strings.EqualFold(userArg, "me") {
		// Use current user's email from config
		resolved, err := resolveUser(client, cfg.Email)
		if err != nil {
			return Errorf("failed to resolve current user: %v", err)
		}
		accountID = &resolved
	} else {
		// Resolve user to accountId
		resolved, err := resolveUser(client, userArg)
		if err != nil {
			if apiErr, ok := err.(*api.APIError); ok {
				return Errorf("API error - %v", apiErr)
			}
			return Errorf("failed to resolve user: %v", err)
		}
		if resolved == "" {
			return Errorf("user not found: %s", userArg)
		}
		accountID = &resolved
	}

	err = assignIssue(client, issueKey, accountID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 {
				return Errorf("authentication failed (401)")
			}
			if apiErr.StatusCode == 404 {
				return Errorf("issue not found: %s", issueKey)
			}
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to assign issue: %v", err)
	}

	if JSONOutput() {
		assignee := "unassigned"
		if accountID != nil {
			assignee = userArg
		}
		result := map[string]string{"key": issueKey, "assignee": assignee}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

// resolveUser resolves a user identifier to an accountId.
// Accepts email address or accountId directly.
func resolveUser(client *api.Client, user string) (string, error) {
	// If it looks like an accountId (no @), try using it directly
	// Jira accountIds are typically alphanumeric strings
	if !strings.Contains(user, "@") && len(user) > 20 {
		// Assume it's an accountId
		return user, nil
	}

	// Search by email or display name
	path := fmt.Sprintf("/user/search?query=%s&maxResults=1", url.QueryEscape(user))

	body, err := client.Get(context.Background(), path)
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

func assignIssue(client *api.Client, key string, accountID *string) error {
	req := assigneeRequest{AccountID: accountID}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/assignee", key)
	_, err = client.Put(context.Background(), path, body)
	return err
}
