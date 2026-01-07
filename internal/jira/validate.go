package jira

import (
	"fmt"
	"strings"

	"github.com/gcarthew/ajira/internal/api"
)

// ValidatePriority checks if the given priority name is valid.
// Returns nil if valid, or an error with valid options listed.
func ValidatePriority(client *api.Client, priority string) error {
	if priority == "" {
		return nil
	}

	priorities, err := GetPriorities(client)
	if err != nil {
		return fmt.Errorf("failed to fetch priorities: %w", err)
	}

	var names []string
	for _, p := range priorities {
		names = append(names, p.Name)
		if strings.EqualFold(p.Name, priority) {
			return nil
		}
	}

	return fmt.Errorf("invalid priority %q, valid options: %s", priority, strings.Join(names, ", "))
}

// ValidateIssueType checks if the given issue type name is valid for the project.
// Returns nil if valid, or an error with valid options listed.
func ValidateIssueType(client *api.Client, projectKey, issueType string) error {
	if issueType == "" {
		return nil
	}

	types, err := GetIssueTypes(client, projectKey)
	if err != nil {
		return fmt.Errorf("failed to fetch issue types: %w", err)
	}

	var names []string
	for _, t := range types {
		names = append(names, t.Name)
		if strings.EqualFold(t.Name, issueType) {
			return nil
		}
	}

	return fmt.Errorf("invalid issue type %q, valid options: %s", issueType, strings.Join(names, ", "))
}

// ValidateStatus checks if the given status name is valid for the project.
// Returns nil if valid, or an error with valid options listed.
func ValidateStatus(client *api.Client, projectKey, status string) error {
	if status == "" {
		return nil
	}

	statuses, err := GetStatuses(client, projectKey)
	if err != nil {
		return fmt.Errorf("failed to fetch statuses: %w", err)
	}

	var names []string
	for _, s := range statuses {
		names = append(names, s.Name)
		if strings.EqualFold(s.Name, status) {
			return nil
		}
	}

	return fmt.Errorf("invalid status %q, valid options: %s", status, strings.Join(names, ", "))
}
