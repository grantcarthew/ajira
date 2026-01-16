package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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
)

var epicCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create epic",
	Long:  "Create an epic. Requires -s for summary.",
	Example: `  ajira epic create -s "Authentication Epic"
  ajira epic create -s "Dashboard" -d "Dashboard features" -P Major
  ajira epic create -s "API" -f description.md`,
	SilenceUsage: true,
	RunE:         runEpicCreate,
}

func init() {
	epicCreateCmd.Flags().StringVarP(&epicCreateSummary, "summary", "s", "", "Epic summary (required)")
	epicCreateCmd.Flags().StringVarP(&epicCreateBody, "description", "d", "", "Epic description in Markdown")
	epicCreateCmd.Flags().StringVarP(&epicCreateFile, "file", "f", "", "Read description from file (use - for stdin)")
	epicCreateCmd.Flags().StringVarP(&epicCreatePriority, "priority", "P", "", "Epic priority")
	epicCreateCmd.Flags().StringSliceVar(&epicCreateLabels, "labels", nil, "Epic labels (comma-separated)")

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
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Validate epic issue type exists and priority
	if err := jira.ValidateIssueType(ctx, client, projectKey, "Epic"); err != nil {
		return fmt.Errorf("%v", err)
	}
	if err := jira.ValidatePriority(ctx, client, epicCreatePriority); err != nil {
		return fmt.Errorf("%v", err)
	}

	// Get description from body, file, or stdin
	description, err := getEpicDescription()
	if err != nil {
		return fmt.Errorf("failed to read description: %v", err)
	}

	result, err := createIssue(ctx, client, projectKey, epicCreateSummary, description, "Epic", epicCreatePriority, epicCreateLabels, "", nil, nil)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to create epic: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, result.Key))
	}

	return nil
}

func getEpicDescription() (string, error) {
	// Priority: file > description flag
	if epicCreateFile != "" {
		if epicCreateFile == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
		// Read from file
		data, err := os.ReadFile(epicCreateFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return epicCreateBody, nil
}
