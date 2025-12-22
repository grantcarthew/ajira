# DR-003: API Client Design

- Date: 2025-12-22
- Status: Accepted
- Category: API

## Problem

ajira needs to communicate with the Jira REST API v3. The client must:

- Authenticate using email and API token
- Handle JSON request/response bodies
- Parse and report API errors clearly
- Respect timeout configuration

## Decision

Use Go standard library `net/http` with Basic Authentication.

Client structure:

- Single `Client` struct holding HTTP client and configuration
- Generic request method for all HTTP operations
- Typed response parsing at call sites

Authentication:

- HTTP Basic Auth header
- Format: `Authorization: Basic base64(email:token)`

Error handling:

- Structured `APIError` type containing:
  - HTTP status code
  - Jira error messages (from response body)
  - Request context (method, path)

Request/Response:

- JSON encoding via standard library `encoding/json`
- Content-Type: application/json for all requests
- Accept: application/json for all responses

## Why

- Standard library has no external dependencies
- Basic Auth is Jira's recommended method for API tokens
- Structured errors allow callers to inspect status codes and messages
- Generic request method reduces code duplication across endpoints

## Structure

APIError fields:

| Field | Type | Description |
| ----- | ---- | ----------- |
| StatusCode | int | HTTP response status code |
| Status | string | HTTP status text (e.g., "404 Not Found") |
| Messages | []string | Error messages from Jira response body |
| Method | string | HTTP method of failed request |
| Path | string | Request path of failed request |

Jira error response format:

```json
{
  "errorMessages": ["Issue does not exist"],
  "errors": {}
}
```

## API Endpoints

Base URL: `{JIRA_BASE_URL}/rest/api/3`

GET /myself - Current user:

```json
{
  "accountId": "5b10a2844c20165700ede21g",
  "displayName": "Mia Krystof",
  "emailAddress": "mia@example.com",
  "timeZone": "Australia/Sydney",
  "active": true
}
```

GET /project/search - List projects (paginated):

Query parameters:
- `startAt` (int): Page offset, default 0
- `maxResults` (int): Page size, default 50
- `query` (string): Filter by name/key

Response:

```json
{
  "values": [
    {
      "id": "10000",
      "key": "EX",
      "name": "Example",
      "lead": { "displayName": "Jane Doe" },
      "style": "classic"
    }
  ],
  "startAt": 0,
  "maxResults": 50,
  "total": 1,
  "isLast": true
}
```

Note: Using v3 API for future ADF support (P-004). The `/project/search` endpoint is Atlassian's recommended replacement for the deprecated `/project` endpoint.

## Trade-offs

Accept:

- Must handle HTTP details manually (no convenience wrapper)
- Error parsing requires understanding Jira's error format

Gain:

- Zero external dependencies
- Full control over request/response handling
- Clear error messages for debugging

## Alternatives

Third-party HTTP client (resty, go-retryablehttp):

- Pro: Convenience methods, automatic retries
- Con: External dependency
- Con: Less control over behaviour
- Rejected: Standard library is sufficient for our needs

Jira Go SDK:

- Pro: Pre-built methods for all endpoints
- Con: Large dependency
- Con: May not support all API features we need
- Con: Abstracts away details we need to control
- Rejected: Too heavy, less flexibility
