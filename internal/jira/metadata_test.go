package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// TestGetPriorities tests fetching priorities from the API.
func TestGetPriorities_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/priority") {
			t.Errorf("expected /priority path, got %s", r.URL.Path)
		}

		resp := []priorityResponse{
			{ID: "1", Name: "Highest", Description: "Critical priority"},
			{ID: "2", Name: "High", Description: "High priority"},
			{ID: "3", Name: "Medium", Description: "Medium priority"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	priorities, err := GetPriorities(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(priorities) != 3 {
		t.Fatalf("expected 3 priorities, got %d", len(priorities))
	}
	if priorities[0].Name != "Highest" {
		t.Errorf("expected first priority 'Highest', got %s", priorities[0].Name)
	}
}

func TestGetPriorities_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := GetPriorities(context.Background(), client)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPriorities_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := GetPriorities(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

// TestGetIssueTypes tests fetching issue types with both JSON formats.
func TestGetIssueTypes_ObjectFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/createmeta/TEST/issuetypes") {
			t.Errorf("expected /issue/createmeta/TEST/issuetypes path, got %s", r.URL.Path)
		}

		// Object wrapper format
		resp := map[string]interface{}{
			"issueTypes": []issueTypeResponse{
				{ID: "1", Name: "Bug", Description: "A bug", Subtask: false},
				{ID: "2", Name: "Task", Description: "A task", Subtask: false},
				{ID: "3", Name: "Sub-task", Description: "A subtask", Subtask: true},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	types, err := GetIssueTypes(context.Background(), client, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(types) != 3 {
		t.Fatalf("expected 3 issue types, got %d", len(types))
	}
	if types[0].Name != "Bug" {
		t.Errorf("expected first type 'Bug', got %s", types[0].Name)
	}
	if !types[2].Subtask {
		t.Error("expected third type to be a subtask")
	}
}

func TestGetIssueTypes_ArrayFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Direct array format (different API versions)
		resp := []issueTypeResponse{
			{ID: "1", Name: "Story", Description: "A story", Subtask: false},
			{ID: "2", Name: "Epic", Description: "An epic", Subtask: false},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	types, err := GetIssueTypes(context.Background(), client, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(types) != 2 {
		t.Fatalf("expected 2 issue types, got %d", len(types))
	}
	if types[0].Name != "Story" {
		t.Errorf("expected first type 'Story', got %s", types[0].Name)
	}
}

func TestGetIssueTypes_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"Project not found"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := GetIssueTypes(context.Background(), client, "NOTEXIST")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetIssueTypes_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := GetIssueTypes(context.Background(), client, "TEST")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected parse error, got: %v", err)
	}
}

// TestGetStatuses tests fetching statuses with deduplication.
func TestGetStatuses_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/project/TEST/statuses") {
			t.Errorf("expected /project/TEST/statuses path, got %s", r.URL.Path)
		}

		resp := []projectStatusesResponse{
			{
				ID:   "1",
				Name: "Bug",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "new", Name: "To Do"}},
					{ID: "2", Name: "In Progress", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "indeterminate", Name: "In Progress"}},
					{ID: "3", Name: "Done", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "done", Name: "Done"}},
				},
			},
			{
				ID:   "2",
				Name: "Task",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "new", Name: "To Do"}},
					{ID: "2", Name: "In Progress", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "indeterminate", Name: "In Progress"}},
					{ID: "3", Name: "Done", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "done", Name: "Done"}},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	statuses, err := GetStatuses(context.Background(), client, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should deduplicate: 3 unique statuses, not 6
	if len(statuses) != 3 {
		t.Fatalf("expected 3 deduplicated statuses, got %d", len(statuses))
	}

	// Verify the statuses
	names := make(map[string]bool)
	for _, s := range statuses {
		names[s.Name] = true
	}
	if !names["Open"] || !names["In Progress"] || !names["Done"] {
		t.Errorf("missing expected statuses, got: %v", names)
	}
}

func TestGetStatuses_Deduplication(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Multiple issue types with overlapping and unique statuses
		resp := []projectStatusesResponse{
			{
				ID:   "1",
				Name: "Bug",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "new", Name: "To Do"}},
					{ID: "4", Name: "Blocked", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "indeterminate", Name: "In Progress"}},
				},
			},
			{
				ID:   "2",
				Name: "Story",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "new", Name: "To Do"}},
					{ID: "5", Name: "In Review", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "indeterminate", Name: "In Progress"}},
				},
			},
			{
				ID:   "3",
				Name: "Epic",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "new", Name: "To Do"}},
					{ID: "3", Name: "Done", StatusCategory: struct {
						Key  string `json:"key"`
						Name string `json:"name"`
					}{Key: "done", Name: "Done"}},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	statuses, err := GetStatuses(context.Background(), client, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 4 unique statuses: Open (shared), Blocked, In Review, Done
	if len(statuses) != 4 {
		t.Fatalf("expected 4 deduplicated statuses, got %d", len(statuses))
	}

	ids := make(map[string]bool)
	for _, s := range statuses {
		ids[s.ID] = true
	}
	expected := []string{"1", "3", "4", "5"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("missing status with ID %s", id)
		}
	}
}

