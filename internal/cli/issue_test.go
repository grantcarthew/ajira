package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/config"
)

func testConfig(serverURL string) *config.Config {
	return &config.Config{
		BaseURL:     serverURL,
		Email:       "test@example.com",
		APIToken:    "test-token",
		HTTPTimeout: 5 * time.Second,
	}
}

// Test buildJQL function
func TestBuildJQL_EmptyFilters(t *testing.T) {
	// Reset global state
	issueListQuery = ""
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = ""
	project = ""

	jql := buildJQL()
	if jql != "" {
		t.Errorf("expected empty JQL, got %q", jql)
	}
}

func TestBuildJQL_WithProject(t *testing.T) {
	issueListQuery = ""
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = ""
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "project = TEST") {
		t.Errorf("expected JQL to contain 'project = TEST', got %q", jql)
	}
}

func TestBuildJQL_WithStatus(t *testing.T) {
	issueListQuery = ""
	issueListStatus = "In Progress"
	issueListType = ""
	issueListAssignee = ""
	project = ""

	jql := buildJQL()
	if !strings.Contains(jql, `status = "In Progress"`) {
		t.Errorf("expected JQL to contain status filter, got %q", jql)
	}
}

func TestBuildJQL_WithType(t *testing.T) {
	issueListQuery = ""
	issueListStatus = ""
	issueListType = "Bug"
	issueListAssignee = ""
	project = ""

	jql := buildJQL()
	if !strings.Contains(jql, `issuetype = "Bug"`) {
		t.Errorf("expected JQL to contain type filter, got %q", jql)
	}
}

func TestBuildJQL_WithAssignee(t *testing.T) {
	issueListQuery = ""
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = "john@example.com"
	project = ""

	jql := buildJQL()
	if !strings.Contains(jql, `assignee = "john@example.com"`) {
		t.Errorf("expected JQL to contain assignee filter, got %q", jql)
	}
}

func TestBuildJQL_Unassigned(t *testing.T) {
	issueListQuery = ""
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = "unassigned"
	project = ""

	jql := buildJQL()
	if !strings.Contains(jql, "assignee IS EMPTY") {
		t.Errorf("expected JQL to contain 'assignee IS EMPTY', got %q", jql)
	}
}

func TestBuildJQL_AssigneeMe(t *testing.T) {
	issueListQuery = ""
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = "me"
	project = ""

	jql := buildJQL()
	if !strings.Contains(jql, "assignee = currentUser()") {
		t.Errorf("expected JQL to contain 'assignee = currentUser()', got %q", jql)
	}
}

func TestBuildJQL_AssigneeMeCaseInsensitive(t *testing.T) {
	testCases := []string{"ME", "Me", "mE", "me"}

	for _, tc := range testCases {
		issueListQuery = ""
		issueListStatus = ""
		issueListType = ""
		issueListAssignee = tc
		project = ""

		jql := buildJQL()
		if !strings.Contains(jql, "assignee = currentUser()") {
			t.Errorf("assignee=%q: expected JQL to contain 'assignee = currentUser()', got %q", tc, jql)
		}
	}
}

func TestBuildJQL_UnassignedCaseInsensitive(t *testing.T) {
	testCases := []string{"UNASSIGNED", "Unassigned", "UnAssigned", "unassigned"}

	for _, tc := range testCases {
		issueListQuery = ""
		issueListStatus = ""
		issueListType = ""
		issueListAssignee = tc
		project = ""

		jql := buildJQL()
		if !strings.Contains(jql, "assignee IS EMPTY") {
			t.Errorf("assignee=%q: expected JQL to contain 'assignee IS EMPTY', got %q", tc, jql)
		}
	}
}

