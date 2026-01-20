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
	moveStdin           bool
)

var issueMoveCmd = &cobra.Command{
	Use:   "move <issue-key> [status]",
	Short: "Move issue",
	Long:  "Transition an issue to a new status. Supports -m comment, -R resolution, -a assignee, --stdin for batch.",
	Example: `  ajira issue move PROJ-123                              # List available transitions
  ajira issue move PROJ-123 "In Progress"                # Move to In Progress
  ajira issue move PROJ-123 Done                         # Move to Done
  ajira issue move PROJ-123 Done -m "Completed work"     # Move with comment
  ajira issue move PROJ-123 Done -R Done                 # Move with resolution
  ajira issue move PROJ-123 "In Progress" -a me          # Move and assign
  echo -e "PROJ-1\nPROJ-2" | ajira issue move --stdin Done  # Batch move`,
	Args: func(cmd *cobra.Command, args []string) error {
		if moveStdin {
			if len(args) != 1 {
				return fmt.Errorf("with --stdin, requires exactly 1 argument: <status>")
			}
		} else {
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("requires 1 or 2 arguments: <issue-key> [status]")
			}
		}
		return nil
	},
	SilenceUsage: true,
	RunE:         runIssueMove,
}

func init() {
	issueMoveCmd.Flags().BoolVar(&moveListTransitions, "list", false, "List available transitions")
	issueMoveCmd.Flags().StringVarP(&moveComment, "comment", "m", "", "Add comment during transition")
	issueMoveCmd.Flags().StringVarP(&moveResolution, "resolution", "R", "", "Set resolution (e.g., Done, Won't Do)")
	issueMoveCmd.Flags().StringVarP(&moveAssignee, "assignee", "a", "", "Set assignee (email, accountId, me)")
	issueMoveCmd.Flags().BoolVar(&moveStdin, "stdin", false, "Read issue keys from stdin (one per line)")

	issueCmd.AddCommand(issueMoveCmd)
}

func runIssueMove(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Handle --stdin mode
	if moveStdin {
		return runIssueMoveStdin(ctx, client, cfg, args[0])
	}

	issueKey := args[0]

	// Get available transitions
	transitions, err := getTransitions(ctx, client, issueKey)
	if err != nil {
		return err
	}

	// List mode: show available transitions
	if moveListTransitions || len(args) == 1 {
		if JSONOutput() {
			output, err := json.MarshalIndent(transitions, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
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
	matchedTransition := findTransition(transitions, targetStatus)
	if matchedTransition == nil {
		var available []string
		for _, t := range transitions {
			available = append(available, t.Name)
		}
		return fmt.Errorf("transition not available: %s (available: %s)", targetStatus, strings.Join(available, ", "))
	}

	// Build fields for transition
	fields, update, err := buildTransitionOptions(ctx, client, cfg)
	if err != nil {
		return err
	}

	// Dry-run mode
	if DryRun() {
		action := fmt.Sprintf("transition %s to %s", issueKey, matchedTransition.To.Name)
		if moveAssignee != "" {
			action += fmt.Sprintf(" and assign to %s", moveAssignee)
		}
		if moveResolution != "" {
			action += fmt.Sprintf(" with resolution %s", moveResolution)
		}
		PrintDryRun(action)
		return nil
	}

	err = doTransition(ctx, client, issueKey, matchedTransition.ID, fields, update)
	if err != nil {
		return err
	}

	if JSONOutput() {
		PrintSuccessJSON(map[string]string{
			"key":    issueKey,
			"status": matchedTransition.To.Name,
		})
	} else {
		PrintSuccess(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

func runIssueMoveStdin(ctx context.Context, client *api.Client, cfg *config.Config, targetStatus string) error {
	issueKeys, err := ReadKeysFromStdin()
	if err != nil {
		return err
	}
	if len(issueKeys) == 0 {
		return fmt.Errorf("no issue keys provided via stdin")
	}

	// Build fields once (shared across all issues)
	fields, update, err := buildTransitionOptions(ctx, client, cfg)
	if err != nil {
		return err
	}

	// Dry-run mode
	if DryRun() {
		action := fmt.Sprintf("transition to %s", targetStatus)
		if moveAssignee != "" {
			action += fmt.Sprintf(" and assign to %s", moveAssignee)
		}
		PrintDryRunBatch(issueKeys, action)
		return nil
	}

	var results []BatchResult
	for _, key := range issueKeys {
		// Get transitions for this specific issue
		transitions, err := getTransitions(ctx, client, key)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
			continue
		}

		// Find matching transition
		matchedTransition := findTransition(transitions, targetStatus)
		if matchedTransition == nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: fmt.Sprintf("transition not available: %s", targetStatus)})
			continue
		}

		err = doTransition(ctx, client, key, matchedTransition.ID, fields, update)
		if err != nil {
			results = append(results, BatchResult{Key: key, Success: false, Error: err.Error()})
		} else {
			results = append(results, BatchResult{Key: key, Success: true})
		}
	}

	return PrintBatchResults(results)
}

func buildTransitionOptions(ctx context.Context, client *api.Client, cfg *config.Config) (map[string]any, map[string]any, error) {
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
		var err error
		if strings.EqualFold(moveAssignee, "me") {
			accountID, err = resolveUser(ctx, client, cfg.Email)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to resolve current user: %w", err)
			}
		} else {
			accountID, err = resolveUser(ctx, client, moveAssignee)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to resolve user: %w", err)
			}
			if accountID == "" {
				return nil, nil, fmt.Errorf("user not found: %s", moveAssignee)
			}
		}
		fields["assignee"] = map[string]string{"accountId": accountID}
	}

	if moveComment != "" {
		update = make(map[string]any)
		adf, err := converter.MarkdownToADF(moveComment)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert comment: %w", err)
		}
		update["comment"] = []map[string]any{
			{"add": map[string]any{"body": adf}},
		}
	}

	return fields, update, nil
}

func findTransition(transitions []transition, targetStatus string) *transition {
	for _, t := range transitions {
		if strings.EqualFold(t.Name, targetStatus) || strings.EqualFold(t.To.Name, targetStatus) {
			return &t
		}
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
		return nil, fmt.Errorf("failed to parse response: %w", err)
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
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/transitions", key)
	_, err = client.Post(ctx, path, body)
	return err
}
