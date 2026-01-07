package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/converter"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

// issueEditRequest represents the request body for editing an issue.
type issueEditRequest struct {
	Fields map[string]any `json:"fields"`
}

var (
	editSummary  string
	editBody     string
	editFile     string
	editType     string
	editPriority string
	editLabels   []string
)

var issueEditCmd = &cobra.Command{
	Use:           "edit <issue-key>",
	Short:         "Edit an existing issue",
	Long:          "Update fields of an existing Jira issue.",
	Example: `  ajira issue edit PROJ-123 -s "New summary"       # Update summary
  ajira issue edit PROJ-123 -d "New description"   # Update description
  ajira issue edit PROJ-123 -t Bug --priority High # Change type and priority`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	RunE:          runIssueEdit,
}

func init() {
	issueEditCmd.Flags().StringVarP(&editSummary, "summary", "s", "", "New issue summary")
	issueEditCmd.Flags().StringVarP(&editBody, "description", "d", "", "New description in Markdown")
	issueEditCmd.Flags().StringVarP(&editFile, "file", "f", "", "Read description from file (use - for stdin)")
	issueEditCmd.Flags().StringVarP(&editType, "type", "t", "", "New issue type")
	issueEditCmd.Flags().StringVar(&editPriority, "priority", "", "New priority")
	issueEditCmd.Flags().StringSliceVar(&editLabels, "labels", nil, "New labels (comma-separated, replaces existing)")

	issueCmd.AddCommand(issueEditCmd)
}

func runIssueEdit(cmd *cobra.Command, args []string) error {
	issueKey := args[0]

	// Check if any field was provided
	hasChanges := editSummary != "" || editBody != "" || editFile != "" ||
		editType != "" || editPriority != "" || editLabels != nil

	if !hasChanges {
		return Errorf("no fields to update (use --summary, --description, --file, --type, --priority, or --labels)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Extract project key from issue key for validation
	projectKey := extractProjectKey(issueKey)

	// Validate issue type and priority before making the update request
	if err := jira.ValidateIssueType(client, projectKey, editType); err != nil {
		return Errorf("%v", err)
	}
	if err := jira.ValidatePriority(client, editPriority); err != nil {
		return Errorf("%v", err)
	}

	// Build fields to update
	fields := make(map[string]any)

	if editSummary != "" {
		fields["summary"] = editSummary
	}

	// Get description from file or description flag
	description := editBody
	if editFile != "" {
		if editFile == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return Errorf("failed to read stdin: %v", err)
			}
			description = string(data)
		} else {
			data, err := os.ReadFile(editFile)
			if err != nil {
				return Errorf("failed to read file: %v", err)
			}
			description = string(data)
		}
	}

	if description != "" {
		adf, err := converter.MarkdownToADF(description)
		if err != nil {
			return Errorf("failed to convert description: %v", err)
		}
		fields["description"] = adf
	}

	if editType != "" {
		fields["issuetype"] = map[string]string{"name": editType}
	}

	if editPriority != "" {
		fields["priority"] = map[string]string{"name": editPriority}
	}

	if editLabels != nil {
		fields["labels"] = editLabels
	}

	err = updateIssue(client, issueKey, fields)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to update issue: %v", err)
	}

	if JSONOutput() {
		result := map[string]string{"key": issueKey, "status": "updated"}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, issueKey))
	}

	return nil
}

func updateIssue(client *api.Client, key string, fields map[string]any) error {
	req := issueEditRequest{Fields: fields}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s", key)
	_, err = client.Put(context.Background(), path, body)
	return err
}
