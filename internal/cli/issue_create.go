package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/converter"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

// CreateResult represents the result of creating an issue.
type CreateResult struct {
	Key  string `json:"key"`
	ID   string `json:"id"`
	Self string `json:"self"`
}

// issueCreateRequest represents the request body for creating an issue.
type issueCreateRequest struct {
	Fields issueCreateFields `json:"fields"`
}

type issueCreateFields struct {
	Project     projectKey      `json:"project"`
	Summary     string          `json:"summary"`
	Description *converter.ADF  `json:"description,omitempty"`
	IssueType   issueTypeName   `json:"issuetype"`
	Priority    *priorityName   `json:"priority,omitempty"`
	Labels      []string        `json:"labels,omitempty"`
	Parent      *parentKey      `json:"parent,omitempty"`
	Components  []componentName `json:"components,omitempty"`
	FixVersions []versionName   `json:"fixVersions,omitempty"`
	Assignee    *assigneeField  `json:"assignee,omitempty"`
}

type assigneeField struct {
	AccountID string `json:"accountId"`
}

type parentKey struct {
	Key string `json:"key"`
}

type componentName struct {
	Name string `json:"name"`
}

type versionName struct {
	Name string `json:"name"`
}

type projectKey struct {
	Key string `json:"key"`
}

type issueTypeName struct {
	Name string `json:"name"`
}

type priorityName struct {
	Name string `json:"name"`
}

var (
	createSummary     string
	createBody        string
	createFile        string
	createType        string
	createPriority    string
	createLabels      []string
	createParent      string
	createComponents  []string
	createFixVersions []string
	createAssignee    string
)

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create issue",
	Long:  "Create a Jira issue. Requires -s for summary.",
	Example: `  ajira issue create -s "Fix login bug"                    # Create task
  ajira issue create -s "New feature" -t Story             # Create story
  ajira issue create -s "Bug" -d "Description in Markdown" # With description
  ajira issue create -s "From file" -f description.md      # Description from file
  ajira issue create -s "Subtask" --parent PROJ-50         # Create under parent/epic
  ajira issue create -s "Task" -C Backend,API              # With components
  ajira issue create -s "Task" --fix-version 1.0.0         # With fix version
  ajira issue create -s "Task" -a me                       # Assign to yourself
  ajira issue create -s "Task" -a user@example.com         # Assign by email
  ajira issue create -s "Task" -a unassigned               # Explicitly unassigned`,
	SilenceUsage: true,
	RunE:         runIssueCreate,
}

func init() {
	issueCreateCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "Issue summary (required)")
	issueCreateCmd.Flags().StringVarP(&createBody, "description", "d", "", "Issue description in Markdown")
	issueCreateCmd.Flags().StringVarP(&createFile, "file", "f", "", "Read description from file (use - for stdin)")
	issueCreateCmd.Flags().StringVarP(&createType, "type", "t", "Task", "Issue type (Task, Bug, Story, etc.)")
	issueCreateCmd.Flags().StringVarP(&createPriority, "priority", "P", "", "Issue priority")
	issueCreateCmd.Flags().StringSliceVar(&createLabels, "labels", nil, "Issue labels (comma-separated)")
	issueCreateCmd.Flags().StringVar(&createParent, "parent", "", "Parent issue or epic key")
	issueCreateCmd.Flags().StringSliceVarP(&createComponents, "component", "C", nil, "Component(s) (comma-separated)")
	issueCreateCmd.Flags().StringSliceVar(&createFixVersions, "fix-version", nil, "Fix version(s) (comma-separated)")
	issueCreateCmd.Flags().StringVarP(&createAssignee, "assignee", "a", "", "Assignee (me, email, account ID, or unassigned)")

	_ = issueCreateCmd.MarkFlagRequired("summary")

	issueCmd.AddCommand(issueCreateCmd)
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if createSummary == "" {
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

	// Validate issue type and priority before making the create request
	if err := jira.ValidateIssueType(ctx, client, projectKey, createType); err != nil {
		return err
	}
	if err := jira.ValidatePriority(ctx, client, createPriority); err != nil {
		return err
	}

	// Get description from body, file, or stdin
	description, err := getDescription()
	if err != nil {
		return fmt.Errorf("failed to read description: %w", err)
	}

	// Resolve assignee to accountId
	assigneeAccountID, err := resolveAssigneeInput(ctx, client, cfg.Email, createAssignee)
	if err != nil {
		return fmt.Errorf("failed to resolve assignee: %w", err)
	}

	result, err := createIssue(ctx, client, projectKey, createSummary, description, createType, createPriority, createLabels, createParent, createComponents, createFixVersions, assigneeAccountID)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to create issue: %w", err)
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

func getDescription() (string, error) {
	return readText(createFile, createBody)
}

func createIssue(ctx context.Context, client *api.Client, project, summary, description, issueType, priority string, labels []string, parent string, components, fixVersions []string, assignee *string) (*CreateResult, error) {
	req := issueCreateRequest{
		Fields: issueCreateFields{
			Project:   projectKey{Key: project},
			Summary:   summary,
			IssueType: issueTypeName{Name: issueType},
		},
	}

	// Convert Markdown description to ADF
	if description != "" {
		adf, err := converter.MarkdownToADF(description)
		if err != nil {
			return nil, fmt.Errorf("failed to convert description: %w", err)
		}
		req.Fields.Description = adf
	}

	if priority != "" {
		req.Fields.Priority = &priorityName{Name: priority}
	}

	if len(labels) > 0 {
		req.Fields.Labels = labels
	}

	if parent != "" {
		req.Fields.Parent = &parentKey{Key: parent}
	}

	if len(components) > 0 {
		for _, c := range components {
			req.Fields.Components = append(req.Fields.Components, componentName{Name: c})
		}
	}

	if len(fixVersions) > 0 {
		for _, v := range fixVersions {
			req.Fields.FixVersions = append(req.Fields.FixVersions, versionName{Name: v})
		}
	}

	if assignee != nil {
		req.Fields.Assignee = &assigneeField{AccountID: *assignee}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respBody, err := client.Post(ctx, "/issue", body)
	if err != nil {
		return nil, err
	}

	var result CreateResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
