package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grantcarthew/ajira/internal/api"
)

// resetCommentListFlags resets all comment-list-related flag variables to their zero values.
func resetCommentListFlags() {
	commentListLimit = 5
}

func TestCommentList_CustomLimit(t *testing.T) {
	resetCommentListFlags()

	var receivedMaxResults string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMaxResults = r.URL.Query().Get("maxResults")
		resp := map[string]interface{}{
			"comments":   []interface{}{},
			"total":      0,
			"maxResults": 20,
			"startAt":    0,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, _, err := getComments(context.Background(), client, "TEST-123", 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedMaxResults != "20" {
		t.Errorf("expected maxResults=20, got %s", receivedMaxResults)
	}
}

func TestCommentList_APIError(t *testing.T) {
	resetCommentListFlags()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, _, err := getComments(context.Background(), client, "NOTEXIST-999", 5)
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

func TestCommentList_ReturnsTotal(t *testing.T) {
	resetCommentListFlags()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"comments": []map[string]interface{}{
				{
					"id":      "10001",
					"author":  map[string]string{"displayName": "Alice"},
					"created": "2024-01-20T14:30:00.000+0000",
					"body": map[string]interface{}{
						"type":    "doc",
						"version": 1,
						"content": []map[string]interface{}{
							{
								"type": "paragraph",
								"content": []map[string]interface{}{
									{"type": "text", "text": "A comment"},
								},
							},
						},
					},
				},
			},
			"total":      15,
			"maxResults": 5,
			"startAt":    0,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	comments, total, err := getComments(context.Background(), client, "TEST-123", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	if total != 15 {
		t.Errorf("expected total 15, got %d", total)
	}
}
