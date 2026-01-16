package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
	"github.com/gcarthew/ajira/internal/jira"
	"github.com/spf13/cobra"
)

// CloneResult represents the result of cloning an issue.
type CloneResult struct {
	OriginalKey string `json:"originalKey"`
	ClonedKey   string `json:"clonedKey"`
	ClonedID    string `json:"clonedId"`
	Linked      bool   `json:"linked"`
	LinkType    string `json:"linkType,omitempty"`
}

// cloneSourceResponse matches the Jira issue API response for cloning.
type cloneSourceResponse struct {
	Key    string            `json:"key"`
	Fields cloneSourceFields `json:"fields"`
}

type cloneSourceFields struct {
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description"`
	IssueType   *issueType      `json:"issuetype"`
	Priority    *priorityField  `json:"priority"`
	Assignee    *userField      `json:"assignee"`
	Reporter    *userField      `json:"reporter"`
	Labels      []string        `json:"labels"`
	Project     *projectField   `json:"project"`
}

// cloneCreateRequest represents the request body for creating a cloned issue.
type cloneCreateRequest struct {
	Fields cloneCreateFields `json:"fields"`
}

type cloneCreateFields struct {
	Project     projectKey      `json:"project"`
	Summary     string          `json:"summary"`
	Description json.RawMessage `json:"description,omitempty"`
	IssueType   issueTypeName   `json:"issuetype"`
	Priority    *priorityName   `json:"priority,omitempty"`
	Labels      []string        `json:"labels,omitempty"`
	Assignee    *accountID      `json:"assignee,omitempty"`
	Reporter    *accountID      `json:"reporter,omitempty"`
}

type accountID struct {
	AccountID string `json:"accountId"`
}

var (
	cloneSummary  string
	cloneAssignee string
	cloneReporter string
	clonePriority string
	cloneType     string
	cloneLabels   []string
	cloneLink     string
	cloneLinkSet  bool
)

const defaultLinkType = "Clones"

var issueCloneCmd = &cobra.Command{
	Use:   "clone <issue-key>",
	Short: "Clone issue",
	Long:  "Clone an issue with the same fields. Override with -s, -a, -t, -P. Use --link to link to original.",
	Example: `  ajira issue clone PROJ-123                        # Clone with same fields
  ajira issue clone PROJ-123 -s "New summary"      # Override summary
  ajira issue clone PROJ-123 --link                # Link to original
  ajira issue clone PROJ-123 --link Duplicate      # Link with specific type
  ajira issue clone PROJ-123 -p OTHER              # Clone to different project`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	PreRun: func(cmd *cobra.Command, args []string) {
		cloneLinkSet = cmd.Flags().Changed("link")
	},
	RunE: runIssueClone,
}

func init() {
	issueCloneCmd.Flags().StringVarP(&cloneSummary, "summary", "s", "", "Override summary")
	issueCloneCmd.Flags().StringVarP(&cloneAssignee, "assignee", "a", "", "Override assignee (email, accountId, 'me', or 'unassigned')")
	issueCloneCmd.Flags().StringVarP(&cloneReporter, "reporter", "r", "", "Override reporter (email, accountId, or 'me')")
	issueCloneCmd.Flags().StringVarP(&clonePriority, "priority", "P", "", "Override priority")
	issueCloneCmd.Flags().StringVarP(&cloneType, "type", "t", "", "Override issue type")
	issueCloneCmd.Flags().StringSliceVarP(&cloneLabels, "labels", "L", nil, "Override labels (comma-separated)")
	issueCloneCmd.Flags().StringVar(&cloneLink, "link", "", "Link to original issue (default: Clones, or specify type)")

	issueCmd.AddCommand(issueCloneCmd)
}

