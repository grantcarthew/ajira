package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gcarthew/ajira/internal/api"
	"github.com/gcarthew/ajira/internal/jira"
)

// Test GetLinkTypes function
func TestGetLinkTypes_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issueLinkType") {
			t.Errorf("expected /issueLinkType path, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"issueLinkTypes": []map[string]string{
				{"id": "10000", "name": "Blocks", "inward": "is blocked by", "outward": "blocks"},
				{"id": "10001", "name": "Duplicate", "inward": "is duplicated by", "outward": "duplicates"},
				{"id": "10002", "name": "Relates", "inward": "relates to", "outward": "relates to"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	linkTypes, err := jira.GetLinkTypes(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(linkTypes) != 3 {
		t.Fatalf("expected 3 link types, got %d", len(linkTypes))
	}

	if linkTypes[0].Name != "Blocks" {
		t.Errorf("expected first link type name 'Blocks', got %s", linkTypes[0].Name)
	}
	if linkTypes[0].Outward != "blocks" {
		t.Errorf("expected outward 'blocks', got %s", linkTypes[0].Outward)
	}
	if linkTypes[0].Inward != "is blocked by" {
		t.Errorf("expected inward 'is blocked by', got %s", linkTypes[0].Inward)
	}
}

// Test findLinkType function
func TestFindLinkType_ExactMatch(t *testing.T) {
	linkTypes := []jira.LinkType{
		{ID: "1", Name: "Blocks", Inward: "is blocked by", Outward: "blocks"},
		{ID: "2", Name: "Duplicate", Inward: "is duplicated by", Outward: "duplicates"},
	}

	result := findLinkType(linkTypes, "Blocks")
	if result == nil {
		t.Fatal("expected to find Blocks link type")
	}
	if result.Name != "Blocks" {
		t.Errorf("expected name 'Blocks', got %s", result.Name)
	}
}

func TestFindLinkType_CaseInsensitive(t *testing.T) {
	linkTypes := []jira.LinkType{
		{ID: "1", Name: "Blocks", Inward: "is blocked by", Outward: "blocks"},
	}

	testCases := []string{"BLOCKS", "blocks", "Blocks", "bLoCkS"}
	for _, tc := range testCases {
		result := findLinkType(linkTypes, tc)
		if result == nil {
			t.Errorf("expected to find link type for %q", tc)
		}
	}
}

func TestFindLinkType_NotFound(t *testing.T) {
	linkTypes := []jira.LinkType{
		{ID: "1", Name: "Blocks", Inward: "is blocked by", Outward: "blocks"},
	}

	result := findLinkType(linkTypes, "Elephant")
	if result != nil {
		t.Errorf("expected nil for non-existent link type, got %v", result)
	}
}

// Test createIssueLink function
func TestCreateIssueLink_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issueLink") {
			t.Errorf("expected /issueLink path, got %s", r.URL.Path)
		}

		var req issueLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Note: API fields are swapped from our command syntax
		// Command: GCP-123 Blocks GCP-456 means "GCP-123 blocks GCP-456"
		// But Jira API requires: inwardIssue=GCP-123 (blocker), outwardIssue=GCP-456 (blocked)
		if req.OutwardIssue.Key != "GCP-456" {
			t.Errorf("expected outward issue GCP-456, got %s", req.OutwardIssue.Key)
		}
		if req.InwardIssue.Key != "GCP-123" {
			t.Errorf("expected inward issue GCP-123, got %s", req.InwardIssue.Key)
		}
		if req.Type.Name != "Blocks" {
			t.Errorf("expected link type 'Blocks', got %s", req.Type.Name)
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := createIssueLink(context.Background(), client, "GCP-123", "GCP-456", "Blocks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test getIssueLinks function
func TestGetIssueLinks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/GCP-123") {
			t.Errorf("expected /issue/GCP-123 path, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"fields": map[string]interface{}{
				"issuelinks": []map[string]interface{}{
					{
						"id": "12345",
						"type": map[string]string{
							"name":    "Blocks",
							"inward":  "is blocked by",
							"outward": "blocks",
						},
						"outwardIssue": map[string]interface{}{
							"key": "GCP-456",
							"fields": map[string]interface{}{
								"summary": "Blocked issue",
								"status":  map[string]string{"name": "Open"},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	links, err := getIssueLinks(context.Background(), client, "GCP-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}

	if links[0].ID != "12345" {
		t.Errorf("expected link ID '12345', got %s", links[0].ID)
	}
	if links[0].OutwardIssue == nil {
		t.Fatal("expected outward issue to be present")
	}
	if links[0].OutwardIssue.Key != "GCP-456" {
		t.Errorf("expected outward issue key 'GCP-456', got %s", links[0].OutwardIssue.Key)
	}
}

func TestGetIssueLinks_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"fields": map[string]interface{}{
				"issuelinks": []map[string]interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	links, err := getIssueLinks(context.Background(), client, "GCP-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(links) != 0 {
		t.Errorf("expected 0 links, got %d", len(links))
	}
}

// Test deleteIssueLink function
func TestDeleteIssueLink_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issueLink/12345") {
			t.Errorf("expected /issueLink/12345 path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := deleteIssueLink(context.Background(), client, "12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test createRemoteLink function
func TestCreateRemoteLink_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/GCP-123/remotelink") {
			t.Errorf("expected /issue/GCP-123/remotelink path, got %s", r.URL.Path)
		}

		var req remoteLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Object.URL != "https://github.com/org/repo/pull/42" {
			t.Errorf("expected URL 'https://github.com/org/repo/pull/42', got %s", req.Object.URL)
		}
		if req.Object.Title != "PR #42" {
			t.Errorf("expected title 'PR #42', got %s", req.Object.Title)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(remoteLinkResponse{
			ID:   10001,
			Self: "https://example.atlassian.net/rest/api/3/issue/GCP-123/remotelink/10001",
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	result, err := createRemoteLink(context.Background(), client, "GCP-123", "https://github.com/org/repo/pull/42", "PR #42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != 10001 {
		t.Errorf("expected ID 10001, got %d", result.ID)
	}
}

// Test issue view with links
func TestGetIssue_WithLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"key": "TEST-123",
			"fields": map[string]interface{}{
				"summary": "Test issue",
				"status":  map[string]string{"name": "Open"},
				"issuelinks": []map[string]interface{}{
					{
						"id": "12345",
						"type": map[string]string{
							"name":    "Blocks",
							"inward":  "is blocked by",
							"outward": "blocks",
						},
						"outwardIssue": map[string]interface{}{
							"key": "TEST-456",
							"fields": map[string]interface{}{
								"summary": "Blocked issue summary",
								"status":  map[string]string{"name": "In Progress"},
							},
						},
					},
					{
						"id": "12346",
						"type": map[string]string{
							"name":    "Duplicate",
							"inward":  "is duplicated by",
							"outward": "duplicates",
						},
						"inwardIssue": map[string]interface{}{
							"key": "TEST-789",
							"fields": map[string]interface{}{
								"summary": "Duplicate issue",
								"status":  map[string]string{"name": "Done"},
							},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	issue, err := getIssue(context.Background(), client, "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(issue.Links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(issue.Links))
	}

	// Check first link (outward)
	if issue.Links[0].Direction != "blocks" {
		t.Errorf("expected direction 'blocks', got %s", issue.Links[0].Direction)
	}
	if issue.Links[0].Key != "TEST-456" {
		t.Errorf("expected key 'TEST-456', got %s", issue.Links[0].Key)
	}
	if issue.Links[0].Status != "In Progress" {
		t.Errorf("expected status 'In Progress', got %s", issue.Links[0].Status)
	}

	// Check second link (inward)
	if issue.Links[1].Direction != "is duplicated by" {
		t.Errorf("expected direction 'is duplicated by', got %s", issue.Links[1].Direction)
	}
	if issue.Links[1].Key != "TEST-789" {
		t.Errorf("expected key 'TEST-789', got %s", issue.Links[1].Key)
	}
}

// Test error handling
func TestCreateIssueLink_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := createIssueLink(context.Background(), client, "NOTEXIST-123", "NOTEXIST-456", "Blocks")
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
