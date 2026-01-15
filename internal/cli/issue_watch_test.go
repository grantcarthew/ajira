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

func TestGetCurrentUserAccountID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/myself") {
			t.Errorf("expected /myself path, got %s", r.URL.Path)
		}

		resp := User{
			AccountID:    "abc123def456",
			DisplayName:  "Test User",
			EmailAddress: "test@example.com",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	accountID, err := getCurrentUserAccountID(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if accountID != "abc123def456" {
		t.Errorf("expected accountID 'abc123def456', got %s", accountID)
	}
}

func TestAddWatcher_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/watchers") {
			t.Errorf("expected /issue/TEST-123/watchers path, got %s", r.URL.Path)
		}

		// Verify the body is a quoted account ID
		var accountID string
		if err := json.NewDecoder(r.Body).Decode(&accountID); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if accountID != "user123" {
			t.Errorf("expected accountID 'user123', got %s", accountID)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := addWatcher(context.Background(), client, "TEST-123", "user123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveWatcher_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/watchers") {
			t.Errorf("expected /issue/TEST-123/watchers path, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("accountId") != "user123" {
			t.Errorf("expected accountId=user123 query param, got %s", r.URL.RawQuery)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := removeWatcher(context.Background(), client, "TEST-123", "user123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddWatcher_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"Issue does not exist"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := addWatcher(context.Background(), client, "NOTEXIST-999", "user123")
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

func TestRemoveWatcher_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"You do not have permission"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := removeWatcher(context.Background(), client, "TEST-123", "user123")
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
