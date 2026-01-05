# ajira

Atlassian Jira CLI designed for AI agents and automation. Non-interactive, environment-configured, with Markdown input/output and JSON support.

See <https://agents.md/> for the full AGENTS.md specification as this project matures.

## Status

Under active development.

## Active Project

Projects are stored in the docs/projects/ directory. Update this when starting a new project.

Active Project: P-006 Integration Testing

Completed: P-005 Comment Functionality (2026-01-05)

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
