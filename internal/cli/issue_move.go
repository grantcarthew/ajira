package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
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
	Transition transitionRef `json:"transition"`
}

type transitionRef struct {
	ID string `json:"id"`
}

var (
	moveListTransitions bool
)

var issueMoveCmd = &cobra.Command{
	Use:           "move <issue-key> [status]",
	Short:         "Transition an issue to a new status",
	Long:          "Move a Jira issue to a new status via workflow transition.",
	Args:          cobra.RangeArgs(1, 2),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runIssueMove,
}

func init() {
	issueMoveCmd.Flags().BoolVarP(&moveListTransitions, "list", "l", false, "List available transitions")

	issueCmd.AddCommand(issueMoveCmd)
}

func runIssueMove(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Get available transitions
	transitions, err := getTransitions(client, issueKey)
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
		return Errorf("failed to get transitions: %v", err)
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
		return Errorf("transition not available: %s (available: %s)", targetStatus, strings.Join(available, ", "))
	}

	err = doTransition(client, issueKey, matchedTransition.ID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to transition issue: %v", err)
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

func getTransitions(client *api.Client, key string) ([]transition, error) {
	path := fmt.Sprintf("/issue/%s/transitions", key)

	body, err := client.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var resp transitionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.Transitions, nil
}

func doTransition(client *api.Client, key, transitionID string) error {
	req := transitionRequest{
		Transition: transitionRef{ID: transitionID},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s/transitions", key)
	_, err = client.Post(context.Background(), path, body)
	return err
}
