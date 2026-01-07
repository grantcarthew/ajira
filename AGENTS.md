# ajira

Atlassian Jira CLI designed for AI agents and automation. Non-interactive, environment-configured, with Markdown input/output and JSON support.

See <https://agents.md/> for the full AGENTS.md specification as this project matures.

## Status

Under active development.

## Active Project

Projects are stored in the docs/projects/ directory. Update this when starting a new project.

Active Project: [P-007 Issue Linking](docs/projects/p-007-cli-issue-linking.md)

Completed: P-006 Integration Testing (2026-01-07), P-005 Comment Functionality (2026-01-05)

Proposed: P-008 through P-014 (see docs/projects/README.md)

## Quick Reference

```bash
# Build
go build -o ajira ./cmd/ajira

# Test
go test ./...

# Run
./ajira me
./ajira project list
```

---

## Documentation Driven Development (DDD)

This project uses Documentation Driven Development. Design decisions are documented in Design Records (DRs) before or during implementation.

For complete DR writing guidelines: See [docs/design/dr-writing-guide.md](docs/design/dr-writing-guide.md)

For project writing guidelines: See [docs/projects/p-writing-guide.md](docs/projects/p-writing-guide.md)

For feature development workflow: See [docs/workflow.md](docs/workflow.md)

Location: `docs/design/design-records/`
