package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/converter"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

// issueEditRequest represents the request body for editing an issue.
type issueEditRequest struct {
	Fields map[string]any `json:"fields,omitempty"`
	Update map[string]any `json:"update,omitempty"`
}

var (
	editSummary           string
	editBody              string
	editFile              string
	editType              string
	editPriority          string
	editLabels            []string
	editParent            string
	editAddLabels         []string
	editRemoveLabels      []string
	editComponents        []string
	editAddComponents     []string
	editRemoveComponents  []string
	editFixVersions       []string
	editAddFixVersions    []string
	editRemoveFixVersions []string
)

var issueEditCmd = &cobra.Command{
	Use:   "edit <issue-key>",
	Short: "Edit an existing issue",
	Long:  "Update fields of an existing Jira issue.",
	Example: `  ajira issue edit PROJ-123 -s "New summary"          # Update summary
  ajira issue edit PROJ-123 -d "New description"      # Update description
  ajira issue edit PROJ-123 -t Bug --priority High    # Change type and priority
  ajira issue edit PROJ-123 --parent PROJ-50          # Set parent/epic
  ajira issue edit PROJ-123 --parent none             # Remove parent
  ajira issue edit PROJ-123 --add-labels urgent       # Add label
  ajira issue edit PROJ-123 --add-component Frontend  # Add component
  ajira issue edit PROJ-123 --add-fix-version 1.1.0   # Add fix version`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         runIssueEdit,
}

func init() {
	issueEditCmd.Flags().StringVarP(&editSummary, "summary", "s", "", "New issue summary")
	issueEditCmd.Flags().StringVarP(&editBody, "description", "d", "", "New description in Markdown")
	issueEditCmd.Flags().StringVarP(&editFile, "file", "f", "", "Read description from file (use - for stdin)")
	issueEditCmd.Flags().StringVarP(&editType, "type", "t", "", "New issue type")
	issueEditCmd.Flags().StringVarP(&editPriority, "priority", "P", "", "New priority")
	issueEditCmd.Flags().StringSliceVar(&editLabels, "labels", nil, "New labels (comma-separated, replaces existing)")
	issueEditCmd.Flags().StringVar(&editParent, "parent", "", "Parent issue or epic key (use none/remove/clear/unset to remove)")
	issueEditCmd.Flags().Lookup("parent").NoOptDefVal = ""
	issueEditCmd.Flags().StringSliceVar(&editAddLabels, "add-labels", nil, "Add label(s) without replacing existing")
	issueEditCmd.Flags().StringSliceVar(&editRemoveLabels, "remove-labels", nil, "Remove specific label(s)")
	issueEditCmd.Flags().StringSliceVarP(&editComponents, "component", "C", nil, "Replace all components")
	issueEditCmd.Flags().StringSliceVar(&editAddComponents, "add-component", nil, "Add component(s)")
	issueEditCmd.Flags().StringSliceVar(&editRemoveComponents, "remove-component", nil, "Remove component(s)")
	issueEditCmd.Flags().StringSliceVar(&editFixVersions, "fix-version", nil, "Replace all fix versions")
	issueEditCmd.Flags().StringSliceVar(&editAddFixVersions, "add-fix-version", nil, "Add fix version(s)")
	issueEditCmd.Flags().StringSliceVar(&editRemoveFixVersions, "remove-fix-version", nil, "Remove fix version(s)")

	issueCmd.AddCommand(issueEditCmd)
}

// isParentRemovalKeyword checks if the value is a keyword indicating parent removal.
func isParentRemovalKeyword(value string) bool {
	keywords := []string{"none", "remove", "clear", "unset"}
	for _, k := range keywords {
		if strings.EqualFold(value, k) {
			return true
		}
	}
	return false
}

