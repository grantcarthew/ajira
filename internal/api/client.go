package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gcarthew/ajira/internal/config"
)

const basePathV3 = "/rest/api/3"

// Rate limit retry configuration
const (
	maxRetries     = 3
	initialBackoff = 1 * time.Second
)

// verboseWriter is the destination for verbose output (nil = disabled).
var verboseWriter io.Writer

// SetVerboseOutput enables verbose HTTP logging to the given writer.
func SetVerboseOutput(w io.Writer) {
	verboseWriter = w
}

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

// PostMultipart performs a multipart/form-data POST request for file uploads.
// Returns the response body and any error.
func (c *Client) PostMultipart(ctx context.Context, path string, contentType string, body []byte) ([]byte, error) {
	return c.doMultipartRequest(ctx, http.MethodPost, basePathV3+path, contentType, body)
}

// GetRaw performs a GET request and returns the raw response body and content type.
// Used for downloading binary content like attachments.
func (c *Client) GetRaw(ctx context.Context, path string) ([]byte, string, error) {
	url := c.baseURL + basePathV3 + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(c.email, c.token)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		if verboseWriter != nil {
			fmt.Fprintf(verboseWriter, "GET %s error (%s)\n", path, duration.Round(time.Millisecond))
		}
		return nil, "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if verboseWriter != nil {
		fmt.Fprintf(verboseWriter, "GET %s %s (%s)\n", path, resp.Status, duration.Round(time.Millisecond))
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Method:     http.MethodGet,
			Path:       path,
		}

		var jiraErr jiraErrorResponse
		if json.Unmarshal(respBody, &jiraErr) == nil {
			apiErr.Messages = jiraErr.ErrorMessages
			apiErr.Errors = jiraErr.Errors
		} else if len(respBody) > 0 {
			apiErr.RawBody = string(respBody)
		}

		return nil, "", apiErr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	return body, contentType, nil
}

func (c *Client) doMultipartRequest(ctx context.Context, method, path, contentType string, body []byte) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.SetBasicAuth(c.email, c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Atlassian-Token", "no-check") // CSRF protection

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		if verboseWriter != nil {
			fmt.Fprintf(verboseWriter, "%s %s error (%s)\n", method, path, duration.Round(time.Millisecond))
		}
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if verboseWriter != nil {
		fmt.Fprintf(verboseWriter, "%s %s %s (%s)\n", method, path, resp.Status, duration.Round(time.Millisecond))
	}

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

		// Map specific HTTP status codes to user-friendly messages
		if resp.StatusCode == 413 {
			apiErr.Messages = []string{"file exceeds size limit"}
			return nil, apiErr
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

func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	return c.doRequestWithRetry(ctx, method, path, body, 0)
}

func (c *Client) doRequestWithRetry(ctx context.Context, method, path string, body []byte, attempt int) ([]byte, error) {
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

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	duration := time.Since(start)

	if err != nil {
		if verboseWriter != nil {
			fmt.Fprintf(verboseWriter, "%s %s error (%s)\n", method, path, duration.Round(time.Millisecond))
		}
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if verboseWriter != nil {
		fmt.Fprintf(verboseWriter, "%s %s %s (%s)\n", method, path, resp.Status, duration.Round(time.Millisecond))
	}

	// Handle rate limiting with retry
	if resp.StatusCode == 429 && attempt < maxRetries {
		retryAfter := getRetryAfter(resp, attempt)
		if verboseWriter != nil {
			fmt.Fprintf(verboseWriter, "Rate limited, retrying in %s (attempt %d/%d)\n", retryAfter, attempt+1, maxRetries)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryAfter):
		}

		return c.doRequestWithRetry(ctx, method, path, body, attempt+1)
	}

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

// getRetryAfter extracts the retry delay from a 429 response.
// Uses Retry-After header if present, otherwise exponential backoff.
func getRetryAfter(resp *http.Response, attempt int) time.Duration {
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		// Try parsing as seconds
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(seconds) * time.Second
		}
		// Try parsing as HTTP date
		if t, err := http.ParseTime(retryAfter); err == nil {
			return time.Until(t)
		}
	}
	// Exponential backoff: 1s, 2s, 4s for attempts 0, 1, 2
	return initialBackoff * (1 << attempt)
}
