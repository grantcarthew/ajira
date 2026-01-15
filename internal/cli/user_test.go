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

func TestSearchUsers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/user/search") {
			t.Errorf("expected /user/search path, got %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "query=john") {
			t.Errorf("expected query=john in query string, got %s", r.URL.RawQuery)
		}

		resp := []userSearchResult{
			{
				AccountID:    "abc123",
				DisplayName:  "John Doe",
				EmailAddress: "john@example.com",
			},
			{
				AccountID:    "def456",
				DisplayName:  "John Smith",
				EmailAddress: "johns@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	users, err := searchUsers(context.Background(), client, "john", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].DisplayName != "John Doe" {
		t.Errorf("expected first user 'John Doe', got %s", users[0].DisplayName)
	}
	if users[0].AccountID != "abc123" {
		t.Errorf("expected first user accountID 'abc123', got %s", users[0].AccountID)
	}
	if users[0].EmailAddress != "john@example.com" {
		t.Errorf("expected first user email 'john@example.com', got %s", users[0].EmailAddress)
	}
}

func TestSearchUsers_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]userSearchResult{})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	users, err := searchUsers(context.Background(), client, "nobody", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestSearchUsers_WithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "maxResults=5") {
			t.Errorf("expected maxResults=5 in query string, got %s", r.URL.RawQuery)
		}

		resp := []userSearchResult{
			{AccountID: "user1", DisplayName: "User 1"},
			{AccountID: "user2", DisplayName: "User 2"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	users, err := searchUsers(context.Background(), client, "user", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestSearchUsers_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"Unauthorized"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := searchUsers(context.Background(), client, "test", 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}
