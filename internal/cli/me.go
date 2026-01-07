package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/spf13/cobra"
)

// User represents the current user from /myself endpoint.
type User struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	TimeZone     string `json:"timeZone"`
	Active       bool   `json:"active"`
}

var meCmd = &cobra.Command{
	Use:           "me",
	Short:         "Display current user information",
	Long:          "Display information about the currently authenticated Jira user. Useful for verifying credentials and connection.",
	Example: `  ajira me                # Verify authentication
  ajira me --json         # Get account ID for automation`,
	SilenceUsage:  true,
	RunE:          runMe,
}

func init() {
	rootCmd.AddCommand(meCmd)
}

func runMe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	body, err := client.Get(context.Background(), "/myself")
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to connect to Jira API: %v", err)
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return Errorf("failed to parse response: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(user, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Printf("Display Name: %s\n", user.DisplayName)
		fmt.Printf("Email: %s\n", user.EmailAddress)
		fmt.Printf("Account ID: %s\n", user.AccountID)
		fmt.Printf("Timezone: %s\n", user.TimeZone)
		fmt.Printf("Active: %t\n", user.Active)
	}

	return nil
}
