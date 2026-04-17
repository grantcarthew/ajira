package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// userSearchResponse matches the Jira user search API response.
type userSearchResponse []userSearchResult

type userSearchResult struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
}

// minAccountIDLength is the minimum length to distinguish a Jira account ID
// from other user identifiers. Jira Cloud account IDs are typically 24-28
// character alphanumeric strings (e.g., "5b10ac8d82e05b22cc7d4ef5").
const minAccountIDLength = 20

// resolveAssigneeInput resolves a user input string to a Jira accountId pointer.
// Accepts "me" (resolved via email), an email address, or a raw accountId.
// Returns nil for "unassigned" or empty input.
// Returns an error if the user cannot be found.
func resolveAssigneeInput(ctx context.Context, client *api.Client, email, input string) (*string, error) {
	if input == "" || strings.EqualFold(input, "unassigned") {
		return nil, nil
	}
	userArg := input
	if strings.EqualFold(input, "me") {
		if email == "" {
			return nil, fmt.Errorf("JIRA_EMAIL is required to resolve 'me'")
		}
		userArg = email
	}
	accountID, err := resolveUser(ctx, client, userArg)
	if err != nil {
		return nil, err
	}
	if accountID == "" {
		if userArg != input {
			return nil, fmt.Errorf("user not found: %s (resolved to %s)", input, userArg)
		}
		return nil, fmt.Errorf("user not found: %s", input)
	}
	return &accountID, nil
}

// resolveUser resolves a user identifier to an accountId.
// Accepts email address or accountId directly.
func resolveUser(ctx context.Context, client *api.Client, user string) (string, error) {
	// If it looks like an accountId (no @ and long enough), use it directly.
	// Jira Cloud account IDs are alphanumeric strings like "5b10ac8d82e05b22cc7d4ef5".
	if !strings.Contains(user, "@") && len(user) > minAccountIDLength {
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

// UserInfo represents a Jira user.
type UserInfo struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Active       bool   `json:"active"`
}

var userSearchLimit int

var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"users"},
	Short:   "Manage users",
	Long:    "Commands for searching and viewing Jira users.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var userSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search users",
	Long:  "Search Jira users by name or email. Returns account IDs for use with assign.",
	Example: `  ajira user search john              # Search by name
  ajira user search john@example.com  # Search by email
  ajira user search john -l 20        # Limit results`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runUserSearch,
}

func init() {
	userSearchCmd.Flags().IntVarP(&userSearchLimit, "limit", "l", 10, "Maximum users to return")

	userCmd.AddCommand(userSearchCmd)
	rootCmd.AddCommand(userCmd)
}

func runUserSearch(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	query := args[0]
	users, err := searchUsers(ctx, client, query, userSearchLimit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to search users: %w", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		if len(users) == 0 {
			fmt.Println("No users found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DISPLAY NAME\tEMAIL\tACCOUNT ID\tACTIVE")
		for _, u := range users {
			active := "Yes"
			if !u.Active {
				active = "No"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", u.DisplayName, u.EmailAddress, u.AccountID, active)
		}
		w.Flush()
	}

	return nil
}

// searchUsers searches for users matching the query.
func searchUsers(ctx context.Context, client *api.Client, query string, limit int) ([]UserInfo, error) {
	path := fmt.Sprintf("/user/search?query=%s&maxResults=%d", url.QueryEscape(query), limit)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var users []userSearchResult
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := make([]UserInfo, len(users))
	for i, u := range users {
		result[i] = UserInfo{
			AccountID:    u.AccountID,
			DisplayName:  u.DisplayName,
			EmailAddress: u.EmailAddress,
			Active:       u.Active,
		}
	}

	return result, nil
}
