package api

import (
	"context"
	"net/http"
)

const basePathCMDB = "/rest/servicedesk/cmdb/1"

// CmdbPost performs a POST request to the Jira Service Management CMDB API.
func (c *Client) CmdbPost(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, basePathCMDB+path, body)
}
