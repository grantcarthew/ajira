package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func clearEnv() {
	os.Unsetenv("JIRA_BASE_URL")
	os.Unsetenv("JIRA_EMAIL")
	os.Unsetenv("JIRA_API_TOKEN")
	os.Unsetenv("ATLASSIAN_BASE_URL")
	os.Unsetenv("ATLASSIAN_EMAIL")
	os.Unsetenv("ATLASSIAN_API_TOKEN")
	os.Unsetenv("JIRA_PROJECT")
}

func setValidEnv() {
	os.Setenv("JIRA_BASE_URL", "https://example.atlassian.net")
	os.Setenv("JIRA_EMAIL", "user@example.com")
	os.Setenv("JIRA_API_TOKEN", "test-token")
}

func TestLoad_Success(t *testing.T) {
	clearEnv()
	setValidEnv()
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.BaseURL != "https://example.atlassian.net" {
		t.Errorf("expected BaseURL 'https://example.atlassian.net', got %q", cfg.BaseURL)
	}
	if cfg.Email != "user@example.com" {
		t.Errorf("expected Email 'user@example.com', got %q", cfg.Email)
	}
	if cfg.APIToken != "test-token" {
		t.Errorf("expected APIToken 'test-token', got %q", cfg.APIToken)
	}
	if cfg.HTTPTimeout != 30*time.Second {
		t.Errorf("expected HTTPTimeout 30s, got %v", cfg.HTTPTimeout)
	}
}

func TestLoad_WithProject(t *testing.T) {
	clearEnv()
	setValidEnv()
	os.Setenv("JIRA_PROJECT", "TEST")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Project != "TEST" {
		t.Errorf("expected Project 'TEST', got %q", cfg.Project)
	}
}

func TestLoad_MissingBaseURL(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_EMAIL", "user@example.com")
	os.Setenv("JIRA_API_TOKEN", "test-token")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JIRA_BASE_URL") {
		t.Errorf("expected error to mention JIRA_BASE_URL, got: %v", err)
	}
}

func TestLoad_MissingEmail(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_BASE_URL", "https://example.atlassian.net")
	os.Setenv("JIRA_API_TOKEN", "test-token")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JIRA_EMAIL") {
		t.Errorf("expected error to mention JIRA_EMAIL, got: %v", err)
	}
}

func TestLoad_MissingToken(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_BASE_URL", "https://example.atlassian.net")
	os.Setenv("JIRA_EMAIL", "user@example.com")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JIRA_API_TOKEN") {
		t.Errorf("expected error to mention JIRA_API_TOKEN, got: %v", err)
	}
}

func TestLoad_AtlassianTokenFallback(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_BASE_URL", "https://example.atlassian.net")
	os.Setenv("JIRA_EMAIL", "user@example.com")
	os.Setenv("ATLASSIAN_API_TOKEN", "atlassian-token")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.APIToken != "atlassian-token" {
		t.Errorf("expected APIToken 'atlassian-token', got %q", cfg.APIToken)
	}
}

func TestLoad_JiraTokenTakesPrecedence(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_BASE_URL", "https://example.atlassian.net")
	os.Setenv("JIRA_EMAIL", "user@example.com")
	os.Setenv("JIRA_API_TOKEN", "jira-token")
	os.Setenv("ATLASSIAN_API_TOKEN", "atlassian-token")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.APIToken != "jira-token" {
		t.Errorf("expected APIToken 'jira-token', got %q", cfg.APIToken)
	}
}

func TestLoad_InvalidBaseURL_NotHTTPS(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_BASE_URL", "http://example.atlassian.net")
	os.Setenv("JIRA_EMAIL", "user@example.com")
	os.Setenv("JIRA_API_TOKEN", "test-token")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Errorf("expected error to mention https, got: %v", err)
	}
}

func TestLoad_InvalidBaseURL_MissingHost(t *testing.T) {
	clearEnv()
	os.Setenv("JIRA_BASE_URL", "https://")
	os.Setenv("JIRA_EMAIL", "user@example.com")
	os.Setenv("JIRA_API_TOKEN", "test-token")
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "host") {
		t.Errorf("expected error to mention host, got: %v", err)
	}
}

func TestLoad_MultipleErrors(t *testing.T) {
	clearEnv()
	defer clearEnv()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "JIRA_BASE_URL") {
		t.Errorf("expected error to mention JIRA_BASE_URL, got: %v", err)
	}
	if !strings.Contains(errStr, "JIRA_EMAIL") {
		t.Errorf("expected error to mention JIRA_EMAIL, got: %v", err)
	}
	if !strings.Contains(errStr, "JIRA_API_TOKEN") {
		t.Errorf("expected error to mention JIRA_API_TOKEN, got: %v", err)
	}
}
