package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grantcarthew/ajira/internal/api"
	"github.com/grantcarthew/ajira/internal/config"
	"github.com/grantcarthew/ajira/internal/converter"
	"github.com/grantcarthew/ajira/internal/jira"
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

const vpWorkspaceID = "e10787b0-3a5d-437c-8957-2a01ca3085b4"

const (
	customerFieldKey      = "customfield_10385"
	customerFieldConfigID = "10589"
)

// customerRef is one Assets object reference on customfield_10385.
type customerRef struct {
	ID string `json:"id"`
}

// vpLocationLabelToObjectID maps first word of label (lowercase) to objectId.
// Derived from "dynamic" locations field 10352
var vpLocationLabelToObjectID = map[string]int{
	"dfw1": 2417, // DFW1 (ALN1 - Allen)
	"dfw2": 2416, // DFW2 (FTW1 - Fort Worth)
	"iad1": 2415, // IAD1 (STR1 - Sterling)
	"sea1": 2407, // SEA1 (PYL - Puyallup)
	"sea2": 2414, // SEA2 (QCY - Quincy)
	"slc1": 2419, // SLC1 (Lambda 127)
	"ord1": 2662,
}

type vpLocationEntry struct {
	WorkspaceID string `json:"workspaceId"`
	ID          string `json:"id"`
	ObjectID    string `json:"objectId"`
}

// vpTeamNameToARI maps team name (lowercase) to Jira team ARI.
// Derived from vpteams_customfield_10001.json - customfield_10001 options.
var vpTeamNameToARI = map[string]string{
	"cx":                          "b85ac7e6-2692-4d18-9d49-866aba0de144",
	"dfw1":                        "238adc48-14b3-4943-bda4-8f356fb61aab",
	"dfw2":                        "54f4242e-58ce-42a0-8250-97e19d8111b7",
	"iad1":                        "7be8c17d-cdcf-44e9-a4dd-ef27d044b679",
	"infraeng":                    "75eff1ff-122a-4d75-8162-7051acf92892",
	"infraops":                    "f0441640-07bf-4568-9f7b-aa2dff77914d",
	"itas":                        "0bf52e77-a2db-4da5-86dc-e5a6497da984",
	"managed services":            "11546b38-c1da-4471-b6b6-bf8beac5b265",
	"netops":                      "9871e921-413c-4d74-a247-03f2cc1bbdc7",
	"sea1":                        "8cf741cf-a2d1-48c8-a3ca-2a6971f91920",
	"sea2":                        "c03c627d-ee50-4c48-8870-c9b1e65b68d8",
	"technical support":           "861101cf-3f95-49ff-89c0-b7e082d70167",
	"vp studio":                   "b064317f-f7d2-4786-8c9a-afa3aca4cd29",
	"vpod":                        "1de0e346-3032-4613-99dd-26b998b83215",
	"dcops":                       "a12343f8-b77b-431e-a05b-1d61b0fa46af",
	"infra delivery tpm":          "f7e6d678-36e1-4a57-bd72-d7d8de75b944",
	"infrastructure automation":   "6ccded5f-f680-441b-824e-59fe4bff5169",
	"lightning ai platform":       "e77bf2f6-a796-494e-b53b-1e934e9df7f8",
	"observability":               "09fb2bbe-537e-412a-a144-324d52e84634",
	"ord1":                        "5eacb517-f48a-458e-8692-55a09b973a13",
	"people experience (px) team": "d5556679-fa62-44b2-ac77-09ebf0dc85c3",
	"qcy dc ops":                  "d3148843-83c9-44da-8b71-48cb0d336a79",
	"security team":               "9fd2c67f-608d-4e73-ae54-95b9fb5b7c6c",
	"slc1":                        "e093f999-324e-44fd-9d63-a660b5d5f9dc",
	"slc2":                        "0c458d92-adac-4ffe-9f97-31454a35e1b0",
	"software engineering":        "c8c1525d-2633-40c7-97d2-2c3c5406e246",
	"swe oncall":                  "c2d018b5-0ab5-4690-bb97-713a8679c0d8", // SWE Oncall; JSON also has "SWE oncall" (1a48d5be-667c-4d74-8cde-40dc75c03137)
}

