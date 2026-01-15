package cli

import "testing"

func TestProjectURL(t *testing.T) {
	tests := []struct {
		baseURL    string
		projectKey string
		expected   string
	}{
		{
			baseURL:    "https://example.atlassian.net",
			projectKey: "TEST",
			expected:   "https://example.atlassian.net/browse/TEST",
		},
		{
			baseURL:    "https://jira.company.com",
			projectKey: "PROJ",
			expected:   "https://jira.company.com/browse/PROJ",
		},
	}

	for _, tt := range tests {
		result := ProjectURL(tt.baseURL, tt.projectKey)
		if result != tt.expected {
			t.Errorf("ProjectURL(%q, %q) = %q, want %q", tt.baseURL, tt.projectKey, result, tt.expected)
		}
	}
}

func TestIssueURL(t *testing.T) {
	tests := []struct {
		baseURL  string
		issueKey string
		expected string
	}{
		{
			baseURL:  "https://example.atlassian.net",
			issueKey: "TEST-123",
			expected: "https://example.atlassian.net/browse/TEST-123",
		},
		{
			baseURL:  "https://jira.company.com",
			issueKey: "PROJ-456",
			expected: "https://jira.company.com/browse/PROJ-456",
		},
	}

	for _, tt := range tests {
		result := IssueURL(tt.baseURL, tt.issueKey)
		if result != tt.expected {
			t.Errorf("IssueURL(%q, %q) = %q, want %q", tt.baseURL, tt.issueKey, result, tt.expected)
		}
	}
}
