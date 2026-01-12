package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gcarthew/ajira/internal/api"
)

func TestBuildEpicListJQL_Basic(t *testing.T) {
	// Reset and set required state
	epicListStatus = ""
	epicListAssignee = ""
	epicListPriority = ""
	project = "GCP"

	jql := buildEpicListJQL()

	if !strings.Contains(jql, "project = GCP") {
		t.Errorf("expected project = GCP, got %q", jql)
	}
	if !strings.Contains(jql, "issuetype = Epic") {
		t.Errorf("expected issuetype = Epic, got %q", jql)
	}
	if !strings.Contains(jql, "ORDER BY updated DESC") {
		t.Errorf("expected ORDER BY updated DESC, got %q", jql)
	}
}

func TestBuildEpicListJQL_WithStatus(t *testing.T) {
	epicListStatus = "In Progress"
	epicListAssignee = ""
	epicListPriority = ""
	project = "GCP"

	jql := buildEpicListJQL()

	if !strings.Contains(jql, `status = "In Progress"`) {
		t.Errorf("expected status filter, got %q", jql)
	}
}

func TestBuildEpicListJQL_WithAssigneeMe(t *testing.T) {
	epicListStatus = ""
	epicListAssignee = "me"
	epicListPriority = ""
	project = "GCP"

	jql := buildEpicListJQL()

	if !strings.Contains(jql, "assignee = currentUser()") {
		t.Errorf("expected assignee = currentUser(), got %q", jql)
	}
}

func TestBuildEpicListJQL_WithAssigneeUnassigned(t *testing.T) {
	epicListStatus = ""
	epicListAssignee = "unassigned"
	epicListPriority = ""
	project = "GCP"

	jql := buildEpicListJQL()

	if !strings.Contains(jql, "assignee IS EMPTY") {
		t.Errorf("expected assignee IS EMPTY, got %q", jql)
	}
}

func TestBuildEpicListJQL_WithPriority(t *testing.T) {
	epicListStatus = ""
	epicListAssignee = ""
	epicListPriority = "Major"
	project = "GCP"

	jql := buildEpicListJQL()

	if !strings.Contains(jql, `priority = "Major"`) {
		t.Errorf("expected priority filter, got %q", jql)
	}
}

func TestBuildEpicListJQL_AllFilters(t *testing.T) {
	epicListStatus = "Done"
	epicListAssignee = "john@example.com"
	epicListPriority = "High"
	project = "GCP"

	jql := buildEpicListJQL()

	if !strings.Contains(jql, "project = GCP") {
		t.Errorf("expected project, got %q", jql)
	}
	if !strings.Contains(jql, "issuetype = Epic") {
		t.Errorf("expected issuetype, got %q", jql)
	}
	if !strings.Contains(jql, `status = "Done"`) {
		t.Errorf("expected status, got %q", jql)
	}
	if !strings.Contains(jql, `assignee = "john@example.com"`) {
		t.Errorf("expected assignee, got %q", jql)
	}
	if !strings.Contains(jql, `priority = "High"`) {
		t.Errorf("expected priority, got %q", jql)
	}
}

func TestAddIssuesToEpic_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/agile/1.0/epic/GCP-50/issue") {
			t.Errorf("expected /rest/agile/1.0/epic/GCP-50/issue path, got %s", r.URL.Path)
		}

		var req epicAddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.Issues) != 3 {
			t.Errorf("expected 3 issues, got %d", len(req.Issues))
		}
		if req.Issues[0] != "GCP-101" {
			t.Errorf("expected first issue 'GCP-101', got %s", req.Issues[0])
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := addIssuesToEpic(context.Background(), client, "GCP-50", []string{"GCP-101", "GCP-102", "GCP-103"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddIssuesToEpic_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"Epic not found"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := addIssuesToEpic(context.Background(), client, "GCP-9999", []string{"GCP-101"})
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

func TestRemoveIssuesFromEpic_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/agile/1.0/epic/none/issue") {
			t.Errorf("expected /rest/agile/1.0/epic/none/issue path, got %s", r.URL.Path)
		}

		var req epicRemoveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.Issues) != 2 {
			t.Errorf("expected 2 issues, got %d", len(req.Issues))
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := removeIssuesFromEpic(context.Background(), client, []string{"GCP-101", "GCP-102"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveIssuesFromEpic_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"Issue not found"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := removeIssuesFromEpic(context.Background(), client, []string{"INVALID-999"})
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