type issueCreateFields struct {
	Project           projectKey        `json:"project"`
	Summary           string            `json:"summary"`
	Description       *converter.ADF    `json:"description,omitempty"`
	IssueType         issueTypeName     `json:"issuetype"`
	Priority          *priorityName     `json:"priority,omitempty"`
	Labels            []string          `json:"labels,omitempty"`
	Parent            *parentKey        `json:"parent,omitempty"`
	Components        []componentName   `json:"components,omitempty"`
	FixVersions       []versionName     `json:"fixVersions,omitempty"`
	Assignee          *assigneeField    `json:"assignee,omitempty"`
	VPLocationDynamic []vpLocationEntry `json:"customfield_10352,omitempty"`
	VPTeam            *string           `json:"customfield_10001,omitempty"`
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

type createIssueOptions struct {
	Project     string
	Summary     string
	Description string
	IssueType   string
	Priority    string
	Labels      []string
	Parent      string
	Components  []string
	FixVersions []string
	Assignee    string // resolved accountId; empty omits the assignee field from the request
	VPObjectID  int    // 0 omits customfield_10352
	VPTeamARI   string // empty omits customfield_10001
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
	createVPLocation  string
	createVPTeam      string
	createCustomer    string
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
  ajira issue create -s "Task" -a unassigned               # Explicitly unassigned
  ajira issue create -s "Task" --vp-team InfraOps          # With VP team
  ajira issue create -s "Task" --vp-customer internal      # With VP customer`,
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
	issueCreateCmd.Flags().StringVar(&createVPLocation, "vp-location", "", "VP location label, e.g. dfw1, sea1 (sets customfield_10352)")
	issueCreateCmd.Flags().StringVar(&createVPTeam, "vp-team", "", "VP team name (sets customfield_10001)")
	issueCreateCmd.Flags().StringVar(&createCustomer, "vp-customer", "", "VP customer name or object key (sets customfield_10385)")

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

	vpObjectID, err := resolveVPLocation(createVPLocation)
	if err != nil {
		return err
	}
	vpTeamARI, err := resolveVPTeam(createVPTeam)
	if err != nil {
		return err
	}

	opts := createIssueOptions{
		Project:     projectKey,
		Summary:     createSummary,
		Description: description,
		IssueType:   createType,
		Priority:    createPriority,
		Labels:      createLabels,
		Parent:      createParent,
		Components:  createComponents,
		FixVersions: createFixVersions,
		VPObjectID:  vpObjectID,
		VPTeamARI:   vpTeamARI,
	}
	if assigneeAccountID != nil {
		opts.Assignee = *assigneeAccountID
	}

	result, err := createIssue(ctx, client, opts)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			return fmt.Errorf("API error: %w", apiErr)
		}
		return fmt.Errorf("failed to create issue: %w", err)
	}

	if strings.TrimSpace(createCustomer) != "" {
		issueID := result.ID
		if issueID == "" {
			issueID, err = fetchIssueID(ctx, client, result.Key)
			if err != nil {
				return fmt.Errorf("created %s but failed to set customer: %w", result.Key, err)
			}
		}
		if err := setIssueCustomer(ctx, client, result.Key, issueID, createCustomer); err != nil {
			return fmt.Errorf("created %s but failed to set customer: %w", result.Key, err)
		}
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

func resolveVPLocation(label string) (int, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return 0, nil
	}
	parts := strings.Fields(label)
	if len(parts) == 0 {
		return 0, nil
	}
	key := strings.ToLower(parts[0])
	objectID, ok := vpLocationLabelToObjectID[key]
	if !ok {
		valid := make([]string, 0, len(vpLocationLabelToObjectID))
		for k := range vpLocationLabelToObjectID {
			valid = append(valid, k)
		}
		return 0, fmt.Errorf("unknown vp-location %q (valid: %s)", label, strings.Join(valid, ", "))
	}
	return objectID, nil
}

func resolveVPTeam(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}
	key := strings.ToLower(name)
	ari, ok := vpTeamNameToARI[key]
	if !ok {
		valid := make([]string, 0, len(vpTeamNameToARI))
		for k := range vpTeamNameToARI {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("unknown vp-team %q (valid: %s)", name, strings.Join(valid, ", "))
	}
	return ari, nil
}

func createIssue(ctx context.Context, client *api.Client, opts createIssueOptions) (*CreateResult, error) {
	req := issueCreateRequest{
		Fields: issueCreateFields{
			Project:   projectKey{Key: opts.Project},
			Summary:   opts.Summary,
			IssueType: issueTypeName{Name: opts.IssueType},
		},
	}

	// Convert Markdown description to ADF
	if opts.Description != "" {
		adf, err := converter.MarkdownToADF(opts.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to convert description: %w", err)
		}
		req.Fields.Description = adf
	}

	if opts.Priority != "" {
		req.Fields.Priority = &priorityName{Name: opts.Priority}
	}

	if len(opts.Labels) > 0 {
		req.Fields.Labels = opts.Labels
	}

	if opts.Parent != "" {
		req.Fields.Parent = &parentKey{Key: opts.Parent}
	}

	if len(opts.Components) > 0 {
		for _, c := range opts.Components {
			req.Fields.Components = append(req.Fields.Components, componentName{Name: c})
		}
	}

	if len(opts.FixVersions) > 0 {
		for _, v := range opts.FixVersions {
			req.Fields.FixVersions = append(req.Fields.FixVersions, versionName{Name: v})
		}
	}

	if opts.Assignee != "" {
		req.Fields.Assignee = &assigneeField{AccountID: opts.Assignee}
	}

	if opts.VPObjectID > 0 {
		objID := fmt.Sprintf("%d", opts.VPObjectID)
		req.Fields.VPLocationDynamic = []vpLocationEntry{{
			WorkspaceID: vpWorkspaceID,
			ID:          vpWorkspaceID + ":" + objID,
			ObjectID:    objID,
		}}
	}
	if opts.VPTeamARI != "" {
		ari := opts.VPTeamARI
		req.Fields.VPTeam = &ari
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
