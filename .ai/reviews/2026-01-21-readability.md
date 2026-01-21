# Readability Code Review

**Date:** 2026-01-21
**Reviewer:** AI (Claude)
**Scope:** All Go source files in cmd/ and internal/

## Summary

The ajira codebase demonstrates strong readability overall. The code follows Go conventions, uses clear naming, and maintains consistent patterns throughout. A new team member would be able to understand and contribute to this codebase without significant difficulty.

**Overall Assessment:** Good readability with minor improvement opportunities.

---

## 1. Formatting

### Findings

**Positive:**
- Code is consistently formatted with `gofmt`
- Import grouping follows convention: stdlib, external, internal
- Vertical spacing appropriately groups related code
- Line lengths are reasonable throughout

**No issues identified.**

---

## 2. Naming

### Findings

**Positive:**
- Variable names are descriptive: `issueKey`, `targetStatus`, `transitions`
- Short names used for short scopes: `i` in loops, `r` for rune, `n` for node
- Acronyms consistently cased: `ID`, `URL`, `ADF`, `API`
- Functions describe what they do: `getTransitions`, `buildJQL`, `convertNode`
- Package names are short, lowercase: `cli`, `api`, `jira`, `config`, `width`
- No package stuttering (good: `api.Client`, not `api.APIClient`)

**Minor Observations:**

| Location | Observation | Suggestion |
|----------|-------------|------------|
| `internal/cli/issue_create.go:65-75` | Module-level vars `createSummary`, `createBody`, etc. | Consider grouping in a struct for testability |
| `internal/cli/issue_list.go:72-86` | Similar pattern with `issueList*` vars | Same suggestion |
| `internal/cli/issue_edit.go:24-40` | Pattern repeats with `edit*` vars | Same suggestion |

These are not readability issues per se - the naming is clear. The observation is that the pattern could benefit from a struct-based approach for consistency and testability.

---

## 3. Function Design

### Findings

**Positive:**
- Functions are generally short and focused
- Early returns used to reduce nesting (e.g., `internal/cli/exitcodes.go:44-79`)
- Happy path is the main flow with errors handled early
- Parameter lists are reasonable in length

**Observations:**

| Location | Observation | Impact |
|----------|-------------|--------|
| `internal/cli/issue_edit.go:91-277` | `runIssueEdit` is 186 lines | Moderate - still readable but could be split |
| `internal/cli/issue_list.go:120-230` | `runIssueList` is 110 lines | Low - acceptable for command handler |
| `internal/converter/markdown_to_adf.go:243-297` | `convertTaskItem` has deep nesting | Low - complexity is inherent to AST traversal |

**Example of good early return pattern:**

```go
// internal/cli/exitcodes.go:44-52
func ExitCodeFromError(err error) int {
    if err == nil {
        return ExitSuccess
    }
    // ... rest of function
}
```

---

## 4. Code Flow

### Findings

**Positive:**
- Code is organised top-to-bottom in reading order
- Helper functions placed after the functions that use them
- Related functions grouped together (e.g., all transition-related functions in `issue_move.go`)
- Exported types at the top of files
- `init()` functions consistently placed after type declarations and before main logic

**Structure follows consistent pattern across files:**
1. Package declaration
2. Imports
3. Types and constants
4. Package-level variables
5. Cobra command definition
6. `init()` function
7. Main run function
8. Helper functions

This pattern is consistently applied and easy to follow.

---

## 5. Comments

### Findings

**Positive:**
- Doc comments present on exported types and functions
- Comments explain "why" not "what" (e.g., `// Jira ADF rejects empty text nodes in code blocks - use space placeholder`)
- No dead commented code observed
- ADF node type constants well-documented

**Good examples:**

```go
// internal/converter/markdown_to_adf.go:119-120
// Jira ADF rejects empty text nodes in code blocks - use space placeholder
if strings.TrimSpace(text) == "" {
    text = " "
}
```

```go
// internal/api/client.go:192-193
// getRetryAfter extracts the retry delay from a 429 response.
// Uses Retry-After header if present, otherwise exponential backoff.
```

