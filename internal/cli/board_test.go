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

func TestListBoards_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/rest/agile/1.0/board") {
			t.Errorf("expected /rest/agile/1.0/board path, got %s", r.URL.Path)
		}

		resp := boardListResponse{
			IsLast: true,
			Values: []boardValue{
				{
					ID:   1342,
					Name: "GCP Board",
					Type: "scrum",
					Location: &boardLocation{
						ProjectKey: "GCP",
					},
				},
				{
					ID:   1455,
					Name: "Support Board",
					Type: "kanban",
					Location: &boardLocation{
						ProjectKey: "SUP",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	boards, err := listBoards(context.Background(), client, "", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(boards) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(boards))
	}

	if boards[0].ID != 1342 {
		t.Errorf("expected first board ID 1342, got %d", boards[0].ID)
	}
	if boards[0].Name != "GCP Board" {
		t.Errorf("expected first board name 'GCP Board', got %s", boards[0].Name)
	}
	if boards[0].Type != "scrum" {
		t.Errorf("expected first board type 'scrum', got %s", boards[0].Type)
	}
	if boards[0].Project != "GCP" {
		t.Errorf("expected first board project 'GCP', got %s", boards[0].Project)
	}
}

func TestListBoards_WithProject(t *testing.T) {
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.String()
		resp := boardListResponse{
			IsLast: true,
			Values: []boardValue{
				{ID: 1342, Name: "GCP Board", Type: "scrum"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := listBoards(context.Background(), client, "GCP", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedPath, "projectKeyOrId=GCP") {
		t.Errorf("expected projectKeyOrId=GCP in path, got %s", capturedPath)
	}
}

func TestListBoards_WithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := boardListResponse{
			IsLast: false,
			Values: []boardValue{
				{ID: 1, Name: "Board 1", Type: "scrum"},
				{ID: 2, Name: "Board 2", Type: "kanban"},
				{ID: 3, Name: "Board 3", Type: "scrum"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	boards, err := listBoards(context.Background(), client, "", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(boards) != 2 {
		t.Errorf("expected 2 boards (limited), got %d", len(boards))
	}
}

func TestListBoards_EmptyResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := boardListResponse{
			IsLast: true,
			Values: []boardValue{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	boards, err := listBoards(context.Background(), client, "", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(boards) != 0 {
		t.Errorf("expected 0 boards, got %d", len(boards))
	}
}

func TestListBoards_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"You do not have permission to view boards"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := listBoards(context.Background(), client, "", 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}
