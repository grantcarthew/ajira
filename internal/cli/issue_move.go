package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/converter"
	"github.com/spf13/cobra"
)

// transitionsResponse matches the Jira transitions API response.
type transitionsResponse struct {
	Transitions []transition `json:"transitions"`
}

type transition struct {
	ID   string           `json:"id"`
	Name string           `json:"name"`
	To   transitionStatus `json:"to"`
}

type transitionStatus struct {
	Name string `json:"name"`
}

// transitionRequest represents the request body for transitioning an issue.
type transitionRequest struct {
	Transition transitionRef  `json:"transition"`
	Fields     map[string]any `json:"fields,omitempty"`
	Update     map[string]any `json:"update,omitempty"`
}

type transitionRef struct {
	ID string `json:"id"`
}

var (
	moveListTransitions bool
	moveComment         string
	moveResolution      string
	moveAssignee        string
)

var issueMoveCmd = &cobra.Command{
	Use:   "move <issue-key> [status]",
	Short: "Transition an issue to a new status",
	Long:  "Move a Jira issue to a new status via workflow transition.",
	Example: `  ajira issue move PROJ-123                              # List available transitions
  ajira issue move PROJ-123 "In Progress"                # Move to In Progress
  ajira issue move PROJ-123 Done                         # Move to Done
  ajira issue move PROJ-123 Done -m "Completed work"     # Move with comment
  ajira issue move PROJ-123 Done -R Done                 # Move with resolution
  ajira issue move PROJ-123 "In Progress" -a me          # Move and assign`,
	Args:         cobra.RangeArgs(1, 2),
	SilenceUsage: true,
	RunE:         runIssueMove,
}

func init() {
	issueMoveCmd.Flags().BoolVar(&moveListTransitions, "list", false, "List available transitions")
	issueMoveCmd.Flags().StringVarP(&moveComment, "comment", "m", "", "Add comment during transition")
	issueMoveCmd.Flags().StringVarP(&moveResolution, "resolution", "R", "", "Set resolution (e.g., Done, Won't Do)")
	issueMoveCmd.Flags().StringVarP(&moveAssignee, "assignee", "a", "", "Set assignee (email, accountId, me)")

	issueCmd.AddCommand(issueMoveCmd)
}

func runIssueMove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Get available transitions
	transitions, err := getTransitions(ctx, client, issueKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("Failed to get transitions: %v", err)
	}

	// List mode: show available transitions
	if moveListTransitions || len(args) == 1 {
		if JSONOutput() {
			output, _ := json.MarshalIndent(transitions, "", "  ")
			fmt.Println(string(output))
		} else {
			if len(transitions) == 0 {
				fmt.Println("No transitions available.")
				return nil
			}
			fmt.Printf("Available transitions for %s:\n", issueKey)
			for _, t := range transitions {
				fmt.Printf("  %s -> %s\n", t.Name, t.To.Name)
			}
		}
		return nil
	}

	// Transition mode: apply transition
	targetStatus := args[1]

	// Find matching transition
	var matchedTransition *transition
	for _, t := range transitions {
		if strings.EqualFold(t.Name, targetStatus) || strings.EqualFold(t.To.Name, targetStatus) {
			matchedTransition = &t
			break
		}
	}

	if matchedTransition == nil {
		var available []string
		for _, t := range transitions {
			available = append(available, t.Name)
		}
		return fmt.Errorf("Transition not available: %s (available: %s)", targetStatus, strings.Join(available, ", "))
	}

	// Build fields and update maps for transition
	var fields map[string]any
	var update map[string]any

	if moveResolution != "" || moveAssignee != "" {
		fields = make(map[string]any)
	}

	if moveResolution != "" {
		fields["resolution"] = map[string]string{"name": moveResolution}
	}

	if moveAssignee != "" {
		var accountID string
		if strings.EqualFold(moveAssignee, "me") {
			accountID, err = resolveUser(ctx, client, cfg.Email)
			if err != nil {
				return fmt.Errorf("Failed to resolve current user: %v", err)
			}
		} else {
			accountID, err = resolveUser(ctx, client, moveAssignee)
			if err != nil {
				return fmt.Errorf("Failed to resolve user: %v", err)
			}
			if accountID == "" {
				return fmt.Errorf("User not found: %s", moveAssignee)
			}
		}
		fields["assignee"] = map[string]string{"accountId": accountID}
	}

	if moveComment != "" {
		update = make(map[string]any)
		adf, err := converter.MarkdownToADF(moveComment)
		if err != nil {
			return fmt.Errorf("Failed to convert comment: %v", err)
		}
		update["comment"] = []map[string]any{
			{"add": map[string]any{"body": adf}},
		}
	}

	err = doTransition(ctx, client, issueKey, matchedTransition.ID, fields, update)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %v", apiErr)
		}
		return fmt.Errorf("Failed to transition issue: %v", err)
	}

	if JSONOutput() {
		result := map[string]string{
			"key":    issueKey,
			"status": matchedTransition.To.Name,
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

func getTransitions(ctx context.Context, client *api.Client, key string) ([]transition, error) {
	path := fmt.Sprintf("/issue/%s/transitions", key)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp transitionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Failed to parse response: %w", err)
	}

	return resp.Transitions, nil
}

func doTransition(ctx context.Context, client *api.Client, key, transitionID string, fields, update map[string]any) error {
	req := transitionRequest{
		Transition: transitionRef{ID: transitionID},
		Fields:     fields,
		Update:     update,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("Failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/transitions", key)
	_, err = client.Post(ctx, path, body)
	return err
}
