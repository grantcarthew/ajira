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
	Project     projectKey     `json:"project"`
	Summary     string         `json:"summary"`
	Description *converter.ADF `json:"description,omitempty"`
	IssueType   issueTypeName  `json:"issuetype"`
	Priority    *priorityName  `json:"priority,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
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
)

var issueCreateCmd = &cobra.Command{
	Use:           "create",
	Short:         "Create a new issue",
	Long:          "Create a new Jira issue with the specified summary and optional description.",
	SilenceUsage:  true,
	RunE:          runIssueCreate,
}

func init() {
	issueCreateCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "Issue summary (required)")
	issueCreateCmd.Flags().StringVarP(&createBody, "description", "d", "", "Issue description in Markdown")
	issueCreateCmd.Flags().StringVarP(&createFile, "file", "f", "", "Read description from file (use - for stdin)")
	issueCreateCmd.Flags().StringVarP(&createType, "type", "t", "Task", "Issue type (Task, Bug, Story, etc.)")
	issueCreateCmd.Flags().StringVar(&createPriority, "priority", "", "Issue priority")
	issueCreateCmd.Flags().StringSliceVar(&createLabels, "labels", nil, "Issue labels (comma-separated)")

	issueCreateCmd.MarkFlagRequired("summary")

	issueCmd.AddCommand(issueCreateCmd)
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	if createSummary == "" {
		return Errorf("summary is required (use -s or --summary)")
	}

	projectKey := Project()
	if projectKey == "" {
		return Errorf("project is required (use -p flag or set JIRA_PROJECT)")
	}

	cfg, err := config.Load()
	if err != nil {
		return Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Validate issue type and priority before making the create request
	if err := jira.ValidateIssueType(client, projectKey, createType); err != nil {
		return Errorf("%v", err)
	}
	if err := jira.ValidatePriority(client, createPriority); err != nil {
		return Errorf("%v", err)
	}

	// Get description from body, file, or stdin
	description, err := getDescription()
	if err != nil {
		return Errorf("failed to read description: %v", err)
	}

	result, err := createIssue(client, projectKey, createSummary, description, createType, createPriority, createLabels)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.StatusCode == 401 {
				return Errorf("authentication failed (401)")
			}
			return Errorf("API error - %v", apiErr)
		}
		return Errorf("failed to create issue: %v", err)
	}

	if JSONOutput() {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, result.Key))
	}

	return nil
}

func getDescription() (string, error) {
	// Priority: file > description flag
	if createFile != "" {
		if createFile == "-" {
			// Read from stdin
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
		// Read from file
		data, err := os.ReadFile(createFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return createBody, nil
}

func createIssue(client *api.Client, project, summary, description, issueType, priority string, labels []string) (*CreateResult, error) {
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

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respBody, err := client.Post(context.Background(), "/issue", body)
	if err != nil {
		return nil, err
	}

	var result CreateResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}
