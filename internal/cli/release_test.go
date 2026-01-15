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

func TestFetchAllReleases_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/project/TEST/version") {
			t.Errorf("expected /project/TEST/version path, got %s", r.URL.Path)
		}

		resp := releaseListResponse{
			IsLast: true,
			Values: []releaseValue{
				{
					ID:          "10001",
					Name:        "1.0.0",
					Description: "First release",
					Released:    true,
					ReleaseDate: "2024-01-15",
				},
				{
					ID:          "10002",
					Name:        "2.0.0",
					Description: "Second release",
					Released:    false,
					StartDate:   "2024-02-01",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	releases, err := fetchAllReleases(context.Background(), client, "TEST", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 2 {
		t.Fatalf("expected 2 releases, got %d", len(releases))
	}

	if releases[0].Name != "1.0.0" {
		t.Errorf("expected first release name '1.0.0', got %s", releases[0].Name)
	}
	if !releases[0].Released {
		t.Error("expected first release to be released")
	}
	if releases[1].Released {
		t.Error("expected second release to be unreleased")
	}
}

func TestFetchAllReleases_WithStatusFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "status=released") {
			t.Errorf("expected status=released in query, got %s", r.URL.RawQuery)
		}

		resp := releaseListResponse{
			IsLast: true,
			Values: []releaseValue{
				{
					ID:       "10001",
					Name:     "1.0.0",
					Released: true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	releases, err := fetchAllReleases(context.Background(), client, "TEST", "released", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 1 {
		t.Errorf("expected 1 release, got %d", len(releases))
	}
}

func TestFetchAllReleases_WithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := releaseListResponse{
			IsLast: false,
			Values: []releaseValue{
				{ID: "10001", Name: "1.0.0"},
				{ID: "10002", Name: "2.0.0"},
				{ID: "10003", Name: "3.0.0"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	releases, err := fetchAllReleases(context.Background(), client, "TEST", "", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 2 {
		t.Errorf("expected 2 releases (limited), got %d", len(releases))
	}
}

func TestFetchAllReleases_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := releaseListResponse{
			IsLast: true,
			Values: []releaseValue{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	releases, err := fetchAllReleases(context.Background(), client, "TEST", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 0 {
		t.Errorf("expected 0 releases, got %d", len(releases))
	}
}

func TestFetchAllReleases_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp releaseListResponse

		if callCount == 1 {
			resp = releaseListResponse{
				IsLast: false,
				Values: []releaseValue{
					{ID: "10001", Name: "1.0.0"},
					{ID: "10002", Name: "2.0.0"},
				},
			}
		} else {
			resp = releaseListResponse{
				IsLast: true,
				Values: []releaseValue{
					{ID: "10003", Name: "3.0.0"},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	releases, err := fetchAllReleases(context.Background(), client, "TEST", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(releases) != 3 {
		t.Errorf("expected 3 releases from pagination, got %d", len(releases))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestFetchAllReleases_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"Project not found"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := fetchAllReleases(context.Background(), client, "NOTEXIST", "", 0)
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
