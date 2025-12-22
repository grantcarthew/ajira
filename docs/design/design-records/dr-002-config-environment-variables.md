# DR-002: Environment Variable Configuration

- Date: 2025-12-22
- Status: Accepted
- Category: Configuration

## Problem

ajira requires configuration for Jira API connectivity:

- Base URL of the Jira instance
- Authentication credentials
- Optional default project

The configuration approach must support:

- Non-interactive execution (no prompts)
- CI/CD and automation environments
- AI agent usage where interactive config is impossible
- Secure credential handling

## Decision

Use environment variables exclusively for configuration. No config files.

Environment variables:

| Variable | Required | Description |
| -------- | -------- | ----------- |
| JIRA_BASE_URL | Yes | Jira instance URL (e.g., https://company.atlassian.net) |
| JIRA_EMAIL | Yes | User email for authentication |
| JIRA_API_TOKEN | Yes* | API token for authentication |
| ATLASSIAN_API_TOKEN | No | Fallback if JIRA_API_TOKEN not set |
| JIRA_PROJECT | No | Default project key for commands |

*JIRA_API_TOKEN is required unless ATLASSIAN_API_TOKEN is set.

Default values:

| Setting | Value |
| ------- | ----- |
| HTTP timeout | 30 seconds |

## Why

- Environment variables are the standard for containerised and CI/CD environments
- No file management complexity (paths, permissions, formats)
- Secrets remain in environment, not on disk
- Compatible with all secret management tools (Vault, AWS Secrets Manager, etc.)
- AI agents can set environment variables programmatically
- ATLASSIAN_API_TOKEN fallback allows shared token across Atlassian tools

## Validation

At application startup:

1. Load all environment variables
2. Check required variables are present and non-empty
3. Validate JIRA_BASE_URL is a valid URL with https scheme
4. If validation fails, exit with code 1 and clear error message

Error message format:

```
ajira: missing required environment variable: JIRA_BASE_URL
```

```
ajira: invalid JIRA_BASE_URL: must use https scheme
```

## Execution Flow

Token resolution order:

1. Check JIRA_API_TOKEN
2. If empty, check ATLASSIAN_API_TOKEN
3. If both empty, fail with error

## Trade-offs

Accept:

- No persistent configuration (must set env vars each session or in shell profile)
- Cannot have multiple named profiles
- All configuration is global (no per-directory overrides)

Gain:

- Zero file I/O for configuration
- Works identically in containers, CI, and local dev
- No config file format to learn or parse
- Secure by default (no credentials on disk)

## Alternatives

Config file (TOML/YAML):

- Pro: Persistent, can have multiple profiles
- Pro: Easier for humans to edit
- Con: File permissions and path complexity
- Con: Secrets stored on disk
- Con: Requires parsing library
- Rejected: Conflicts with non-interactive, automation-first design

Config file with env var override:

- Pro: Best of both worlds
- Con: Complexity of merge logic
- Con: Confusion about which value is active
- Rejected: Unnecessary complexity for target use case

Keychain/credential store integration:

- Pro: Most secure credential storage
- Con: Platform-specific implementation
- Con: Complex for CI/CD environments
- Con: AI agents cannot easily access
- Rejected: Does not support primary use case (automation)
