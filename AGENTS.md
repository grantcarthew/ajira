# ajira

Atlassian Jira Cloud CLI designed for AI agents and automation. Non-interactive, environment-configured, with Markdown input/output and JSON support. Targets Jira Cloud only — Jira Server and Data Center are not supported.

Module path: `github.com/grantcarthew/ajira`.

## Setup

```bash
git clone https://github.com/grantcarthew/ajira.git
cd ajira
go build -o ajira ./cmd/ajira
```

Configure via environment variables. `ATLASSIAN_*` variables are shared with the sibling `acon` (Confluence) tool; `JIRA_*` variables override them when set.

| Variable | Purpose |
| --- | --- |
| `ATLASSIAN_BASE_URL` | Instance URL, e.g. `https://example.atlassian.net` |
| `ATLASSIAN_EMAIL` | Account email |
| `ATLASSIAN_API_TOKEN` | API token |
| `JIRA_BASE_URL` | Overrides `ATLASSIAN_BASE_URL` |
| `JIRA_EMAIL` | Overrides `ATLASSIAN_EMAIL` |
| `JIRA_API_TOKEN` | Overrides `ATLASSIAN_API_TOKEN` |
| `JIRA_PROJECT` | Default project key (optional) |
| `JIRA_BOARD` | Default board ID (optional) |

Verify: `./ajira me`

## Build and Test

```bash
go build -o ajira ./cmd/ajira
go test ./...
go test -v ./...
go test ./internal/cli/...
go test ./internal/converter/...
```

Quick smoke check against a live instance:

```bash
./ajira me
./ajira project list
./ajira issue list -l 3
```

Integration scripts live in `testdata/` (`clone-test.sh`, `roundtrip-test.sh`); see `testdata/README.md` before running — they hit a real Jira instance.

## Code Style

- Go standard formatting (`gofmt`); run before committing
- Pure Go only, no cgo
- Return errors, do not panic
- Cobra for CLI commands (`spf13/cobra`)
- Non-interactive: input via flags, arguments, or stdin — never prompt
- Environment variables for configuration; no config files
- Markdown in, Markdown out; conversion through `internal/converter` (ADF ↔ Markdown)

## Project Structure

```
cmd/ajira/         Main entry point
internal/api/      Jira REST client and transport
internal/cli/      Cobra command implementations
internal/config/   Environment and configuration loading
internal/converter/Markdown ↔ ADF conversion
internal/jira/     Jira domain helpers (metadata, validation)
internal/width/    Terminal width and formatting helpers
testdata/          Integration scripts and fixtures
.ai/               Agent role definitions and working notes
```

## Global CLI Flags

Most commands accept these flags; prefer them in scripts and automation.

| Flag | Purpose |
| --- | --- |
| `--json` | Machine-parseable JSON output |
| `--dry-run` | Preview without executing |
| `--quiet` | Suppress non-essential output |
| `--no-color` | Disable coloured output |
| `--verbose` | Show HTTP request/response details |
| `-p, --project` | Override default project |
| `--board` | Override default board ID |

## Commit and PR Conventions

- Conventional Commits with optional scope: `type(scope): subject`
- Types in use: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`
- Scopes seen: `cli`, `api`, `projects`, `issues`, `error`, `changelog`, `ai`
- Subject in lowercase imperative, no trailing period
- Do not add `Co-Authored-By` trailers
- Keep commits focused; one logical change per commit

## Agent Guidance

- Run `./ajira help {agents,schemas,markdown,agile}` for token-efficient CLI references tailored to automation
- Role definitions for this repo live in `.ai/roles/` (default role: Go expert)
- When scripting, always pass `--json` and parse with `jq`; fall back to text only for human-readable summaries
- Exit codes are stable and documented — see `internal/cli/exitcodes.go`
