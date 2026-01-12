package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gcarthew/ajira/internal/config"
)

const basePathV3 = "/rest/api/3"

// Client is a Jira REST API client.
type Client struct {
	baseURL    string
	email      string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Jira API client from config.
func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		email:   cfg.Email,
		token:   cfg.APIToken,
		httpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

// APIError represents an error response from the Jira API.
type APIError struct {
	StatusCode int
	Status     string
	Messages   []string
	Errors     map[string]string
	RawBody    string // Raw response body when JSON parsing fails
	Method     string
	Path       string
}

func (e *APIError) Error() string {
	parts := append([]string{}, e.Messages...)
	for field, msg := range e.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	if len(parts) == 0 {
		if e.RawBody != "" {
			return fmt.Sprintf("%s %s: %s - %s", e.Method, e.Path, e.Status, e.RawBody)
		}
		return fmt.Sprintf("%s %s: %s", e.Method, e.Path, e.Status)
	}
	return fmt.Sprintf("%s %s: %s - %s", e.Method, e.Path, e.Status, strings.Join(parts, "; "))
}

// jiraErrorResponse matches Jira's error response format.
type jiraErrorResponse struct {
	ErrorMessages   []string          `json:"errorMessages"`
	Errors          map[string]string `json:"errors"`
	WarningMessages []string          `json:"warningMessages"`
}

// Get performs a GET request to the Jira v3 API.
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.request(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request to the Jira v3 API.
func (c *Client) Post(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.request(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request to the Jira v3 API.
func (c *Client) Put(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.request(ctx, http.MethodPut, path, body)
}

// Delete performs a DELETE request to the Jira v3 API.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.request(ctx, http.MethodDelete, path, nil)
}

func (c *Client) request(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	return c.doRequest(ctx, method, basePathV3+path, body)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(c.email, c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Method:     method,
			Path:       path,
		}

		var jiraErr jiraErrorResponse
		if json.Unmarshal(respBody, &jiraErr) == nil {
			apiErr.Messages = jiraErr.ErrorMessages
			apiErr.Errors = jiraErr.Errors
		} else if len(respBody) > 0 {
			apiErr.RawBody = string(respBody)
		}

		return nil, apiErr
	}

	return respBody, nil
}
