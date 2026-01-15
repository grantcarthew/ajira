package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// UserInfo represents a Jira user.
type UserInfo struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Active       bool   `json:"active"`
}

var userSearchLimit int

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage Jira users",
	Long:  "Commands for searching and viewing Jira users.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var userSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for users",
	Long: `Search for Jira users by name or email address.

The query matches against display name and email. Useful for finding
account IDs for assignment or other operations.`,
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
		return fmt.Errorf("failed to search users: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
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
			Active:       true, // Default to true if not in response
		}
	}

	return result, nil
}
