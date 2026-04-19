package cli

import (
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

var (
	epicCreateSummary  string
	epicCreateBody     string
	epicCreateFile     string
	epicCreatePriority string
	epicCreateLabels   []string
	epicCreateAssignee string
)

var epicCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create epic",
	Long:  "Create an epic. Requires -s for summary.",
	Example: `  ajira epic create -s "Authentication Epic"
  ajira epic create -s "Dashboard" -d "Dashboard features" -P Major
  ajira epic create -s "API" -f description.md
  ajira epic create -s "Auth" -a me                        # Assign to yourself
  ajira epic create -s "Auth" -a user@example.com          # Assign by email
  ajira epic create -s "Auth" -a unassigned                # Explicitly unassigned`,
	SilenceUsage: true,
	RunE:         runEpicCreate,
}

func init() {
	epicCreateCmd.Flags().StringVarP(&epicCreateSummary, "summary", "s", "", "Epic summary (required)")
	epicCreateCmd.Flags().StringVarP(&epicCreateBody, "description", "d", "", "Epic description in Markdown")
	epicCreateCmd.Flags().StringVarP(&epicCreateFile, "file", "f", "", "Read description from file (use - for stdin)")
	epicCreateCmd.Flags().StringVarP(&epicCreatePriority, "priority", "P", "", "Epic priority")
	epicCreateCmd.Flags().StringSliceVar(&epicCreateLabels, "labels", nil, "Epic labels (comma-separated)")
	epicCreateCmd.Flags().StringVarP(&epicCreateAssignee, "assignee", "a", "", "Assignee (me, email, account ID, or unassigned)")

	_ = epicCreateCmd.MarkFlagRequired("summary")

	epicCmd.AddCommand(epicCreateCmd)
}

func runEpicCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if epicCreateSummary == "" {
		return fmt.Errorf("summary is required (use -s or --summary)")
	}

	projectKey := Project()
	if projectKey == "" {
		return fmt.Errorf("project is required (use -p flag or set JIRA_PROJECT)")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg)

	// Validate epic issue type exists and priority
	if err := jira.ValidateIssueType(ctx, client, projectKey, "Epic"); err != nil {
		return err
	}
	if err := jira.ValidatePriority(ctx, client, epicCreatePriority); err != nil {
		return err
	}

	// Get description from body, file, or stdin
	description, err := getEpicDescription()
	if err != nil {
		return fmt.Errorf("failed to read description: %w", err)
	}

	// Resolve assignee to accountId
	assigneeAccountID, err := resolveAssigneeInput(ctx, client, cfg.Email, epicCreateAssignee)
	if err != nil {
		return fmt.Errorf("failed to resolve assignee: %w", err)
	}

	opts := createIssueOptions{
		Project:     projectKey,
		Summary:     epicCreateSummary,
		Description: description,
		IssueType:   "Epic",
		Priority:    epicCreatePriority,
		Labels:      epicCreateLabels,
	}
	if assigneeAccountID != nil {
		opts.Assignee = *assigneeAccountID
	}

	result, err := createIssue(ctx, client, opts)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to create epic: %w", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, result.Key))
	}

	return nil
}

func getEpicDescription() (string, error) {
	return readText(epicCreateFile, epicCreateBody)
}
