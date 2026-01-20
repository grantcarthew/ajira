# Documentation Review

Date: 2026-01-20

## Summary

Overall, ajira has comprehensive documentation that is well-organised and generally accurate. The documentation covers installation, configuration, CLI usage, and includes AI-agent specific reference material. However, several discrepancies and gaps were identified between the documentation and actual implementation.

## Issues by Type

### 1. Documentation Inaccuracies

#### 1.1 issue link remove - Incorrect Arguments (HIGH)

**Location:** `docs/flags-and-arguments.md` lines 256-266

**Issue:** Documentation states the command requires 3 arguments including link-type:
```
ajira issue link remove <issue-key> <link-type> <linked-key>
```

**Actual Behaviour:** The command takes only 2 arguments and removes ALL links between two issues:
```go
// internal/cli/issue_link_remove.go:54
Use:   "remove <key1> <key2>",
Args:  cobra.ExactArgs(2),
```

**Recommendation:** Update `docs/flags-and-arguments.md` to match the actual implementation:
```
ajira issue link remove <key1> <key2>
```
And update the description to reflect that ALL links between the two issues are removed.

### 2. Missing Documentation

#### 2.1 README Commands Table Incomplete (MEDIUM)

**Location:** `README.md` lines 289-312

**Issue:** The Commands table in README.md omits many implemented commands:
- `issue clone` - exists and works
- `issue watch` / `issue unwatch` - exists and works
- `board list` - exists and works
- `sprint list` / `sprint add` - exists and works
- `epic list` / `epic create` / `epic add` / `epic remove` - exists and works
- `release list` - exists and works
- `user search` - exists and works
- `field list` - exists and works
- `open` - exists and works

**Recommendation:** Add all implemented commands to the Commands table.

#### 2.2 README Roadmap is Outdated (LOW)

**Location:** `README.md` lines 332-338

**Issue:** The Roadmap lists features that have been implemented:
- "Issue Clone" - implemented in `internal/cli/issue_clone.go`
- "Agile Features" - implemented (epic and sprint commands)
- "Automation Support" - implemented (dry-run, batch, exit codes)

**Recommendation:** Update or remove the Roadmap section to reflect current status.

### 3. Missing Package Documentation

#### 3.1 No doc.go Files (LOW)

**Issue:** None of the ajira packages have `doc.go` files or package-level documentation comments. The only package with any package comment is `converter`:

```go
// Package converter provides bidirectional conversion between Markdown and
// Atlassian Document Format (ADF).
package converter
```

Missing package documentation in:
- `internal/api` - no package comment
- `internal/cli` - no package comment
- `internal/config` - no package comment
- `internal/jira` - no package comment
- `internal/width` - no package comment

**Recommendation:** Add package-level comments to each package describing its purpose.

### 4. Inconsistencies Between Documentation Sources

#### 4.1 Agent Reference vs flags-and-arguments.md (LOW)

**Location:** `internal/cli/help/agents.md` and `docs/flags-and-arguments.md`

**Issue:** The agent reference in `help/agents.md` shows:
```
ajira issue link remove PROJ-123 PROJ-456
```
Which is correct, but `docs/flags-and-arguments.md` shows incorrect 3-argument syntax.

**Recommendation:** Ensure all documentation sources are consistent.

### 5. Environment Variable Documentation

#### 5.1 Environment Variables Documented Well (OK)

The environment variables are documented in multiple places:
- README.md Configuration section - accurate
- `internal/cli/root.go` Long description - accurate
- Supports fallback to ATLASSIAN_* variants - documented

One minor note: The `--board` flag and `JIRA_BOARD` environment variable are documented but could be more prominent since they're required for sprint commands.

### 6. Godoc Comments Quality

#### 6.1 Exported Symbols Generally Well Documented (OK)

Most exported types and functions have appropriate doc comments:
- `internal/api/client.go` - Client, APIError, methods documented
- `internal/jira/metadata.go` - Types and functions documented
- `internal/cli/exitcodes.go` - Exit codes and types documented

Areas for improvement:
- `internal/config/config.go`: `Config` struct and `Load()` function lack doc comments
- Some internal types could benefit from more context

### 7. CLI Help Output Accuracy

#### 7.1 Help Output Generally Accurate (OK)

The embedded help files (`agents.md`, `schemas.md`, `markdown.md`) are accurate and provide useful reference material for AI agents.

The Cobra command help strings are accurate and include useful examples.

## Documentation Quality Checklist

| Criterion | Status | Notes |
|-----------|--------|-------|
| Accurate | Mostly | One significant inaccuracy (issue link remove) |
| Complete | Partial | Commands table missing many commands |
| Clear | Good | Well-written and understandable |
| Current | Partial | Roadmap outdated |
| Consistent | Partial | Discrepancy between docs sources |
| Accessible | Good | Well-organised structure |

## Priority Recommendations

### High Priority
1. Fix `docs/flags-and-arguments.md` issue link remove documentation

### Medium Priority
2. Add missing commands to README.md Commands table

### Low Priority
3. Update or remove outdated Roadmap section
4. Add package-level documentation comments
5. Add doc comment to `Config` struct and `Load()` function

## Files Reviewed

- `README.md`
- `AGENTS.md`
- `docs/README.md`
- `docs/cli/automation.md`
- `docs/flags-and-arguments.md`
- `docs/thoughts.md`
- `docs/projects/README.md`
- `internal/cli/help/agents.md`
- `internal/cli/help/schemas.md`
- `internal/cli/help/markdown.md`
- `internal/cli/root.go`
- `internal/cli/help.go`
- `internal/cli/exitcodes.go`
- `internal/cli/issue_*.go` (various)
- `internal/api/client.go`
- `internal/jira/metadata.go`
- `internal/converter/adf.go`
- `internal/config/config.go`
- `cmd/ajira/main.go`
