# DR-001: Project Structure

- Date: 2025-12-22
- Status: Accepted
- Category: Architecture

## Problem

ajira needs a foundational project structure that:

- Follows Go community conventions for discoverability
- Separates concerns cleanly (CLI, API client, configuration)
- Prevents external import of internal packages
- Supports future growth without restructuring

## Decision

Use standard Go project layout with the following structure:

```
ajira/
├── cmd/ajira/main.go       # Application entry point
├── internal/
│   ├── cli/                # Cobra commands and CLI logic
│   ├── api/                # Jira REST API client
│   ├── config/             # Environment configuration
│   ├── converter/          # Markdown ↔ ADF conversion
│   ├── jira/               # Jira field metadata and validation
│   └── width/              # Unicode display width calculation
├── docs/                   # Documentation and design records
├── go.mod
└── go.sum
```

Module path: `github.com/gcarthew/ajira`

## Why

- cmd/ pattern is standard for Go applications with multiple binaries potential
- internal/ prevents external packages from importing implementation details
- Separation into cli/, api/, config/ provides clear boundaries:
  - cli/ depends on api/ and config/
  - api/ depends on config/
  - config/ has no internal dependencies
- Module path uses GitHub convention for potential open-source hosting

## Trade-offs

Accept:

- Deeper directory nesting than a flat structure
- Must use full import paths within the project

Gain:

- Enforced encapsulation via internal/
- Clear dependency direction
- Familiar structure for Go developers
- Room for growth (docs/, scripts/, testdata/, etc.)

## Alternatives

Flat structure (all .go files in root):

- Pro: Simple, fewer directories
- Con: No encapsulation, all packages importable
- Con: Becomes unwieldy as project grows
- Rejected: Does not scale, no protection for internal code

pkg/ instead of internal/:

- Pro: Allows external import of packages
- Con: ajira is an application, not a library
- Rejected: No benefit to exposing internals

Single package for everything:

- Pro: No import paths to manage
- Con: No separation of concerns
- Con: Testing becomes difficult
- Rejected: Violates Go best practices for non-trivial projects
