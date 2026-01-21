# Error Handling Review

Review of Go error handling patterns, wrapping practices, and edge case coverage.

## Summary

The ajira codebase demonstrates **strong error handling practices** overall. The architecture includes well-designed custom error types (`APIError`, `ExitError`), proper use of `errors.As()` for type checking, and consistent error wrapping with context. However, there are some inconsistencies in error wrapping style and a few places where errors are silently handled or lack context.

## Findings

### Critical Issues

None identified.

### High Severity

**H1: Inconsistent error wrapping loses error chain**

Several places use `fmt.Errorf("%v", err)` instead of `fmt.Errorf("...: %w", err)`, which breaks the error chain and prevents `errors.Is()` and `errors.As()` from working correctly on the returned error.

| Location | Pattern |
|----------|---------|
| `internal/cli/issue_create.go:122` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_create.go:129` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_create.go:132` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_view.go:141` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_list.go:125` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_list.go:132` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_list.go:137` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_list.go:144` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_edit.go:122` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_edit.go:132` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_edit.go:135` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_clone.go:109` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_clone.go:144` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_clone.go:152` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_clone.go:163` | `return fmt.Errorf("%v", err)` |
| `internal/cli/issue_link_add.go:61` | `return fmt.Errorf("%v", err)` |
| `internal/cli/me.go:41` | `return fmt.Errorf("%v", err)` |

**Recommendation**: Replace `fmt.Errorf("%v", err)` with `fmt.Errorf("context: %w", err)` or just `return err` if no additional context is needed.

### Medium Severity

**M1: Silent error swallowing in ADFToMarkdown**

`internal/converter/adf_to_markdown.go:11-16` silently returns an empty string when JSON unmarshalling fails. This hides parsing errors from callers.

```go
func ADFToMarkdown(adfJSON []byte) string {
    var doc ADF
    if err := json.Unmarshal(adfJSON, &doc); err != nil {
        return ""  // Error is silently swallowed
    }
    return renderNodes(doc.Content, 0)
}
```

**Impact**: Callers have no way to distinguish between empty content and a parsing failure.

**Recommendation**: Consider returning `(string, error)` or logging the parsing error in verbose mode.

**M2: Bare error returns lose context**

Some functions return errors without adding context, making debugging harder:

| Location | Pattern |
|----------|---------|
| `internal/cli/issue_delete.go:51` | `return err` from `config.Load()` |
| `internal/cli/issue_delete.go:60` | `return err` from `ReadKeysFromStdin()` |
| `internal/cli/issue_assign.go:69` | `return err` from `config.Load()` |
| `internal/cli/issue_assign.go:83` | `return err` from `ReadKeysFromStdin()` |
| `internal/cli/issue_comment.go:107` | `return err` from `config.Load()` |
| `internal/cli/issue_comment.go:119` | `return err` from `ReadKeysFromStdin()` |
| `internal/cli/issue_move.go:91` | `return err` from `config.Load()` |
| `internal/cli/sprint_add.go:69` | `return err` from `config.Load()` |
| `internal/cli/user.go:59` | `return err` from `config.Load()` |

**Recommendation**: Consider wrapping with context like `return fmt.Errorf("loading config: %w", err)` for consistency with other commands. Note: This is a style preference - bare returns are acceptable for well-typed errors from `config.Load()` since the error message already includes context.

**M3: Inconsistent error type checking style**

The codebase mixes type assertion and `errors.As()`:

- Type assertion: `if apiErr, ok := err.(*api.APIError); ok { ... }`
- errors.As: Used in `internal/cli/exitcodes.go`