func TestGetStatuses_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]projectStatusesResponse{})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	statuses, err := GetStatuses(context.Background(), client, "TEST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestGetStatuses_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := GetStatuses(context.Background(), client, "TEST")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestValidatePriority tests priority validation.
func TestValidatePriority_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []priorityResponse{
			{ID: "1", Name: "Highest"},
			{ID: "2", Name: "High"},
			{ID: "3", Name: "Medium"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidatePriority(context.Background(), client, "High")
	if err != nil {
		t.Errorf("expected no error for valid priority, got: %v", err)
	}
}

func TestValidatePriority_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []priorityResponse{
			{ID: "1", Name: "Highest"},
			{ID: "2", Name: "High"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))

	testCases := []string{"HIGH", "high", "HiGh"}
	for _, tc := range testCases {
		err := ValidatePriority(context.Background(), client, tc)
		if err != nil {
			t.Errorf("expected no error for %q (case insensitive), got: %v", tc, err)
		}
	}
}

func TestValidatePriority_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []priorityResponse{
			{ID: "1", Name: "Highest"},
			{ID: "2", Name: "High"},
			{ID: "3", Name: "Medium"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidatePriority(context.Background(), client, "Critical")
	if err == nil {
		t.Fatal("expected error for invalid priority, got nil")
	}
	if !strings.Contains(err.Error(), "invalid priority") {
		t.Errorf("expected 'invalid priority' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Highest") || !strings.Contains(err.Error(), "High") {
		t.Errorf("expected valid options in error message, got: %v", err)
	}
}

func TestValidatePriority_Empty(t *testing.T) {
	// Empty priority should be valid (no validation needed)
	err := ValidatePriority(context.Background(), nil, "")
	if err != nil {
		t.Errorf("expected no error for empty priority, got: %v", err)
	}
}

func TestValidatePriority_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidatePriority(context.Background(), client, "High")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to fetch priorities") {
		t.Errorf("expected 'failed to fetch priorities' in error, got: %v", err)
	}
}

// TestValidateIssueType tests issue type validation.
func TestValidateIssueType_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"issueTypes": []issueTypeResponse{
				{ID: "1", Name: "Bug"},
				{ID: "2", Name: "Task"},
				{ID: "3", Name: "Story"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidateIssueType(context.Background(), client, "TEST", "Bug")
	if err != nil {
		t.Errorf("expected no error for valid issue type, got: %v", err)
	}
}

func TestValidateIssueType_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"issueTypes": []issueTypeResponse{
				{ID: "1", Name: "Bug"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))

	testCases := []string{"BUG", "bug", "BuG"}
	for _, tc := range testCases {
		err := ValidateIssueType(context.Background(), client, "TEST", tc)
		if err != nil {
			t.Errorf("expected no error for %q (case insensitive), got: %v", tc, err)
		}
	}
}

func TestValidateIssueType_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"issueTypes": []issueTypeResponse{
				{ID: "1", Name: "Bug"},
				{ID: "2", Name: "Task"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidateIssueType(context.Background(), client, "TEST", "Epic")
	if err == nil {
		t.Fatal("expected error for invalid issue type, got nil")
	}
	if !strings.Contains(err.Error(), "invalid issue type") {
		t.Errorf("expected 'invalid issue type' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Bug") || !strings.Contains(err.Error(), "Task") {
		t.Errorf("expected valid options in error message, got: %v", err)
	}
}

func TestValidateIssueType_Empty(t *testing.T) {
	err := ValidateIssueType(context.Background(), nil, "TEST", "")
	if err != nil {
		t.Errorf("expected no error for empty issue type, got: %v", err)
	}
}

func TestValidateIssueType_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidateIssueType(context.Background(), client, "NOTEXIST", "Bug")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to fetch issue types") {
		t.Errorf("expected 'failed to fetch issue types' in error, got: %v", err)
	}
}

// TestValidateStatus tests status validation.
func TestValidateStatus_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []projectStatusesResponse{
			{
				ID:   "1",
				Name: "Task",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open"},
					{ID: "2", Name: "In Progress"},
					{ID: "3", Name: "Done"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidateStatus(context.Background(), client, "TEST", "In Progress")
	if err != nil {
		t.Errorf("expected no error for valid status, got: %v", err)
	}
}

func TestValidateStatus_CaseInsensitive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []projectStatusesResponse{
			{
				ID:   "1",
				Name: "Task",
				Statuses: []statusResponse{
					{ID: "1", Name: "Done"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))

	testCases := []string{"DONE", "done", "DoNe"}
	for _, tc := range testCases {
		err := ValidateStatus(context.Background(), client, "TEST", tc)
		if err != nil {
			t.Errorf("expected no error for %q (case insensitive), got: %v", tc, err)
		}
	}
}

func TestValidateStatus_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []projectStatusesResponse{
			{
				ID:   "1",
				Name: "Task",
				Statuses: []statusResponse{
					{ID: "1", Name: "Open"},
					{ID: "2", Name: "Done"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidateStatus(context.Background(), client, "TEST", "Blocked")
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Open") || !strings.Contains(err.Error(), "Done") {
		t.Errorf("expected valid options in error message, got: %v", err)
	}
}

func TestValidateStatus_Empty(t *testing.T) {
	err := ValidateStatus(context.Background(), nil, "TEST", "")
	if err != nil {
		t.Errorf("expected no error for empty status, got: %v", err)
	}
}

func TestValidateStatus_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := ValidateStatus(context.Background(), client, "TEST", "Open")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to fetch statuses") {
		t.Errorf("expected 'failed to fetch statuses' in error, got: %v", err)
	}
}

// Ensure context is importable (used by metadata.go)
var _ = context.Background