func TestBuildJQL_AssigneeMeWithOtherFilters(t *testing.T) {
	issueListQuery = ""
	issueListStatus = "In Progress"
	issueListType = "Bug"
	issueListAssignee = "me"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "project = TEST") {
		t.Errorf("expected JQL to contain 'project = TEST', got %q", jql)
	}
	if !strings.Contains(jql, `status = "In Progress"`) {
		t.Errorf("expected JQL to contain status filter, got %q", jql)
	}
	if !strings.Contains(jql, `issuetype = "Bug"`) {
		t.Errorf("expected JQL to contain type filter, got %q", jql)
	}
	if !strings.Contains(jql, "assignee = currentUser()") {
		t.Errorf("expected JQL to contain 'assignee = currentUser()', got %q", jql)
	}
	if !strings.Contains(jql, " AND ") {
		t.Errorf("expected AND in JQL, got %q", jql)
	}
}

func TestBuildJQL_RawQueryOverridesAssigneeMe(t *testing.T) {
	issueListQuery = "project = CUSTOM ORDER BY created"
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = "me"
	project = ""

	jql := buildJQL()
	if jql != "project = CUSTOM ORDER BY created" {
		t.Errorf("expected raw query to override 'me' filter, got %q", jql)
	}
	if strings.Contains(jql, "currentUser()") {
		t.Errorf("raw query should not contain currentUser(), got %q", jql)
	}
}

func TestBuildJQL_RawQueryOverrides(t *testing.T) {
	issueListQuery = "project = CUSTOM ORDER BY created"
	issueListStatus = "Done"
	issueListType = "Task"
	issueListAssignee = "jane@example.com"
	project = "TEST"

	jql := buildJQL()
	if jql != "project = CUSTOM ORDER BY created" {
		t.Errorf("expected raw query to override filters, got %q", jql)
	}
}

func TestBuildJQL_MultipleFilters(t *testing.T) {
	issueListQuery = ""
	issueListStatus = "Open"
	issueListType = "Story"
	issueListAssignee = ""
	project = "PROJ"

	jql := buildJQL()
	if !strings.Contains(jql, "project = PROJ") {
		t.Errorf("expected project in JQL, got %q", jql)
	}
	if !strings.Contains(jql, `status = "Open"`) {
		t.Errorf("expected status in JQL, got %q", jql)
	}
	if !strings.Contains(jql, `issuetype = "Story"`) {
		t.Errorf("expected type in JQL, got %q", jql)
	}
	if !strings.Contains(jql, " AND ") {
		t.Errorf("expected AND in JQL, got %q", jql)
	}
}

func TestBuildJQL_WithReporter(t *testing.T) {
	resetIssueListFlags()
	issueListReporter = "john@example.com"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, `reporter = "john@example.com"`) {
		t.Errorf("expected reporter filter, got %q", jql)
	}
}

func TestBuildJQL_WithReporterMe(t *testing.T) {
	resetIssueListFlags()
	issueListReporter = "me"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "reporter = currentUser()") {
		t.Errorf("expected reporter = currentUser(), got %q", jql)
	}
}

func TestBuildJQL_WithPriority(t *testing.T) {
	resetIssueListFlags()
	issueListPriority = "High"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, `priority = "High"`) {
		t.Errorf("expected priority filter, got %q", jql)
	}
}

func TestBuildJQL_WithLabels(t *testing.T) {
	resetIssueListFlags()
	issueListLabels = []string{"bug"}
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, `labels IN ("bug")`) {
		t.Errorf("expected labels IN filter, got %q", jql)
	}
}

func TestBuildJQL_WithMultipleLabels(t *testing.T) {
	resetIssueListFlags()
	issueListLabels = []string{"bug", "urgent", "backend"}
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, `labels IN ("bug", "urgent", "backend")`) {
		t.Errorf("expected labels IN filter with multiple labels, got %q", jql)
	}
}

func TestBuildJQL_WithWatching(t *testing.T) {
	resetIssueListFlags()
	issueListWatching = true
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "watcher = currentUser()") {
		t.Errorf("expected watcher = currentUser(), got %q", jql)
	}
}

func TestBuildJQL_WithOrderBy(t *testing.T) {
	resetIssueListFlags()
	issueListOrderBy = "created"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "ORDER BY created DESC") {
		t.Errorf("expected ORDER BY created DESC, got %q", jql)
	}
}

