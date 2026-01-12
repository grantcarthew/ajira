package api

import (
	"context"
	"net/http"
)

const basePathAgile = "/rest/agile/1.0"

// AgileGet performs a GET request to the Jira Agile API.
func (c *Client) AgileGet(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, basePathAgile+path, nil)
}

// AgilePost performs a POST request to the Jira Agile API.
func (c *Client) AgilePost(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, basePathAgile+path, body)
}
