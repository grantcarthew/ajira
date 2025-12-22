# DR-005: Me Command

- Date: 2025-12-22
- Status: Accepted
- Category: CLI

## Problem

Users need to verify their Jira connection and credentials are working correctly. The `me` command serves as:

- Connection validation (can we reach the API?)
- Authentication validation (are credentials correct?)
- User identification (who am I authenticated as?)

## Decision

Implement `ajira me` command that calls GET /rest/api/3/myself and displays current user information.

Command structure:

```
ajira me [flags]
```

No subcommands or additional flags beyond global flags (`--json`, `--project`).

## Output Formats

Plain text (default):

```
Display Name: Mia Krystof
Email: mia@example.com
Account ID: 5b10a2844c20165700ede21g
Timezone: Australia/Sydney
Active: true
```

JSON (`ajira me -j`):

```json
{
  "accountId": "5b10a2844c20165700ede21g",
  "displayName": "Mia Krystof",
  "emailAddress": "mia@example.com",
  "timeZone": "Australia/Sydney",
  "active": true
}
```

## Error Handling

| Scenario | Exit Code | Message |
| -------- | --------- | ------- |
| Missing config | 1 | ajira: missing required environment variable: JIRA_BASE_URL |
| Invalid credentials | 1 | ajira: authentication failed (401) |
| Network error | 1 | ajira: failed to connect to Jira API |
| API error | 1 | ajira: API error - {message from Jira} |

## Why

- Simplest possible command to validate end-to-end functionality
- No arguments required - just authentication
- Proves config, API client, and CLI framework work together
- Standard pattern in CLIs (cf. `gh auth status`, `gcloud auth list`)

## Trade-offs

Accept:

- Limited utility beyond validation
- Returns more fields than strictly necessary

Gain:

- Simple first command to implement
- Validates entire stack
- Useful for debugging connection issues
