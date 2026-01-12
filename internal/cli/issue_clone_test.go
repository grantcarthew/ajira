package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/jira"
)

// Test getSourceIssue function
func TestGetSourceIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/PROJ-123") {
			t.Errorf("expected /issue/PROJ-123 path, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"key": "PROJ-123",
			"fields": map[string]interface{}{
				"summary":     "Original issue",
				"description": map[string]interface{}{"type": "doc", "content": []interface{}{}},
				"issuetype":   map[string]string{"name": "Task"},
				"priority":    map[string]string{"name": "Major"},
				"labels":      []string{"bug", "urgent"},
				"project":     map[string]string{"key": "PROJ"},
				"assignee":    map[string]string{"accountId": "user123", "displayName": "Test User"},
				"reporter":    map[string]string{"accountId": "user456", "displayName": "Reporter User"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	source, err := getSourceIssue(context.Background(), client, "PROJ-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if source.Key != "PROJ-123" {
		t.Errorf("expected key 'PROJ-123', got %s", source.Key)
	}
	if source.Fields.Summary != "Original issue" {
		t.Errorf("expected summary 'Original issue', got %s", source.Fields.Summary)
	}
	if source.Fields.IssueType.Name != "Task" {
		t.Errorf("expected issue type 'Task', got %s", source.Fields.IssueType.Name)
	}
	if source.Fields.Priority.Name != "Major" {
		t.Errorf("expected priority 'Major', got %s", source.Fields.Priority.Name)
	}
	if len(source.Fields.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(source.Fields.Labels))
	}
}

// Test resolveCloneUser function
func TestResolveCloneUser_Override(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/user/search") {
			resp := []map[string]string{
				{"accountId": "resolved123", "displayName": "Resolved User"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	cfg := testConfig(server.URL)

	// Test with override
	sourceUser := &userField{AccountID: "original123", DisplayName: "Original User"}
	result, err := resolveCloneUser(context.Background(), client, cfg, "new@example.com", sourceUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "resolved123" {
		t.Errorf("expected 'resolved123', got %s", result)
	}
}

func TestResolveCloneUser_Unassigned(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	cfg := testConfig("http://unused")

	sourceUser := &userField{AccountID: "original123", DisplayName: "Original User"}
	result, err := resolveCloneUser(context.Background(), client, cfg, "unassigned", sourceUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for unassigned, got %s", result)
	}
}

func TestResolveCloneUser_NoOverride(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	cfg := testConfig("http://unused")

	// No override, use source user
	sourceUser := &userField{AccountID: "original123", DisplayName: "Original User"}
	result, err := resolveCloneUser(context.Background(), client, cfg, "", sourceUser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "original123" {
		t.Errorf("expected 'original123', got %s", result)
	}
}

func TestResolveCloneUser_NoOverrideNoSource(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	cfg := testConfig("http://unused")

	// No override, no source user
	result, err := resolveCloneUser(context.Background(), client, cfg, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

// Test ValidateLinkType function
func TestValidateLinkType_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"issueLinkTypes": []map[string]string{
				{"id": "1", "name": "Clones", "inward": "is cloned by", "outward": "clones"},
				{"id": "2", "name": "Blocks", "inward": "is blocked by", "outward": "blocks"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := jira.ValidateLinkType(context.Background(), client, "Clones")
	if err != nil {
		t.Errorf("expected no error for valid link type, got %v", err)
	}
}

func TestValidateLinkType_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"issueLinkTypes": []map[string]string{
				{"id": "1", "name": "Clones", "inward": "is cloned by", "outward": "clones"},
				{"id": "2", "name": "Blocks", "inward": "is blocked by", "outward": "blocks"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := jira.ValidateLinkType(context.Background(), client, "InvalidType")
	if err == nil {
		t.Error("expected error for invalid link type")
	}
	if !strings.Contains(err.Error(), "invalid link type") {
		t.Errorf("expected 'invalid link type' in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "Clones") {
		t.Errorf("expected valid options in error, got %v", err)
	}
}

func TestValidateLinkType_Empty(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	err := jira.ValidateLinkType(context.Background(), client, "")
	if err != nil {
		t.Errorf("expected no error for empty link type, got %v", err)
	}
}

func TestValidateLinkType_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"issueLinkTypes": []map[string]string{
				{"id": "1", "name": "Clones", "inward": "is cloned by", "outward": "clones"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))

	testCases := []string{"CLONES", "clones", "Clones", "cLoNeS"}
	for _, tc := range testCases {
		err := jira.ValidateLinkType(context.Background(), client, tc)
		if err != nil {
			t.Errorf("expected no error for %q, got %v", tc, err)
		}
	}
}

// Test createClonedIssue function
func TestCreateClonedIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue") {
			t.Errorf("expected /issue path, got %s", r.URL.Path)
		}

		// Verify request body
		var req cloneCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Fields.Summary != "Cloned issue" {
			t.Errorf("expected summary 'Cloned issue', got %s", req.Fields.Summary)
		}
		if req.Fields.Project.Key != "PROJ" {
			t.Errorf("expected project 'PROJ', got %s", req.Fields.Project.Key)
		}

		resp := map[string]string{
			"key":  "PROJ-456",
			"id":   "10001",
			"self": "https://example.atlassian.net/rest/api/3/issue/10001",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))

	req := &cloneCreateRequest{
		Fields: cloneCreateFields{
			Project:   projectKey{Key: "PROJ"},
			Summary:   "Cloned issue",
			IssueType: issueTypeName{Name: "Task"},
		},
	}

	result, err := createClonedIssue(context.Background(), client, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Key != "PROJ-456" {
		t.Errorf("expected key 'PROJ-456', got %s", result.Key)
	}
	if result.ID != "10001" {
		t.Errorf("expected id '10001', got %s", result.ID)
	}
}

// Test buildCloneRequest function
func TestBuildCloneRequest_WithOverrides(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	cfg := testConfig("http://unused")

	source := &cloneSourceResponse{
		Key: "PROJ-123",
		Fields: cloneSourceFields{
			Summary:   "Original summary",
			IssueType: &issueType{Name: "Task"},
			Priority:  &priorityField{Name: "Major"},
			Labels:    []string{"original"},
			Project:   &projectField{Key: "PROJ"},
		},
	}

	// Set overrides via package variables
	oldSummary := cloneSummary
	oldLabels := cloneLabels
	oldPriority := clonePriority
	defer func() {
		cloneSummary = oldSummary
		cloneLabels = oldLabels
		clonePriority = oldPriority
	}()

	cloneSummary = "Overridden summary"
	cloneLabels = []string{"new-label"}
	clonePriority = "Minor"

	req, err := buildCloneRequest(context.Background(), client, cfg, source, "PROJ", "Task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Fields.Summary != "Overridden summary" {
		t.Errorf("expected summary 'Overridden summary', got %s", req.Fields.Summary)
	}
	if req.Fields.Priority.Name != "Minor" {
		t.Errorf("expected priority 'Minor', got %s", req.Fields.Priority.Name)
	}
	if len(req.Fields.Labels) != 1 || req.Fields.Labels[0] != "new-label" {
		t.Errorf("expected labels ['new-label'], got %v", req.Fields.Labels)
	}
}

func TestBuildCloneRequest_NoOverrides(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	cfg := testConfig("http://unused")

	source := &cloneSourceResponse{
		Key: "PROJ-123",
		Fields: cloneSourceFields{
			Summary:   "Original summary",
			IssueType: &issueType{Name: "Bug"},
			Priority:  &priorityField{Name: "Critical"},
			Labels:    []string{"bug", "urgent"},
			Project:   &projectField{Key: "PROJ"},
			Assignee:  &userField{AccountID: "user123"},
			Reporter:  &userField{AccountID: "user456"},
		},
	}

	// Clear overrides
	oldSummary := cloneSummary
	oldLabels := cloneLabels
	oldPriority := clonePriority
	oldAssignee := cloneAssignee
	oldReporter := cloneReporter
	defer func() {
		cloneSummary = oldSummary
		cloneLabels = oldLabels
		clonePriority = oldPriority
		cloneAssignee = oldAssignee
		cloneReporter = oldReporter
	}()

	cloneSummary = ""
	cloneLabels = nil
	clonePriority = ""
	cloneAssignee = ""
	cloneReporter = ""

	req, err := buildCloneRequest(context.Background(), client, cfg, source, "PROJ", "Bug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Fields.Summary != "Original summary" {
		t.Errorf("expected summary 'Original summary', got %s", req.Fields.Summary)
	}
	if req.Fields.Priority.Name != "Critical" {
		t.Errorf("expected priority 'Critical', got %s", req.Fields.Priority.Name)
	}
	if len(req.Fields.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(req.Fields.Labels))
	}
	if req.Fields.Assignee.AccountID != "user123" {
		t.Errorf("expected assignee 'user123', got %s", req.Fields.Assignee.AccountID)
	}
	if req.Fields.Reporter.AccountID != "user456" {
		t.Errorf("expected reporter 'user456', got %s", req.Fields.Reporter.AccountID)
	}
}