func TestBuildJQL_WithOrderByAndReverse(t *testing.T) {
	resetIssueListFlags()
	issueListOrderBy = "priority"
	issueListReverse = true
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "ORDER BY priority ASC") {
		t.Errorf("expected ORDER BY priority ASC, got %q", jql)
	}
}

func TestBuildJQL_DefaultOrderBy(t *testing.T) {
	resetIssueListFlags()
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "ORDER BY updated DESC") {
		t.Errorf("expected default ORDER BY updated DESC, got %q", jql)
	}
}

func TestBuildJQL_AllNewFilters(t *testing.T) {
	resetIssueListFlags()
	issueListReporter = "me"
	issueListPriority = "Medium"
	issueListLabels = []string{"feature"}
	issueListWatching = true
	issueListOrderBy = "key"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "reporter = currentUser()") {
		t.Errorf("expected reporter filter, got %q", jql)
	}
	if !strings.Contains(jql, `priority = "Medium"`) {
		t.Errorf("expected priority filter, got %q", jql)
	}
	if !strings.Contains(jql, `labels IN ("feature")`) {
		t.Errorf("expected labels filter, got %q", jql)
	}
	if !strings.Contains(jql, "watcher = currentUser()") {
		t.Errorf("expected watcher filter, got %q", jql)
	}
	if !strings.Contains(jql, "ORDER BY key DESC") {
		t.Errorf("expected ORDER BY key DESC, got %q", jql)
	}
}

// resetIssueListFlags resets all issue list flag variables to their zero values.
func resetIssueListFlags() {
	issueListQuery = ""
	issueListStatus = ""
	issueListType = ""
	issueListAssignee = ""
	issueListReporter = ""
	issueListPriority = ""
	issueListLabels = nil
	issueListWatching = false
	issueListOrderBy = ""
	issueListReverse = false
	issueListSprint = ""
	issueListEpic = ""
	project = ""
}

func TestBuildJQL_WithSprint(t *testing.T) {
	resetIssueListFlags()
	issueListSprint = "42"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "sprint = 42") {
		t.Errorf("expected sprint = 42, got %q", jql)
	}
}

func TestBuildJQL_WithEpic(t *testing.T) {
	resetIssueListFlags()
	issueListEpic = "GCP-50"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "parent = GCP-50") {
		t.Errorf("expected parent = GCP-50, got %q", jql)
	}
}

func TestBuildJQL_WithSprintAndEpic(t *testing.T) {
	resetIssueListFlags()
	issueListSprint = "42"
	issueListEpic = "GCP-50"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "sprint = 42") {
		t.Errorf("expected sprint = 42, got %q", jql)
	}
	if !strings.Contains(jql, "parent = GCP-50") {
		t.Errorf("expected parent = GCP-50, got %q", jql)
	}
	if !strings.Contains(jql, " AND ") {
		t.Errorf("expected AND in JQL, got %q", jql)
	}
}

func TestBuildJQL_SprintWithOtherFilters(t *testing.T) {
	resetIssueListFlags()
	issueListSprint = "42"
	issueListStatus = "In Progress"
	issueListAssignee = "me"
	project = "TEST"

	jql := buildJQL()
	if !strings.Contains(jql, "sprint = 42") {
		t.Errorf("expected sprint filter, got %q", jql)
	}
	if !strings.Contains(jql, `status = "In Progress"`) {
		t.Errorf("expected status filter, got %q", jql)
	}
	if !strings.Contains(jql, "assignee = currentUser()") {
		t.Errorf("expected assignee filter, got %q", jql)
	}
}

