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

func TestFetchFields_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/field") {
			t.Errorf("expected /field path, got %s", r.URL.Path)
		}

		resp := []fieldResponse{
			{
				ID:     "summary",
				Name:   "Summary",
				Custom: false,
				Schema: &struct {
					Type string `json:"type"`
				}{Type: "string"},
			},
			{
				ID:     "customfield_10001",
				Name:   "Story Points",
				Custom: true,
				Schema: &struct {
					Type string `json:"type"`
				}{Type: "number"},
			},
			{
				ID:     "status",
				Name:   "Status",
				Custom: false,
				Schema: &struct {
					Type string `json:"type"`
				}{Type: "status"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	fields, err := fetchFields(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}

	// Check system field
	var summaryField *FieldInfo
	for i := range fields {
		if fields[i].ID == "summary" {
			summaryField = &fields[i]
			break
		}
	}
	if summaryField == nil {
		t.Fatal("expected to find 'summary' field")
	}
	if summaryField.Name != "Summary" {
		t.Errorf("expected name 'Summary', got %s", summaryField.Name)
	}
	if summaryField.Custom {
		t.Error("expected summary to not be custom")
	}
	if summaryField.Type != "string" {
		t.Errorf("expected type 'string', got %s", summaryField.Type)
	}

	// Check custom field
	var customField *FieldInfo
	for i := range fields {
		if fields[i].ID == "customfield_10001" {
			customField = &fields[i]
			break
		}
	}
	if customField == nil {
		t.Fatal("expected to find 'customfield_10001' field")
	}
	if !customField.Custom {
		t.Error("expected customfield_10001 to be custom")
	}
	if customField.Type != "number" {
		t.Errorf("expected type 'number', got %s", customField.Type)
	}
}

func TestFetchFields_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]fieldResponse{})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	fields, err := fetchFields(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(fields))
	}
}

func TestFetchFields_NoSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []fieldResponse{
			{
				ID:     "description",
				Name:   "Description",
				Custom: false,
				Schema: nil, // No schema
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	fields, err := fetchFields(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}

	if fields[0].Type != "" {
		t.Errorf("expected empty type for field without schema, got %s", fields[0].Type)
	}
}

func TestFetchFields_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errorMessages": []string{"No permission to view fields"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := fetchFields(context.Background(), client)
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