func runIssueClone(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	sourceKey := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	client := api.NewClient(cfg)

	// Fetch source issue
	source, err := getSourceIssue(ctx, client, sourceKey)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to fetch source issue: %v", err)
	}

	// Determine target project
	targetProject := Project()
	if targetProject == "" {
		if source.Fields.Project != nil {
			targetProject = source.Fields.Project.Key
		} else {
			return fmt.Errorf("project is required (use -p flag or set JIRA_PROJECT)")
		}
	}

	// Determine issue type
	issueType := cloneType
	if issueType == "" && source.Fields.IssueType != nil {
		issueType = source.Fields.IssueType.Name
	}
	if issueType == "" {
		return fmt.Errorf("could not determine issue type from source issue")
	}

	// Validate issue type in target project (especially important for cross-project clone)
	if err := jira.ValidateIssueType(ctx, client, targetProject, issueType); err != nil {
		return fmt.Errorf("%v", err)
	}

	// Validate priority if overriding
	if clonePriority != "" {
		if err := jira.ValidatePriority(ctx, client, clonePriority); err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	// Determine link type if linking
	linkType := ""
	if cloneLinkSet {
		if cloneLink == "" {
			linkType = defaultLinkType
		} else {
			linkType = cloneLink
		}
		if err := jira.ValidateLinkType(ctx, client, linkType); err != nil {
			return fmt.Errorf("%v", err)
		}
	}

	// Build create request
	req, err := buildCloneRequest(ctx, client, cfg, source, targetProject, issueType)
	if err != nil {
		return fmt.Errorf("failed to build clone request: %v", err)
	}

	// Create the cloned issue
	result, err := createClonedIssue(ctx, client, req)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to create cloned issue: %v", err)
	}

	// Create link to original if requested
	linked := false
	if linkType != "" {
		err = createCloneLink(ctx, client, result.Key, sourceKey, linkType)
		if err != nil {
			// Non-fatal: issue was created, just couldn't link
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to link to original: %v\n", err)
		} else {
			linked = true
		}
	}

	if JSONOutput() {
		output := CloneResult{
			OriginalKey: sourceKey,
			ClonedKey:   result.Key,
			ClonedID:    result.ID,
			Linked:      linked,
		}
		if linked {
			output.LinkType = linkType
		}
		jsonOutput, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %v", err)
		}
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println(IssueURL(cfg.BaseURL, result.Key))
	}

	return nil
}

func getSourceIssue(ctx context.Context, client *api.Client, key string) (*cloneSourceResponse, error) {
	path := fmt.Sprintf("/issue/%s", key)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp cloneSourceResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

func buildCloneRequest(ctx context.Context, client *api.Client, cfg *config.Config, source *cloneSourceResponse, targetProject, issueType string) (*cloneCreateRequest, error) {
	// Start with source values
	summary := source.Fields.Summary
	if cloneSummary != "" {
		summary = cloneSummary
	}

	priority := ""
	if clonePriority != "" {
		priority = clonePriority
	} else if source.Fields.Priority != nil {
		priority = source.Fields.Priority.Name
	}

	labels := source.Fields.Labels
	if len(cloneLabels) > 0 {
		labels = cloneLabels
	}

	req := &cloneCreateRequest{
		Fields: cloneCreateFields{
			Project:     projectKey{Key: targetProject},
			Summary:     summary,
			Description: source.Fields.Description,
			IssueType:   issueTypeName{Name: issueType},
		},
	}

	// Set priority if available
	if priority != "" {
		req.Fields.Priority = &priorityName{Name: priority}
	}

	// Set labels if available
	if len(labels) > 0 {
		req.Fields.Labels = labels
	}

	// Handle assignee
	assigneeID, err := resolveCloneUser(ctx, client, cfg, cloneAssignee, source.Fields.Assignee)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve assignee: %w", err)
	}
	if assigneeID != "" {
		req.Fields.Assignee = &accountID{AccountID: assigneeID}
	}

	// Handle reporter
	reporterID, err := resolveCloneUser(ctx, client, cfg, cloneReporter, source.Fields.Reporter)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve reporter: %w", err)
	}
	if reporterID != "" {
		req.Fields.Reporter = &accountID{AccountID: reporterID}
	}

	return req, nil
}

// resolveCloneUser resolves a user for cloning.
// If override is provided, it resolves that. Otherwise uses the source user's accountId.
func resolveCloneUser(ctx context.Context, client *api.Client, cfg *config.Config, override string, sourceUser *userField) (string, error) {
	if override != "" {
		if strings.EqualFold(override, "unassigned") {
			return "", nil
		}
		if strings.EqualFold(override, "me") {
			return resolveUser(ctx, client, cfg.Email)
		}
		return resolveUser(ctx, client, override)
	}

	// Use source user's accountId if available
	if sourceUser != nil && sourceUser.AccountID != "" {
		return sourceUser.AccountID, nil
	}

	return "", nil
}

func createClonedIssue(ctx context.Context, client *api.Client, req *cloneCreateRequest) (*CreateResult, error) {
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

// createCloneLink creates a link from the cloned issue to the original.
// The cloned issue "clones" the original issue.
func createCloneLink(ctx context.Context, client *api.Client, clonedKey, originalKey, linkType string) error {
	// Use the existing createIssueLink function
	// clonedKey is the outward issue (shows "clones")
	// originalKey is the inward issue (shows "is cloned by")
	return createIssueLink(ctx, client, clonedKey, originalKey, linkType)
}