// Test searchIssues function
func TestSearchIssues_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/search/jql") {
			t.Errorf("expected /search/jql path, got %s", r.URL.Path)
		}

		resp := issueSearchResponse{
			IsLast: true,
			Issues: []issueValue{
				{
					Key: "TEST-1",
					Fields: issueFields{
						Summary:   "First issue",
						Status:    &statusField{Name: "Open"},
						IssueType: &issueType{Name: "Bug"},
						Priority:  &priorityField{Name: "High"},
						Assignee:  &userField{DisplayName: "John Doe"},
					},
				},
				{
					Key: "TEST-2",
					Fields: issueFields{
						Summary:   "Second issue",
						Status:    &statusField{Name: "Done"},
						IssueType: &issueType{Name: "Task"},
						Priority:  &priorityField{Name: "Low"},
						Assignee:  nil,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issues, err := searchIssues(context.Background(), client, "project = TEST", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issues))
	}

	if issues[0].Key != "TEST-1" {
		t.Errorf("expected key TEST-1, got %s", issues[0].Key)
	}
	if issues[0].Summary != "First issue" {
		t.Errorf("expected summary 'First issue', got %s", issues[0].Summary)
	}
	if issues[0].Status != "Open" {
		t.Errorf("expected status 'Open', got %s", issues[0].Status)
	}
	if issues[0].Type != "Bug" {
		t.Errorf("expected type 'Bug', got %s", issues[0].Type)
	}
	if issues[0].Assignee != "John Doe" {
		t.Errorf("expected assignee 'John Doe', got %s", issues[0].Assignee)
	}

	// Second issue has no assignee
	if issues[1].Assignee != "" {
		t.Errorf("expected empty assignee, got %s", issues[1].Assignee)
	}
}

func TestSearchIssues_WithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := issueSearchResponse{
			IsLast: false,
			Issues: []issueValue{
				{Key: "TEST-1", Fields: issueFields{Summary: "Issue 1"}},
				{Key: "TEST-2", Fields: issueFields{Summary: "Issue 2"}},
				{Key: "TEST-3", Fields: issueFields{Summary: "Issue 3"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issues, err := searchIssues(context.Background(), client, "project = TEST", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("expected 2 issues (limited), got %d", len(issues))
	}
}

func TestSearchIssues_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := issueSearchResponse{
			IsLast: true,
			Issues: []issueValue{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issues, err := searchIssues(context.Background(), client, "project = EMPTY", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

// Test getIssue function
func TestGetIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/TEST-123") {
			t.Errorf("expected /issue/TEST-123 path, got %s", r.URL.Path)
		}

		resp := issueDetailResponse{
			Key: "TEST-123",
			Fields: issueDetailFields{
				Summary:   "Test issue summary",
				Status:    &statusField{Name: "In Progress"},
				IssueType: &issueType{Name: "Story"},
				Priority:  &priorityField{Name: "Medium"},
				Assignee:  &userField{DisplayName: "Jane Smith"},
				Reporter:  &userField{DisplayName: "Bob Jones"},
				Created:   "2024-01-15T10:30:00.000+0000",
				Updated:   "2024-01-16T14:45:00.000+0000",
				Labels:    []string{"backend", "api"},
				Project:   &projectField{Key: "TEST", Name: "Test Project"},
				Description: json.RawMessage(`{
					"version": 1,
					"type": "doc",
					"content": [
						{
							"type": "paragraph",
							"content": [
								{"type": "text", "text": "This is the description."}
							]
						}
					]
				}`),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issue, err := getIssue(context.Background(), client, "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if issue.Key != "TEST-123" {
		t.Errorf("expected key TEST-123, got %s", issue.Key)
	}
	if issue.Summary != "Test issue summary" {
		t.Errorf("expected summary 'Test issue summary', got %s", issue.Summary)
	}
	if issue.Status != "In Progress" {
		t.Errorf("expected status 'In Progress', got %s", issue.Status)
	}
	if issue.Type != "Story" {
		t.Errorf("expected type 'Story', got %s", issue.Type)
	}
	if issue.Assignee != "Jane Smith" {
		t.Errorf("expected assignee 'Jane Smith', got %s", issue.Assignee)
	}
	if issue.Reporter != "Bob Jones" {
		t.Errorf("expected reporter 'Bob Jones', got %s", issue.Reporter)
	}
	if issue.Project != "TEST" {
		t.Errorf("expected project 'TEST', got %s", issue.Project)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(issue.Labels))
	}
	if issue.Description != "This is the description." {
		t.Errorf("expected description 'This is the description.', got %s", issue.Description)
	}
}

func TestGetIssue_NullDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := issueDetailResponse{
			Key: "TEST-456",
			Fields: issueDetailFields{
				Summary:     "No description issue",
				Description: json.RawMessage(`null`),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issue, err := getIssue(context.Background(), client, "TEST-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if issue.Description != "" {
		t.Errorf("expected empty description, got %s", issue.Description)
	}
}

// Test createIssue function
func TestCreateIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue") {
			t.Errorf("expected /issue path, got %s", r.URL.Path)
		}

		var req issueCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Fields.Project.Key != "TEST" {
			t.Errorf("expected project TEST, got %s", req.Fields.Project.Key)
		}
		if req.Fields.Summary != "New issue" {
			t.Errorf("expected summary 'New issue', got %s", req.Fields.Summary)
		}
		if req.Fields.IssueType.Name != "Task" {
			t.Errorf("expected type 'Task', got %s", req.Fields.IssueType.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(CreateResult{
			Key:  "TEST-999",
			ID:   "12345",
			Self: "https://example.atlassian.net/rest/api/3/issue/12345",
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	result, err := createIssue(context.Background(), client, "TEST", "New issue", "Description here", "Task", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Key != "TEST-999" {
		t.Errorf("expected key TEST-999, got %s", result.Key)
	}
	if result.ID != "12345" {
		t.Errorf("expected ID 12345, got %s", result.ID)
	}
}

func TestCreateIssue_WithPriorityAndLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req issueCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Fields.Priority == nil || req.Fields.Priority.Name != "High" {
			t.Errorf("expected priority 'High', got %v", req.Fields.Priority)
		}
		if len(req.Fields.Labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(req.Fields.Labels))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(CreateResult{Key: "TEST-1000"})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	result, err := createIssue(context.Background(), client, "TEST", "Issue with extras", "", "Bug", "High", []string{"urgent", "frontend"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Key != "TEST-1000" {
		t.Errorf("expected key TEST-1000, got %s", result.Key)
	}
}

// Test updateIssue function
func TestUpdateIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123") {
			t.Errorf("expected /issue/TEST-123 path, got %s", r.URL.Path)
		}

		var req issueEditRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Fields["summary"] != "Updated summary" {
			t.Errorf("expected summary 'Updated summary', got %v", req.Fields["summary"])
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := updateIssue(context.Background(), client, "TEST-123", map[string]any{"summary": "Updated summary"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test deleteIssue function
func TestDeleteIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123") {
			t.Errorf("expected /issue/TEST-123 path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := deleteIssue(context.Background(), client, "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test resolveUser function
func TestResolveUser_ByEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/user/search") {
			t.Errorf("expected /user/search path, got %s", r.URL.Path)
		}

		resp := userSearchResponse{
			{
				AccountID:    "abc123def456",
				DisplayName:  "John Doe",
				EmailAddress: "john@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	accountID, err := resolveUser(context.Background(), client, "john@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accountID != "abc123def456" {
		t.Errorf("expected accountID 'abc123def456', got %s", accountID)
	}
}

func TestResolveUser_DirectAccountID(t *testing.T) {
	// Long string without @ should be treated as direct accountId
	accountID, err := resolveUser(context.Background(), nil, "5f4dcc3b5aa765d61d8327deb882cf99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accountID != "5f4dcc3b5aa765d61d8327deb882cf99" {
		t.Errorf("expected direct accountID, got %s", accountID)
	}
}

func TestResolveUser_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(userSearchResponse{})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	accountID, err := resolveUser(context.Background(), client, "nobody@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accountID != "" {
		t.Errorf("expected empty accountID for not found user, got %s", accountID)
	}
}

// Test assignIssue function
func TestAssignIssue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/assignee") {
			t.Errorf("expected /issue/TEST-123/assignee path, got %s", r.URL.Path)
		}

		var req assigneeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.AccountID == nil || *req.AccountID != "user123" {
			t.Errorf("expected accountId 'user123', got %v", req.AccountID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	accountID := "user123"
	err := assignIssue(context.Background(), client, "TEST-123", &accountID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssignIssue_Unassign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req assigneeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.AccountID != nil {
			t.Errorf("expected null accountId for unassign, got %v", req.AccountID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := assignIssue(context.Background(), client, "TEST-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test getTransitions function
func TestGetTransitions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/transitions") {
			t.Errorf("expected /issue/TEST-123/transitions path, got %s", r.URL.Path)
		}

		resp := transitionsResponse{
			Transitions: []transition{
				{ID: "11", Name: "Start Progress", To: transitionStatus{Name: "In Progress"}},
				{ID: "21", Name: "Done", To: transitionStatus{Name: "Done"}},
				{ID: "31", Name: "Reopen", To: transitionStatus{Name: "Open"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	transitions, err := getTransitions(context.Background(), client, "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(transitions) != 3 {
		t.Fatalf("expected 3 transitions, got %d", len(transitions))
	}

	if transitions[0].ID != "11" {
		t.Errorf("expected first transition ID '11', got %s", transitions[0].ID)
	}
	if transitions[0].Name != "Start Progress" {
		t.Errorf("expected first transition name 'Start Progress', got %s", transitions[0].Name)
	}
	if transitions[0].To.Name != "In Progress" {
		t.Errorf("expected first transition to status 'In Progress', got %s", transitions[0].To.Name)
	}
}

// Test doTransition function
func TestDoTransition_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/transitions") {
			t.Errorf("expected /issue/TEST-123/transitions path, got %s", r.URL.Path)
		}

		var req transitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Transition.ID != "21" {
			t.Errorf("expected transition ID '21', got %s", req.Transition.ID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := doTransition(context.Background(), client, "TEST-123", "21")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test getComments function
func TestGetComments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/comment") {
			t.Errorf("expected /issue/TEST-123/comment path, got %s", r.URL.Path)
		}

		resp := commentsResponse{
			StartAt:    0,
			MaxResults: 5,
			Total:      2,
			Comments: []commentValue{
				{
					ID:      "10001",
					Author:  &userField{DisplayName: "Alice"},
					Created: "2024-01-16T14:30:00.000+0000",
					Body: json.RawMessage(`{
						"version": 1,
						"type": "doc",
						"content": [
							{
								"type": "paragraph",
								"content": [
									{"type": "text", "text": "First comment"}
								]
							}
						]
					}`),
				},
				{
					ID:      "10002",
					Author:  &userField{DisplayName: "Bob"},
					Created: "2024-01-16T15:00:00.000+0000",
					Body: json.RawMessage(`{
						"version": 1,
						"type": "doc",
						"content": [
							{
								"type": "paragraph",
								"content": [
									{"type": "text", "text": "Second comment"}
								]
							}
						]
					}`),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	comments, err := getComments(context.Background(), client, "TEST-123", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	if comments[0].ID != "10001" {
		t.Errorf("expected first comment ID '10001', got %s", comments[0].ID)
	}
	if comments[0].Author != "Alice" {
		t.Errorf("expected first comment author 'Alice', got %s", comments[0].Author)
	}
	if comments[0].Body != "First comment" {
		t.Errorf("expected first comment body 'First comment', got %s", comments[0].Body)
	}
}

// Test addComment function
func TestAddComment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/comment") {
			t.Errorf("expected /issue/TEST-123/comment path, got %s", r.URL.Path)
		}

		var req commentAddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Body == nil {
			t.Error("expected body to be non-nil")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(CommentResult{
			ID:      "10003",
			Self:    "https://example.atlassian.net/rest/api/3/issue/TEST-123/comment/10003",
			Created: "2024-01-16T16:00:00.000+0000",
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	result, err := addComment(context.Background(), client, "TEST-123", "This is a **bold** comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "10003" {
		t.Errorf("expected comment ID '10003', got %s", result.ID)
	}
}

// Test formatDateTime function
func TestFormatDateTime(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2024-01-16T14:30:00.000+0000", "2024-01-16 14:30"},
		{"2024-12-31T23:59:59.999+0000", "2024-12-31 23:59"},
		{"2024-01-01T00:00:00.000+0000", "2024-01-01 00:00"},
		{"short", "short"},
		{"", ""},
	}

	for _, tt := range tests {
		result := formatDateTime(tt.input)
		if result != tt.expected {
			t.Errorf("formatDateTime(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Test error handling
func TestSearchIssues_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Invalid JQL query"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := searchIssues(context.Background(), client, "invalid jql !!!", 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
}

func TestGetIssue_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist or you do not have permission to see it."},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := getIssue(context.Background(), client, "NOTEXIST-999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

// Test getCommentText helper
func TestGetCommentText_PositionalArg(t *testing.T) {
	commentBody = ""
	commentFile = ""

	text, err := getCommentText([]string{"TEST-123", "Comment from arg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Comment from arg" {
		t.Errorf("expected 'Comment from arg', got %q", text)
	}
}

func TestGetCommentText_BodyFlag(t *testing.T) {
	commentBody = "Comment from body flag"
	commentFile = ""

	text, err := getCommentText([]string{"TEST-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Comment from body flag" {
		t.Errorf("expected 'Comment from body flag', got %q", text)
	}
}

func TestGetCommentText_Empty(t *testing.T) {
	commentBody = ""
	commentFile = ""

	text, err := getCommentText([]string{"TEST-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}

func TestGetCommentText_FromFile(t *testing.T) {
	// Create a temporary file with comment content
	tmpFile, err := os.CreateTemp("", "comment-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "This is a comment from a file.\n\nWith multiple lines."
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	commentBody = ""
	commentFile = tmpFile.Name()

	text, err := getCommentText([]string{"TEST-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != content {
		t.Errorf("expected %q, got %q", content, text)
	}
}

func TestGetCommentText_FilePriority(t *testing.T) {
	// File should take priority over body flag and positional arg
	tmpFile, err := os.CreateTemp("", "comment-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	fileContent := "Content from file"
	if _, err := tmpFile.WriteString(fileContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	commentBody = "Content from body flag"
	commentFile = tmpFile.Name()

	text, err := getCommentText([]string{"TEST-123", "Content from arg"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != fileContent {
		t.Errorf("expected file content %q, got %q", fileContent, text)
	}
}

func TestGetCommentText_FileNotFound(t *testing.T) {
	commentBody = ""
	commentFile = "/nonexistent/path/to/file.md"

	_, err := getCommentText([]string{"TEST-123"})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

// Additional edge case tests
func TestSearchIssues_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp issueSearchResponse

		if callCount == 1 {
			resp = issueSearchResponse{
				NextPageToken: "page2token",
				IsLast:        false,
				Issues: []issueValue{
					{Key: "TEST-1", Fields: issueFields{Summary: "Issue 1"}},
					{Key: "TEST-2", Fields: issueFields{Summary: "Issue 2"}},
				},
			}
		} else {
			resp = issueSearchResponse{
				IsLast: true,
				Issues: []issueValue{
					{Key: "TEST-3", Fields: issueFields{Summary: "Issue 3"}},
					{Key: "TEST-4", Fields: issueFields{Summary: "Issue 4"}},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issues, err := searchIssues(context.Background(), client, "project = TEST", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issues) != 4 {
		t.Errorf("expected 4 issues from pagination, got %d", len(issues))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestSearchIssues_PaginationSafetyGuard(t *testing.T) {
	// Simulate an API that never returns IsLast: true (potential infinite loop)
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Always return more pages - simulates buggy API
		resp := issueSearchResponse{
			NextPageToken: "always-more",
			IsLast:        false,
			Issues: []issueValue{
				{Key: "TEST-" + string(rune('0'+callCount)), Fields: issueFields{Summary: "Issue"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	// No limit set - would loop forever without safety guard
	issues, err := searchIssues(context.Background(), client, "project = TEST", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have stopped at maxPages (100) iterations
	if callCount != 100 {
		t.Errorf("expected 100 API calls (maxPages), got %d", callCount)
	}
	if len(issues) != 100 {
		t.Errorf("expected 100 issues from pagination guard, got %d", len(issues))
	}
}

// Ensure we don't have import issues
var _ = context.Background
