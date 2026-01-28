package cli

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gcarthew/ajira/internal/api"
)

// resetAttachmentFlags resets all attachment-related flag variables to their zero values.
func resetAttachmentFlags() {
	downloadOutput = ""
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		// Bytes
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		// Kilobytes
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{2048, "2 KB"},
		{10240, "10 KB"},
		{45056, "44 KB"},
		{102400, "100 KB"},
		{1047552, "1023 KB"},
		// Megabytes
		{1048576, "1 MB"},
		{1258291, "1.19 MB"},
		{10485760, "10 MB"},
		{104857600, "100 MB"},
		{1072693248, "1023 MB"},
		// Gigabytes
		{1073741824, "1 GB"},
		{1288490189, "1.2 GB"},
		{10737418240, "10 GB"},
		{107374182400, "100 GB"},
	}

	for _, tt := range tests {
		result := FormatFileSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
		}
	}
}

func TestGetAttachments_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/issue/TEST-123") {
			t.Errorf("expected /issue/TEST-123 path, got %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "fields=attachment") {
			t.Errorf("expected fields=attachment query, got %s", r.URL.RawQuery)
		}

		resp := map[string]interface{}{
			"fields": map[string]interface{}{
				"attachment": []map[string]interface{}{
					{
						"id":       "10001",
						"filename": "screenshot.png",
						"size":     45056,
						"mimeType": "image/png",
						"author":   map[string]string{"displayName": "John Doe"},
						"created":  "2024-01-20T14:30:00.000+0000",
						"content":  "https://example.atlassian.net/rest/api/3/attachment/content/10001",
					},
					{
						"id":       "10002",
						"filename": "document.pdf",
						"size":     1258291,
						"mimeType": "application/pdf",
						"author":   map[string]string{"displayName": "Jane Smith"},
						"created":  "2024-01-21T09:15:00.000+0000",
						"content":  "https://example.atlassian.net/rest/api/3/attachment/content/10002",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	attachments, err := getAttachments(context.Background(), client, "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(attachments))
	}

	if attachments[0].ID != "10001" {
		t.Errorf("expected ID '10001', got %s", attachments[0].ID)
	}
	if attachments[0].Filename != "screenshot.png" {
		t.Errorf("expected filename 'screenshot.png', got %s", attachments[0].Filename)
	}
	if attachments[0].Size != 45056 {
		t.Errorf("expected size 45056, got %d", attachments[0].Size)
	}
	if attachments[0].Author != "John Doe" {
		t.Errorf("expected author 'John Doe', got %s", attachments[0].Author)
	}

	if attachments[1].ID != "10002" {
		t.Errorf("expected ID '10002', got %s", attachments[1].ID)
	}
	if attachments[1].MimeType != "application/pdf" {
		t.Errorf("expected mimeType 'application/pdf', got %s", attachments[1].MimeType)
	}
}

func TestGetAttachments_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"fields": map[string]interface{}{
				"attachment": []interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	attachments, err := getAttachments(context.Background(), client, "TEST-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(attachments))
	}
}

func TestGetAttachments_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := getAttachments(context.Background(), client, "NOTEXIST-999")
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

func TestGetAttachmentMeta_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/attachment/10001") {
			t.Errorf("expected /attachment/10001 path, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"id":       10001, // Note: API returns number, not string
			"filename": "screenshot.png",
			"size":     45056,
			"mimeType": "image/png",
			"author":   map[string]string{"displayName": "John Doe"},
			"created":  "2024-01-20T14:30:00.000+0000",
			"content":  "https://example.atlassian.net/rest/api/3/attachment/content/10001",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	meta, err := getAttachmentMeta(context.Background(), client, "10001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.ID.String() != "10001" {
		t.Errorf("expected ID '10001', got %s", meta.ID)
	}
	if meta.Filename != "screenshot.png" {
		t.Errorf("expected filename 'screenshot.png', got %s", meta.Filename)
	}
	if meta.Size != 45056 {
		t.Errorf("expected size 45056, got %d", meta.Size)
	}
}

func TestGetAttachmentMeta_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Attachment not found"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err := getAttachmentMeta(context.Background(), client, "99999")
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

func TestDeleteAttachment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/attachment/10001") {
			t.Errorf("expected /attachment/10001 path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := deleteAttachment(context.Background(), client, "10001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAttachment_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Attachment not found"},
		})
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	err := deleteAttachment(context.Background(), client, "99999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUploadAttachments_Success(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test-upload-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Test file content for upload"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/issue/TEST-123/attachments") {
			t.Errorf("expected /issue/TEST-123/attachments path, got %s", r.URL.Path)
		}

		// Verify X-Atlassian-Token header
		if r.Header.Get("X-Atlassian-Token") != "no-check" {
			t.Errorf("expected X-Atlassian-Token: no-check, got %s", r.Header.Get("X-Atlassian-Token"))
		}

		// Verify Content-Type is multipart/form-data
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", contentType)
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to get form file: %v", err)
		}
		defer file.Close()

		// Verify filename
		expectedFilename := filepath.Base(tmpFile.Name())
		if header.Filename != expectedFilename {
			t.Errorf("expected filename %q, got %q", expectedFilename, header.Filename)
		}

		// Verify content
		uploadedContent, _ := io.ReadAll(file)
		if string(uploadedContent) != content {
			t.Errorf("expected content %q, got %q", content, string(uploadedContent))
		}

		// Return successful response
		resp := []map[string]interface{}{
			{
				"id":       "10003",
				"filename": header.Filename,
				"size":     len(content),
				"mimeType": "text/plain",
				"author":   map[string]string{"displayName": "Test User"},
				"created":  "2024-01-28T10:00:00.000+0000",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	attachments, err := uploadAttachments(context.Background(), client, "TEST-123", []string{tmpFile.Name()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].ID != "10003" {
		t.Errorf("expected ID '10003', got %s", attachments[0].ID)
	}
}

func TestUploadAttachments_MultipleFiles(t *testing.T) {
	// Create two temporary test files
	tmpFile1, err := os.CreateTemp("", "test-upload1-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file 1: %v", err)
	}
	defer os.Remove(tmpFile1.Name())
	tmpFile1.WriteString("Content 1")
	tmpFile1.Close()

	tmpFile2, err := os.CreateTemp("", "test-upload2-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file 2: %v", err)
	}
	defer os.Remove(tmpFile2.Name())
	tmpFile2.WriteString("Content 2")
	tmpFile2.Close()

	fileCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		// Count files in multipart form
		files := r.MultipartForm.File["file"]
		fileCount = len(files)

		// Return successful response for each file
		var resp []map[string]interface{}
		for i, fh := range files {
			resp = append(resp, map[string]interface{}{
				"id":       "1000" + string(rune('3'+i)),
				"filename": fh.Filename,
				"size":     9,
				"mimeType": "text/plain",
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	attachments, err := uploadAttachments(context.Background(), client, "TEST-123", []string{tmpFile1.Name(), tmpFile2.Name()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fileCount != 2 {
		t.Errorf("expected 2 files in multipart form, got %d", fileCount)
	}
	if len(attachments) != 2 {
		t.Errorf("expected 2 attachments in response, got %d", len(attachments))
	}
}

func TestUploadAttachments_FileNotFound(t *testing.T) {
	client := api.NewClient(testConfig("http://unused"))
	_, err := uploadAttachments(context.Background(), client, "TEST-123", []string{"/nonexistent/file.txt"})
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open") {
		t.Errorf("expected 'failed to open' error, got %v", err)
	}
}

func TestUploadAttachments_APIError413(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-upload-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("test content")
	tmpFile.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	}))
	defer server.Close()

	client := api.NewClient(testConfig(server.URL))
	_, err = uploadAttachments(context.Background(), client, "TEST-123", []string{tmpFile.Name()})
	if err == nil {
		t.Fatal("expected error for 413, got nil")
	}

	apiErr, ok := err.(*api.APIError)
	if !ok {
		t.Fatalf("expected *api.APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", apiErr.StatusCode)
	}
	// Check for user-friendly message
	if len(apiErr.Messages) == 0 || !strings.Contains(apiErr.Messages[0], "size limit") {
		t.Errorf("expected 'size limit' in error message, got %v", apiErr.Messages)
	}
}

// Test issue view includes attachments
func TestGetIssue_WithAttachments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := issueDetailResponse{
			Key: "TEST-123",
			Fields: issueDetailFields{
				Summary: "Issue with attachments",
				Attachment: []attachmentDetail{
					{
						ID:       "10001",
						Filename: "doc.pdf",
						Size:     1024,
						MimeType: "application/pdf",
						Author:   &userField{DisplayName: "Test User"},
						Created:  "2024-01-28T10:00:00.000+0000",
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

	if len(issue.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(issue.Attachments))
	}
	if issue.Attachments[0].ID != "10001" {
		t.Errorf("expected attachment ID '10001', got %s", issue.Attachments[0].ID)
	}
	if issue.Attachments[0].Filename != "doc.pdf" {
		t.Errorf("expected filename 'doc.pdf', got %s", issue.Attachments[0].Filename)
	}
	if issue.Attachments[0].Author != "Test User" {
		t.Errorf("expected author 'Test User', got %s", issue.Attachments[0].Author)
	}
}