The type assertion style works but doesn't handle wrapped errors. While this isn't currently a problem (API errors aren't wrapped), it could become an issue if error wrapping is added.

**Recommendation**: Prefer `errors.As()` for consistency:
```go
var apiErr *api.APIError
if errors.As(err, &apiErr) {
    return fmt.Errorf("API error: %w", apiErr)
}
```

### Low Severity

**L1: Deferred Close without error handling**

`internal/api/client.go:142` uses `defer resp.Body.Close()` without checking the error return value.

```go
defer resp.Body.Close()
```

**Impact**: Minimal - `Close()` errors on read operations are generally safe to ignore. This is an accepted Go convention.

**Recommendation**: No action required, but for completeness in highly reliable systems:
```go
defer func() {
    if err := resp.Body.Close(); err != nil {
        // Log if verbose mode, or ignore
    }
}()
```

**L2: RenderMarkdown silently falls back**

`internal/cli/root.go:195-206` silently falls back to plain text when markdown rendering fails. This is correct behaviour but could be logged in verbose mode.

**L3: User search returns empty result vs error**

`internal/cli/issue_assign.go:184-186` returns empty string when no users are found, which callers must check separately from errors:

```go
if len(users) == 0 {
    return "", nil
}
```

The callers do handle this correctly (e.g., `issue_assign.go:108`), but it's an unusual pattern.

## Positive Patterns

The codebase demonstrates several excellent error handling practices:

### Custom Error Types

**APIError** (`internal/api/client.go:53-76`) provides rich context:
- HTTP status code and status text
- Jira API error messages and field-specific errors
- Method and path for debugging
- Raw body as fallback when JSON parsing fails

**ExitError** (`internal/cli/exitcodes.go:20-35`) with proper `Unwrap()`:
```go
func (e *ExitError) Unwrap() error {
    return e.Err
}
```

### Exit Code Classification

`internal/cli/exitcodes.go:42-80` properly classifies errors:
- Uses `errors.As()` for type checking
- Distinguishes auth errors (401/403) from other API errors
- Handles network and DNS errors separately
- Provides meaningful exit codes

### Error Aggregation

`internal/config/config.go:29-64` collects multiple validation errors using `errors.Join()`:
```go
var errs []error
// ... append errors ...
if len(errs) > 0 {
    return nil, errors.Join(errs...)
}
```

### Validation Before API Calls

The codebase validates inputs before making API calls:
- Issue type validation (`jira.ValidateIssueType`)
- Priority validation (`jira.ValidatePriority`)
- Status validation (`jira.ValidateStatus`)
- Link type validation (`jira.ValidateLinkType`)

These validations provide clear error messages listing valid options.

### Rate Limit Retry

`internal/api/client.go:148-162` implements proper retry with:
- Retry-After header parsing
- Exponential backoff fallback
- Maximum retry limit
- Context cancellation support

### Batch Operation Results

`internal/cli/batch.go:44-86` properly handles partial failures:
- Returns `ExitPartial` when some operations succeed
- Returns `ExitAPIError` when all fail
- Collects and reports individual errors

## Recommendations Summary

| Priority | Issue | Recommendation |
|----------|-------|----------------|
| High | H1 | Replace `%v` with `%w` in error wrapping |
| Medium | M1 | Consider returning error from `ADFToMarkdown` |
| Medium | M2 | Add context to bare error returns for consistency |
| Medium | M3 | Use `errors.As()` instead of type assertions |
| Low | L1-L3 | No action required |

## Edge Case Coverage

### Nil Handling
- Nil pointer checks are used consistently (e.g., `issue_view.go:201-217`)
- Nil slices are handled correctly (safe to range over)

### Empty Values
- Empty strings are validated at command entry points
- Empty slices handled with appropriate messaging ("No issues found")

### Bounds
- Pagination safety guard in `issue_list.go:324` (`maxPages = 100`)
- Integer bounds for heading levels in `adf_to_markdown.go:100-105`

### Context Cancellation
- All API operations accept context and propagate cancellation
- Rate limit retry respects context cancellation

## Conclusion

The error handling in ajira is well-architected with proper custom types, exit codes, and validation. The main areas for improvement are:

1. Standardising error wrapping to use `%w` consistently
2. Converting type assertions to `errors.As()` for future-proofing
3. Adding context to bare error returns for debugging consistency

None of these issues affect current functionality - they're improvements for maintainability and debugging.
