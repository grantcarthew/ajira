package jira

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcarthew/ajira/internal/api"
)

// Priority represents a Jira priority.
type Priority struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// IssueType represents a Jira issue type.
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

// Status represents a Jira status.
type Status struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// LinkType represents a Jira issue link type.
type LinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

type priorityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type issueTypeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

type statusResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StatusCategory struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"statusCategory"`
}

type projectStatusesResponse struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Statuses []statusResponse `json:"statuses"`
}

// GetPriorities fetches all priorities from the Jira instance.
func GetPriorities(ctx context.Context, client *api.Client) ([]Priority, error) {
	body, err := client.Get(ctx, "/priority")
	if err != nil {
		return nil, err
	}

	var resp []priorityResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	priorities := make([]Priority, len(resp))
	for i, p := range resp {
		priorities[i] = Priority(p)
	}

	return priorities, nil
}

// GetIssueTypes fetches issue types for a project.
func GetIssueTypes(ctx context.Context, client *api.Client, projectKey string) ([]IssueType, error) {
	path := fmt.Sprintf("/issue/createmeta/%s/issuetypes", projectKey)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp struct {
		IssueTypes []issueTypeResponse `json:"issueTypes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		// Try parsing as direct array (different API versions)
		var types []issueTypeResponse
		if err2 := json.Unmarshal(body, &types); err2 != nil {
			return nil, fmt.Errorf("failed to parse response (object: %v, array: %v)", err, err2)
		}
		resp.IssueTypes = types
	}

	issueTypes := make([]IssueType, len(resp.IssueTypes))
	for i, t := range resp.IssueTypes {
		issueTypes[i] = IssueType(t)
	}

	return issueTypes, nil
}

// GetStatuses fetches statuses for a project.
func GetStatuses(ctx context.Context, client *api.Client, projectKey string) ([]Status, error) {
	path := fmt.Sprintf("/project/%s/statuses", projectKey)

	body, err := client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp []projectStatusesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Deduplicate statuses across issue types
	seen := make(map[string]bool)
	var statuses []Status
	for _, issueType := range resp {
		for _, s := range issueType.Statuses {
			if seen[s.ID] {
				continue
			}
			seen[s.ID] = true
			statuses = append(statuses, Status{
				ID:       s.ID,
				Name:     s.Name,
				Category: s.StatusCategory.Name,
			})
		}
	}

	return statuses, nil
}

// GetLinkTypes fetches all issue link types from the Jira instance.
func GetLinkTypes(ctx context.Context, client *api.Client) ([]LinkType, error) {
	body, err := client.Get(ctx, "/issueLinkType")
	if err != nil {
		return nil, err
	}

	var resp struct {
		IssueLinkTypes []LinkType `json:"issueLinkTypes"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp.IssueLinkTypes, nil
}