func runIssueEdit(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	issueKey := args[0]

	// Check if parent flag was explicitly provided
	parentChanged := cmd.Flags().Changed("parent")

	// Check if any field was provided
	hasChanges := editSummary != "" || editBody != "" || editFile != "" ||
		editType != "" || editPriority != "" || editLabels != nil || parentChanged ||
		editAddLabels != nil || editRemoveLabels != nil ||
		editComponents != nil || editAddComponents != nil || editRemoveComponents != nil ||
		editFixVersions != nil || editAddFixVersions != nil || editRemoveFixVersions != nil

	if !hasChanges {
		return fmt.Errorf("no fields to update")
	}

	// Check for conflicting flags
	if editLabels != nil && (editAddLabels != nil || editRemoveLabels != nil) {
		return fmt.Errorf("cannot use --labels with --add-labels or --remove-labels")
	}
	if editComponents != nil && (editAddComponents != nil || editRemoveComponents != nil) {
		return fmt.Errorf("cannot use --component with --add-component or --remove-component")
	}
	if editFixVersions != nil && (editAddFixVersions != nil || editRemoveFixVersions != nil) {
		return fmt.Errorf("cannot use --fix-version with --add-fix-version or --remove-fix-version")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Extract project key from issue key for validation
	projectKey := extractProjectKey(issueKey)

	// Validate issue type and priority before making the update request
	if err := jira.ValidateIssueType(ctx, client, projectKey, editType); err != nil {
		return fmt.Errorf("%v", err)
	}
	if err := jira.ValidatePriority(ctx, client, editPriority); err != nil {
		return fmt.Errorf("%v", err)
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
				return fmt.Errorf("failed to read stdin: %v", err)
			}
			description = string(data)
		} else {
			data, err := os.ReadFile(editFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %v", err)
			}
			description = string(data)
		}
	}

	if description != "" {
		adf, err := converter.MarkdownToADF(description)
		if err != nil {
			return fmt.Errorf("failed to convert description: %v", err)
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

	if parentChanged {
		if editParent == "" || isParentRemovalKeyword(editParent) {
			// Remove parent
			fields["parent"] = nil
		} else {
			// Set parent
			fields["parent"] = map[string]string{"key": editParent}
		}
	}

	// Replace components (fields)
	if editComponents != nil {
		var components []map[string]string
		for _, c := range editComponents {
			components = append(components, map[string]string{"name": c})
		}
		fields["components"] = components
	}

	// Replace fix versions (fields)
	if editFixVersions != nil {
		var versions []map[string]string
		for _, v := range editFixVersions {
			versions = append(versions, map[string]string{"name": v})
		}
		fields["fixVersions"] = versions
	}

	// Build update map for add/remove operations
	var update map[string]any
	needsUpdate := editAddLabels != nil || editRemoveLabels != nil ||
		editAddComponents != nil || editRemoveComponents != nil ||
		editAddFixVersions != nil || editRemoveFixVersions != nil

	if needsUpdate {
		update = make(map[string]any)
	}

	// Label add/remove
	if editAddLabels != nil || editRemoveLabels != nil {
		var labelOps []map[string]string
		for _, label := range editAddLabels {
			labelOps = append(labelOps, map[string]string{"add": label})
		}
		for _, label := range editRemoveLabels {
			labelOps = append(labelOps, map[string]string{"remove": label})
		}
		update["labels"] = labelOps
	}

	// Component add/remove
	if editAddComponents != nil || editRemoveComponents != nil {
		var componentOps []map[string]any
		for _, c := range editAddComponents {
			componentOps = append(componentOps, map[string]any{"add": map[string]string{"name": c}})
		}
		for _, c := range editRemoveComponents {
			componentOps = append(componentOps, map[string]any{"remove": map[string]string{"name": c}})
		}
		update["components"] = componentOps
	}

	// Fix version add/remove
	if editAddFixVersions != nil || editRemoveFixVersions != nil {
		var versionOps []map[string]any
		for _, v := range editAddFixVersions {
			versionOps = append(versionOps, map[string]any{"add": map[string]string{"name": v}})
		}
		for _, v := range editRemoveFixVersions {
			versionOps = append(versionOps, map[string]any{"remove": map[string]string{"name": v}})
		}
		update["fixVersions"] = versionOps
	}

	err = updateIssue(ctx, client, issueKey, fields, update)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to update issue: %v", err)
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

func updateIssue(ctx context.Context, client *api.Client, key string, fields, update map[string]any) error {
	req := issueEditRequest{
		Fields: fields,
		Update: update,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/issue/%s", key)
	_, err = client.Put(ctx, path, body)
	return err
}
