package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/grantcarthew/ajira/internal/api"
)

type customerAutocompleteRequest struct {
	Query         string         `json:"query"`
	IssueID       string         `json:"issueId,omitempty"`
	FieldValueMap map[string]any `json:"fieldValueMap"`
}

type customerAutocompleteResult struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	ObjectKey string `json:"objectKey"`
}

type customerAutocompleteResponse struct {
	Results     []customerAutocompleteResult `json:"results"`
	HasNextPage bool                         `json:"hasNextPage"`
}

// resolveCustomer looks up a customer name via the Assets autocomplete API.
// issueID is the numeric id of the issue being created or edited.
// An empty query returns an empty id and no error.
func resolveCustomer(ctx context.Context, client *api.Client, query, issueID string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", nil
	}

	reqBody := customerAutocompleteRequest{
		Query:         query,
		IssueID:       issueID,
		FieldValueMap: map[string]any{},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal customer lookup: %w", err)
	}

	path := fmt.Sprintf("/field/%s/config/%s/autocomplete?start=0&limit=50&excludeCurrentValues=true", customerFieldKey, customerFieldConfigID)
	respBody, err := client.CmdbPost(ctx, path, body)
	if err != nil {
		return "", fmt.Errorf("failed to look up customer: %w\nrequest: %s", err, body)
	}

	var resp customerAutocompleteResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("failed to parse customer lookup: %w", err)
	}

	return pickCustomer(query, resp.Results, resp.HasNextPage)
}

// customerFieldForEdit returns the customfield_10385 value for an edit.
// An empty query clears the field. A name or object key is resolved to an object id.
func customerFieldForEdit(ctx context.Context, client *api.Client, issueKey, query string) (any, error) {
	if strings.TrimSpace(query) == "" {
		return []customerRef{}, nil
	}

	issueID, err := fetchIssueID(ctx, client, issueKey)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve customer: %w", err)
	}
	customerID, err := resolveCustomer(ctx, client, query, issueID)
	if err != nil {
		return nil, err
	}
	return []customerRef{{ID: customerID}}, nil
}

// setIssueCustomer resolves query against issueID and writes customfield_10385.
func setIssueCustomer(ctx context.Context, client *api.Client, issueKey, issueID, query string) error {
	customerID, err := resolveCustomer(ctx, client, query, issueID)
	if err != nil {
		return err
	}
	if customerID == "" {
		return nil
	}
	return updateIssue(ctx, client, issueKey, map[string]any{
		customerFieldKey: []customerRef{{ID: customerID}},
	}, nil)
}

func pickCustomer(query string, results []customerAutocompleteResult, hasNextPage bool) (string, error) {
	for _, result := range results {
		if strings.EqualFold(result.Label, query) {
			return result.ID, nil
		}
	}
	for _, result := range results {
		if strings.EqualFold(result.ObjectKey, query) {
			return result.ID, nil
		}
	}
	if len(results) == 1 && !hasNextPage {
		return results[0].ID, nil
	}
	if len(results) == 0 {
		return "", fmt.Errorf("no customer matching %q", query)
	}

	labels := make([]string, 0, len(results))
	for _, result := range results {
		labels = append(labels, result.Label)
	}
	return "", fmt.Errorf("ambiguous customer %q (matches: %s)", query, strings.Join(labels, ", "))
}

func fetchIssueID(ctx context.Context, client *api.Client, key string) (string, error) {
	body, err := client.Get(ctx, fmt.Sprintf("/issue/%s?fields=summary", key))
	if err != nil {
		return "", err
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse issue %s: %w", key, err)
	}
	if resp.ID == "" {
		return "", fmt.Errorf("issue %s has no id", key)
	}
	return resp.ID, nil
}
