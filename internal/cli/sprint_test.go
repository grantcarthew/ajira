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

func TestListSprints_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/rest/agile/1.0/board/1342/sprint") {
			t.Errorf("expected /rest/agile/1.0/board/1342/sprint path, got %s", r.URL.Path)
		}

		resp := sprintListResponse{
			IsLast: true,
			Values: []sprintValue{
				{
					ID:        42,
					Name:      "Sprint 23",
					State:     "active",
					StartDate: "2026-01-06T00:00:00.000Z",
					EndDate:   "2026-01-20T00:00:00.000Z",
					Goal:      "Finish auth",
				},
				{
					ID:        43,
					Name:      "Sprint 24",
					State:     "future",
					StartDate: "2026-01-20T00:00:00.000Z",
					EndDate:   "2026-02-03T00:00:00.000Z",
					Goal:      "",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	sprints, err := listSprints(context.Background(), client, "1342", "", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sprints) != 2 {
		t.Fatalf("expected 2 sprints, got %d", len(sprints))
	}

	if sprints[0].ID != 42 {
		t.Errorf("expected first sprint ID 42, got %d", sprints[0].ID)
	}
	if sprints[0].Name != "Sprint 23" {
		t.Errorf("expected first sprint name 'Sprint 23', got %s", sprints[0].Name)
	}
	if sprints[0].State != "active" {
		t.Errorf("expected first sprint state 'active', got %s", sprints[0].State)
	}
	if sprints[0].Goal != "Finish auth" {
		t.Errorf("expected first sprint goal 'Finish auth', got %s", sprints[0].Goal)
	}
}

func TestListSprints_WithStateFilter(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		resp := sprintListResponse{
			IsLast: true,
			Values: []sprintValue{
				{ID: 42, Name: "Sprint 23", State: "active"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := listSprints(context.Background(), client, "1342", "active", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedPath, "state=active") {
		t.Errorf("expected state=active in path, got %s", capturedPath)
	}
}

func TestListSprints_WithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := sprintListResponse{
			IsLast: false,
			Values: []sprintValue{
				{ID: 1, Name: "Sprint 1", State: "closed"},
				{ID: 2, Name: "Sprint 2", State: "closed"},
				{ID: 3, Name: "Sprint 3", State: "active"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	sprints, err := listSprints(context.Background(), client, "1342", "", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sprints) != 2 {
		t.Errorf("expected 2 sprints (limited), got %d", len(sprints))
	}
}

func TestListSprints_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := sprintListResponse{
			IsLast: true,
			Values: []sprintValue{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	sprints, err := listSprints(context.Background(), client, "1342", "", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sprints) != 0 {
		t.Errorf("expected 0 sprints, got %d", len(sprints))
	}
}

func TestListSprints_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Board does not exist"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := listSprints(context.Background(), client, "9999", "", 50)
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

func TestAddIssuesToSprint_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/rest/agile/1.0/sprint/42/issue") {
			t.Errorf("expected /rest/agile/1.0/sprint/42/issue path, got %s", r.URL.Path)
		}

		var req sprintAddRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.Issues) != 3 {
			t.Errorf("expected 3 issues, got %d", len(req.Issues))
		}
		if req.Issues[0] != "GCP-123" {
			t.Errorf("expected first issue 'GCP-123', got %s", req.Issues[0])
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := addIssuesToSprint(context.Background(), client, "42", []string{"GCP-123", "GCP-124", "GCP-125"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddIssuesToSprint_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Cannot move issues to closed sprint"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := addIssuesToSprint(context.Background(), client, "42", []string{"GCP-123"})
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

func TestFormatSprintDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-01-06T00:00:00.000Z", "2026-01-06"},
		{"2026-12-31T23:59:59.999Z", "2026-12-31"},
		{"", "-"},
		{"short", "short"},
	}

	for _, tt := range tests {
		result := formatSprintDate(tt.input)
		if result != tt.expected {
			t.Errorf("formatSprintDate(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestColorSprintState(t *testing.T) {
	// Just verify no panics and returns non-empty strings
	tests := []string{"active", "future", "closed", "unknown"}
	for _, state := range tests {
		result := colorSprintState(state, state)
		if result == "" {
			t.Errorf("colorSprintState(%q) returned empty string", state)
		}
	}
}
