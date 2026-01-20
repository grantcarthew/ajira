# Go Correctness Code Review

**Date:** 2026-01-20
**Reviewer:** Claude Opus 4.5
**Scope:** Full codebase correctness review

## Summary

The codebase is well-structured with generally sound logic. No critical correctness bugs were identified that would cause incorrect behaviour in normal use.

| Severity | Count | Description |
|----------|-------|-------------|
| High | 0 | No critical correctness bugs |
| Medium | 2 | Package-level flags, mark comparison order |
| Low | 4 | Format assumptions, heuristics, edge cases |

---

## Findings

### MEDIUM SEVERITY

#### 1. Package-Level Flag Variables Could Cause State Leakage

**Files:** `internal/cli/issue_create.go`, `issue_edit.go`, `issue_comment.go`, `issue_clone.go`, `issue_move.go`, `issue_list.go`, and others

**Issue:** Flag values are stored in package-level variables that are not reset between invocations:

```go
// issue_create.go:65-75
var (
    createSummary     string
    createBody        string
    createFile        string
    // ...
)
```

**Impact:** In the current CLI usage (one command per process), this works correctly. However:
- If the CLI were used as a library with multiple command invocations
- In tests that run multiple commands sequentially
- If commands are reused in the same process

Flag values from previous invocations could persist and affect subsequent commands.

**Assessment:** Low practical impact for current CLI design, but would become a bug if usage patterns change.

---

#### 2. `marksEqual` Comparison is Order-Dependent

**File:** `internal/converter/adf_to_markdown.go:315-325`

```go
func marksEqual(a, b []ADFMark) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i].Type != b[i].Type {
            return false
        }
    }
    return true
}
```

**Issue:** Two ADFMark slices with identical types in different orders are considered unequal. For example, `[strong, em]` ≠ `[em, strong]`.

**Impact:** Adjacent text nodes with the same marks in different orders won't be merged, potentially causing over-escaping in the Markdown output. This is a minor rendering issue, not a data loss issue.

---

### LOW SEVERITY

#### 3. `formatDateTime` Assumes Fixed ISO 8601 Format

**File:** `internal/cli/issue_view.go:301-308`

```go
func formatDateTime(iso string) string {
    if len(iso) >= 16 {
        return iso[:10] + " " + iso[11:16]
    }
    return iso
}
```

**Issue:** Hardcoded string slicing assumes Jira always returns dates in format `YYYY-MM-DDTHH:MM:SS...`. If Jira returns dates in a different format, the output could be garbled.

**Mitigation:** The fallback (`return iso`) handles shorter strings gracefully.

---

#### 4. `extractProjectKey` Returns Full Input on Missing Delimiter

**File:** `internal/cli/issue.go:10-15`

```go
func extractProjectKey(issueKey string) string {
    if idx := strings.Index(issueKey, "-"); idx > 0 {
        return issueKey[:idx]
    }
    return issueKey
}
```

**Issue:** Malformed issue keys without hyphens return the full input rather than reporting an error. Used in `issue_edit.go:128` for validation, this could pass an invalid project key to the validation function.

**Impact:** Low - the subsequent API call would fail with a clear error.

---

#### 5. `resolveUser` AccountID Heuristic

**File:** `internal/cli/issue_assign.go:167-168`

```go
if !strings.Contains(user, "@") && len(user) > minAccountIDLength {
    return user, nil
}
```

**Issue:** Assumes account IDs are >20 chars and don't contain `@`. Unusual account ID formats could be misclassified and incorrectly passed to the search API.

**Impact:** Low - would result in a "user not found" error.

---

#### 6. Retry-After with Past Date

**File:** `internal/api/client.go:200-202`

```go
if t, err := http.ParseTime(retryAfter); err == nil {
    return time.Until(t)
}
```

**Issue:** If the server sends a Retry-After date in the past, `time.Until(t)` returns negative duration, causing immediate retry.

**Impact:** None - immediate retry is acceptable behaviour.

---

## Verified Correct Logic

### Issue Link API Swap

**File:** `internal/cli/issue_link_add.go:121-129`

The code intentionally swaps `inwardIssue` and `outwardIssue` to match CLI semantics with Jira's counterintuitive API naming. This is correctly documented and implemented:

```go
// ajira issue link add GCP-123 Blocks GCP-456
// Results in: GCP-123 "blocks" GCP-456, GCP-456 "is blocked by" GCP-123
```

---

## Areas Reviewed

### Algorithm Correctness
- Unicode width calculation (`width.go`): Correctly handles CJK characters, combining marks, and zero-width characters
- ADF ↔ Markdown conversion: Properly handles nested marks, code blocks, tables, and task lists
- Pagination (`issue_list.go:322`): Has safety guard (`maxPages = 100`) to prevent infinite loops
- Validation functions (`validate.go`): Case-insensitive comparison with `strings.EqualFold` is correct

### Boundary Conditions
- Empty slices: Properly handled throughout (e.g., `renderNodes`, `mergeAdjacentTextNodes`)
- Nil pointers: Consistently checked before dereferencing (e.g., `issue_view.go:201-217`)
- Zero values: Appropriately handled for optional fields

### Type Conversions
- JSON unmarshalling handles optional fields correctly with pointer types
- Heading level conversion (`adf_to_markdown.go:94-105`) clamps values to valid range 1-6

### Control Flow
- All switch statements have appropriate default cases
- Early returns are used correctly throughout

### Concurrency
- The codebase is single-threaded per invocation
- No concurrency issues identified

---

## Files Reviewed

- `cmd/ajira/main.go`
- `internal/api/client.go`
- `internal/api/agile.go`
- `internal/jira/validate.go`
- `internal/jira/metadata.go`
- `internal/converter/adf.go`
- `internal/converter/adf_to_markdown.go`
- `internal/converter/markdown_to_adf.go`
- `internal/cli/root.go`
- `internal/cli/issue.go`
- `internal/cli/issue_create.go`
- `internal/cli/issue_edit.go`
- `internal/cli/issue_list.go`
- `internal/cli/issue_view.go`
- `internal/cli/issue_move.go`
- `internal/cli/issue_assign.go`
- `internal/cli/issue_comment.go`
- `internal/cli/issue_clone.go`
- `internal/cli/issue_delete.go`
- `internal/cli/issue_link_add.go`
- `internal/cli/epic.go`
- `internal/cli/epic_add.go`
- `internal/cli/sprint.go`
- `internal/cli/sprint_add.go`
- `internal/cli/batch.go`
- `internal/cli/exitcodes.go`
- `internal/cli/help.go`
- `internal/config/config.go`
- `internal/width/width.go`

---

## Conclusion

The ajira codebase demonstrates solid implementation with proper error handling and edge case coverage. The identified issues are minor and unlikely to cause problems in normal CLI usage. The code follows Go idioms well and maintains consistent patterns across commands.
