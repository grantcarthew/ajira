package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/rest/api/3/") {
			t.Errorf("expected path to start with /rest/api/3/, got %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept: application/json, got %s", r.Header.Get("Accept"))
		}

		// Verify Basic Auth
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("expected Basic Auth header")
		}
		if username != "test@example.com" || password != "test-token" {
			t.Errorf("unexpected credentials: %s:%s", username, password)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL))
	body, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", result["status"])
	}
}

func TestClient_Post_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL))
	body, err := client.Post(context.Background(), "/test", []byte(`{"name":"test"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["id"] != "123" {
		t.Errorf("expected id '123', got %s", result["id"])
	}
}

func TestClient_APIError_WithMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist"},
			"errors":        map[string]string{},
		})
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL))
	_, err := client.Get(context.Background(), "/issue/INVALID-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
	if len(apiErr.Messages) != 1 || apiErr.Messages[0] != "Issue does not exist" {
		t.Errorf("unexpected messages: %v", apiErr.Messages)
	}
	if !strings.Contains(apiErr.Error(), "Issue does not exist") {
		t.Errorf("expected error message to contain 'Issue does not exist', got: %s", apiErr.Error())
	}
}

func TestClient_APIError_WithFieldErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{},
			"errors": map[string]string{
				"summary": "Summary is required",
			},
		})
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL))
	_, err := client.Post(context.Background(), "/issue", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", apiErr.StatusCode)
	}
	if apiErr.Errors["summary"] != "Summary is required" {
		t.Errorf("unexpected errors: %v", apiErr.Errors)
	}
	if !strings.Contains(apiErr.Error(), "summary") {
		t.Errorf("expected error message to contain 'summary', got: %s", apiErr.Error())
	}
}

func TestClient_APIError_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(testConfig(server.URL))
	_, err := client.Get(context.Background(), "/myself")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}

	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestClient_TrailingSlashInBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not have double slashes
		if strings.Contains(r.URL.Path, "//") {
			t.Errorf("path contains double slashes: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := testConfig(server.URL + "/") // trailing slash
	client := NewClient(cfg)
	_, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