**Observations:**

| Location | Observation | Suggestion |
|----------|-------------|------------|
| `internal/cli/issue_list.go:300-312` | `buildOrderBy` lacks doc comment | Add doc comment for completeness |
| `internal/width/width.go:129-215` | `wideRanges` could use summary comment | Consider brief explanation of coverage |

---

## 6. Complexity

### Findings

**Positive:**
- Complex expressions broken into named intermediates
- Cyclomatic complexity reasonable across most functions
- Deeply nested structures avoided

**Observations:**

| Location | Observation | Impact |
|----------|-------------|--------|
| `internal/cli/issue_edit.go:99-103` | Complex boolean expression for `hasChanges` | Low - clear though long |
| `internal/converter/markdown_to_adf.go:186-212` | Multiple nested type checks for task list detection | Low - inherent to goldmark AST |

**Example of good intermediate value usage:**

```go
// internal/cli/issue_list.go:182-192
keyWidth, statusWidth, typeWidth, assigneeWidth := 8, 11, 4, 8
for _, issue := range issues {
    if w := width.StringWidth(issue.Key); w > keyWidth {
        keyWidth = w
    }
    // ...
}
```

---

## 7. Consistency

### Findings

**Positive:**
- Similar things done in similar ways throughout
- Error handling follows consistent pattern: wrap with context, return early
- JSON output pattern consistent: check `JSONOutput()`, marshal with indent
- API error handling pattern consistent across all commands
- Command structure follows same template

**Consistent error wrapping pattern:**

```go
if err != nil {
    return fmt.Errorf("failed to <action>: %w", err)
}
```

**Consistent API error handling:**

```go
if apiErr, ok := err.(*api.APIError); ok {
    return fmt.Errorf("API error: %w", apiErr)
}
```

**No inconsistencies identified.**

---

## 8. Magic Values

### Findings

**Positive:**
- Constants defined for exit codes (`ExitSuccess`, `ExitUserError`, etc.)
- Constants defined for ADF node and mark types
- Timeout constants defined (`DefaultTimeout`, `maxRetries`, `initialBackoff`)

**Observations:**

| Location | Value | Suggestion |
|----------|-------|------------|
| `internal/cli/issue_list.go:316` | `maxResults := 50` | Could be a package constant |
| `internal/cli/issue_list.go:322` | `maxPages = 100` | Good - already a constant |
| `internal/cli/issue_view.go:253` | `60` in `strings.Repeat("-", 60)` | Could be a named constant |
| `internal/cli/issue_list.go:223` | `60` in `width.Truncate(issue.Summary, 60, "...")` | Could be named for clarity |

These are minor and do not significantly impact readability.

---

## 9. Dead Code

### Findings

**No dead code identified.**

- No unreachable code paths observed
- No unused exported functions or types
- No obsolete comments

---

## Patterns That Work Well

1. **Consistent Command Structure** - Every command file follows the same structure making it easy to navigate
2. **Clear Type Definitions** - JSON response types mirror API structure exactly
3. **Validation Before Action** - Input validated before making API calls
4. **Error Context** - Errors wrapped with action context at each level
5. **Unicode Width Handling** - Well-documented width calculation with comprehensive ranges

---

## Recommendations

### High Priority

None - the codebase has good readability.

### Medium Priority

1. **Consider struct-based flag handling** - The pattern of module-level flag variables (`createSummary`, `editSummary`, etc.) could be replaced with a struct per command for improved testability and reduced global state.

### Low Priority

1. **Extract magic numbers** - Values like `60` for display widths and `50` for page sizes could be named constants for clarity.

2. **Add doc comments** - A few helper functions lack doc comments (`buildOrderBy`, `colorStatus`).

---

## Conclusion

The ajira codebase demonstrates mature Go code practices. The code reads like well-written prose, with clear naming, consistent patterns, and appropriate comments. A developer new to the codebase would be able to understand the code flow and contribute effectively.

The recommendations are minor improvements rather than essential fixes. The current readability is suitable for production code.
