# Duplication Review

Code review focused on identifying repeated patterns that may benefit from consolidation.

**Date:** 2026-01-21
**Tool:** golangci-lint with dupl linter (threshold: 50 tokens)
**Findings:** 16 duplication issues detected

---

## Summary

The codebase shows two main categories of duplication:

1. **Production Code**: High duplication in agile command pairs (epic/sprint add)
2. **Test Code**: Repeated API error testing patterns

The project already demonstrates good consolidation practices with shared helpers in `batch.go` (ReadKeysFromStdin, PrintDryRunBatch, PrintSuccess, etc.). The remaining duplication is mostly acceptable given Go's preference for clarity over abstraction.

---

## 1. High-Priority: Epic/Sprint Add Commands

**Files:**
- `internal/cli/epic_add.go` (116 lines)
- `internal/cli/sprint_add.go` (116 lines)

**Issue:** Near-identical implementations for adding issues to epics and sprints.

**Differences:**
| Aspect | Epic | Sprint |
|--------|------|--------|
| Target identifier | `epicKey` (issue key) | `sprintID` (numeric ID) |
| API path | `/epic/{key}/issue` | `/sprint/{id}/issue` |
| JSON key in output | `epicKey` | `sprintId` |
| Command examples | Uses epic keys | Uses sprint IDs |

**Analysis:** These commands share ~95% identical structure:
- Same stdin handling logic
- Same dry-run logic
- Same success/error output pattern
- Same JSON request structure (`{issues: [...]}`)

**Recommendation:** **Consider consolidation** but low priority.

A generic helper could handle the common flow:
```go
type agileAddConfig struct {
    targetType   string  // "epic" or "sprint"
    targetID     string
    apiPath      string
    jsonKey      string
}

func runAgileAdd(ctx context.Context, cfg agileAddConfig, issueKeys []string) error
```

However, the current duplication is acceptable because:
- The files are self-contained and easy to understand
- Future divergence is possible (epics/sprints have different semantics)
- Abstraction would add indirection for minimal benefit
- Total code is only ~230 lines across both files

---

## 2. Low-Priority: Test API Error Patterns

**Files:**
- `internal/cli/board_test.go:144-167` duplicates `internal/cli/user_test.go:105-128`
- `internal/cli/epic_test.go:111-141` duplicates `internal/cli/sprint_test.go:169-199`
- `internal/cli/issue_test.go:1182-1208` duplicates `internal/cli/issue_test.go:1211-1237`

**Issue:** Repeated patterns for testing API error handling.

**Example Pattern:**
```go
func TestXxx_APIError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusXxx)
        _ = json.NewEncoder(w).Encode(map[string]any{
            "errorMessages": []string{"Error message"},
        })
    }))
    defer server.Close()

    client := api.NewClient(testConfig(server.URL))
    _, err := functionUnderTest(context.Background(), client, ...)
    // assertions...
}
```

**Recommendation:** **No action required.**

Test code duplication is acceptable because:
- Test clarity is more important than DRY
- Each test documents the expected behaviour for its specific function
- A table-driven approach would obscure the intent
- The pattern is consistent and idiomatic Go testing

---

## 3. Structural Patterns (Not Duplication)

Several patterns appear similar but are appropriately implemented separately:

### List Command Pattern
Each list command (board, sprint, project, release, field, epic, user) follows a similar structure:
1. Load config
2. Create client
3. Fetch data with pagination
4. Output JSON or tabular format

**Analysis:** These are not duplication candidates because:
- Each has unique API endpoints and response structures
- Display formatting differs (tabwriter vs coloured output)
- Type-specific filtering and transformations
- Go generics would add complexity without benefit here

### Pagination Pattern
The `fetchAllX` functions share a similar loop structure:
```go
for {
    path := fmt.Sprintf(...)
    body, err := client.Get(ctx, path)
    // unmarshal and accumulate
    if resp.IsLast { break }
    startAt += maxResults
}
```

**Analysis:** **No action required.**
- Each function handles different response types
- Go's type system makes generic pagination helpers verbose
- The pattern is clear and consistent

### JQL Building Pattern
Both `issue_list.go` and `epic_list.go` build JQL queries with similar logic:
- Start with project condition
- Append filters based on flags
- Join with " AND "

**Analysis:** **Minor improvement possible** but low priority.
A JQL builder helper could reduce the string manipulation, but:
- Current implementation is readable
- Only two locations use this pattern
- JQL edge cases differ between issue types

---

## 4. Detected but Acceptable Duplication

### Transition Tests
`internal/cli/issue_test.go:1182-1237`

Two tests (`TestDoTransition_WithResolution` and `TestDoTransition_WithAssignee`) share the same structure but test different field handling. These are intentionally separate to document each case clearly.

### Add Issues Tests
`internal/cli/epic_test.go` and `internal/cli/sprint_test.go` contain similar success tests for adding issues. These correspond to the production code duplication and would benefit from the same refactoring (or none).

---

## 5. Related Findings: Unchecked Errors

The linter also detected 20 unchecked error returns. Notable ones:

| File | Line | Issue |
|------|------|-------|
| `api/client.go` | 142 | `resp.Body.Close()` |
| `field.go` | 110, 116, 118 | `fmt.Fprintln`, `fmt.Fprintf`, `w.Flush` |
| `project.go` | 107, 111 | `fmt.Fprintln`, `w.Flush` |
| `release.go` | 116, 131 | `fmt.Fprintln`, `w.Flush` |

**Recommendation:** These are low risk for a CLI tool but could be addressed for completeness. The `resp.Body.Close()` error is safe to ignore. The tabwriter flush errors are unlikely but should technically be checked.

---

## Recommendations Summary

| Priority | Finding | Action |
|----------|---------|--------|
| Low | Epic/Sprint add duplication | Consider consolidation if more agile add commands are added |
| None | Test API error patterns | Accept - test clarity over DRY |
| None | List command structure | Accept - appropriate separation |
| None | Pagination patterns | Accept - type-specific requirements |
| Low | Unchecked tabwriter errors | Consider adding error checks |

---

## Conclusion

The codebase has minimal problematic duplication. The existing `batch.go` helper module demonstrates good consolidation practices. The detected duplication in epic/sprint add commands is borderline - consolidation is possible but not necessary given the small codebase size and potential for feature divergence.

The Go philosophy of "a little copying is better than a little dependency" applies here. The code is readable, maintainable, and each file is self-contained.
