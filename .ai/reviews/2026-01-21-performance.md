# Performance Review

Review analysing code efficiency, resource usage, and potential optimisation opportunities in the ajira CLI.

## Summary

The ajira codebase has **good overall performance characteristics** for a CLI tool. Resource lifecycle is handled correctly (HTTP responses are closed, contexts are propagated), pagination prevents unbounded data fetching, and rate limiting is implemented. However, there are several micro-optimisation opportunities in string handling, slice allocation, and repeated operations that could provide marginal improvements.

## Context

ajira is a non-interactive CLI designed for automation and AI agents. Performance considerations:
- Typical usage: single command execution, not long-running
- Network latency to Jira API dominates execution time
- Local processing (string conversion, JSON parsing) is secondary
- Memory usage is bounded by API response sizes

Given this context, most findings are **low priority** - the network round-trip time dwarfs any local processing overhead. However, for completeness and potential high-volume batch operations, optimisations are documented.

## Hot Paths Identified

| Path | Description | Performance Impact |
|------|-------------|-------------------|
| `internal/api/client.go:105-188` | HTTP request execution | Network-bound, well-optimised |
| `internal/converter/` | Markdown/ADF conversion | CPU-bound, medium frequency |
| `internal/cli/issue_list.go:314-377` | Issue search pagination | Network + memory |
| `internal/width/width.go` | Unicode width calculation | CPU-bound, high frequency |

## Findings

### Critical Issues

None identified.

### High Severity

None identified. The codebase follows good practices for a CLI tool.

### Medium Severity

**M1: No caching for validation metadata**

Validation functions fetch metadata from the Jira API for each invocation:

| Location | API Call |
|----------|----------|
| `internal/jira/validate.go:18` | `GetPriorities()` |
| `internal/jira/validate.go:41` | `GetIssueTypes()` |
| `internal/jira/validate.go:64` | `GetStatuses()` |
| `internal/jira/validate.go:87` | `GetLinkTypes()` |

Commands that validate multiple fields make redundant API calls. For example, `issue create` with `--type` and `--priority` flags results in 2 validation API calls.

**Impact**: Each validation adds ~50-200ms network latency. For commands with multiple validations, this compounds.

**Example workflow**:
```
issue list --status "In Progress" --type Bug --priority High
```
This triggers 3 separate API calls for validation, plus the actual search.

**Recommendation**: Consider in-memory caching of metadata within a single CLI invocation:
```go
var (
    prioritiesCache     []Priority
    prioritiesCacheOnce sync.Once
)

func GetPrioritiesCached(ctx context.Context, client *api.Client) ([]Priority, error) {
    var err error
    prioritiesCacheOnce.Do(func() {
        prioritiesCache, err = GetPriorities(ctx, client)
    })
    return prioritiesCache, err
}
```

**Priority**: Medium - reduces latency for multi-field validation.

**M2: Repeated color object allocation**

Color functions create new `color.Color` objects on every call:

| Location | Pattern |
|----------|---------|
| `internal/cli/issue_list.go:381-383` | `color.New()` called per status |
| `internal/cli/sprint_list.go:227-229` | `color.New()` called per state |
| `internal/cli/issue_list.go:177-179` | `color.New()` called per list display |
| `internal/cli/epic_list.go:87-89` | `color.New()` called per list display |

**Impact**: Minor - allocates on heap, adds GC pressure for large lists.

**Recommendation**: Move color objects to package-level variables:
```go
var (
    boldStyle   = color.New(color.Bold)
    faintStyle  = color.New(color.Faint)
    greenStyle  = color.New(color.FgGreen)
    // etc.
)
```

**Priority**: Low - micro-optimisation with minimal real-world impact.

### Low Severity

**L1: String concatenation in loops**

Several locations use `+=` for string building in loops instead of `strings.Builder`:

| Location | Code Pattern |
|----------|--------------|
| `internal/converter/adf_to_markdown.go:118-122` | `code += child.Text` in loop |
| `internal/converter/markdown_to_adf.go:536-539` | `textContent += string(...)` in loop |
| `internal/cli/epic_list.go:165-172` | `jql +=` in loop |

**Impact**: Minimal for typical input sizes. Each `+=` creates a new string allocation.

**Recommendation**: Use `strings.Builder` for multi-iteration concatenation:
```go
var buf strings.Builder
for _, child := range node.Content {
    if child.Type == NodeTypeText {
        buf.WriteString(child.Text)
    }
}
code := buf.String()
```

**L2: Slice growth without capacity hints**

Several slices grow via `append` without initial capacity:

| Location | Declaration |
|----------|-------------|
| `internal/cli/issue_list.go:315` | `var allIssues []IssueInfo` |
| `internal/jira/metadata.go:134` | `var statuses []Status` |
| `internal/cli/issue_view.go:324` | `var comments []CommentInfo` |
| `internal/converter/adf_to_markdown.go:30` | `var parts []string` |

**Impact**: Causes multiple reallocations as slice grows. For typical sizes (10-50 items), this is negligible.

**Recommendation**: Pre-allocate when size is known or estimable:
```go
// When maxResults is known
issues := make([]IssueInfo, 0, maxResults)

// When iterating a known-size source
statuses := make([]Status, 0, len(resp))
```

**L3: Linear search in unicode range lookup**

`internal/width/width.go:89-96` performs linear search through unicode ranges:

```go
func inRanges(r rune, ranges []runeRange) bool {
    for _, rng := range ranges {
        if r >= rng.lo && r <= rng.hi {
            return true
        }
    }
    return false
}
```

The `wideRanges` slice contains 50+ ranges, searched linearly for each rune.

**Impact**: For ASCII text, this exits quickly (no match). For CJK text or long strings with emoji, linear search adds overhead.

**Recommendation**: Use binary search for sorted ranges:
```go
func inRanges(r rune, ranges []runeRange) bool {
    i := sort.Search(len(ranges), func(i int) bool {
        return ranges[i].hi >= r
    })
    return i < len(ranges) && r >= ranges[i].lo
}
```

Alternatively, for hot paths, use a bitmap or hash for common ranges.

**L4: Fallback JSON parsing on failure**

`internal/jira/metadata.go:101-108` attempts to parse JSON twice on failure:

```go
if err := json.Unmarshal(body, &resp); err != nil {
    var types []issueTypeResponse
    if err2 := json.Unmarshal(body, &types); err2 != nil {
        return nil, fmt.Errorf("failed to parse response...")
    }
    resp.IssueTypes = types
}
```

**Impact**: Double parsing only occurs on error path. Minimal impact in normal operation.

**Recommendation**: Acceptable as-is. This handles API version differences gracefully.

**L5: UUID generation for every task item**

`internal/converter/markdown_to_adf.go:253,294` generates UUIDs for task lists and items:

```go
Attrs: map[string]any{"localId": uuid.New().String()},
```

**Impact**: UUID generation has some overhead, but task lists are infrequent.

**Recommendation**: Acceptable as-is. The `google/uuid` package is efficient, and task lists are uncommon in typical issue descriptions.

## Positive Patterns

### P1: Correct HTTP resource lifecycle

`internal/api/client.go:142` properly closes response bodies:
```go
defer resp.Body.Close()
```

This is critical for connection reuse and preventing file descriptor leaks.

### P2: Pagination with limits

`internal/cli/issue_list.go:316-324` implements bounded pagination:
```go
maxResults := 50
if limit > 0 && limit < maxResults {
    maxResults = limit
}
// ...
const maxPages = 100 // Safety guard
```

This prevents unbounded memory growth and runaway API calls.

### P3: Context propagation

All API calls properly accept and use `context.Context`, enabling cancellation:
```go
func (c *Client) Get(ctx context.Context, path string) ([]byte, error)
```

### P4: Rate limiting with backoff

`internal/api/client.go:149-161` implements exponential backoff for rate limits:
```go
if resp.StatusCode == 429 && attempt < maxRetries {
    retryAfter := getRetryAfter(resp, attempt)
    // ...
    return c.doRequestWithRetry(ctx, method, path, body, attempt+1)
}
```

### P5: Efficient slice allocation in some paths

`internal/jira/metadata.go:81-86` pre-allocates result slice:
```go
priorities := make([]Priority, len(resp))
for i, p := range resp {
    priorities[i] = Priority(p)
}
```

### P6: Embedded assets avoid file I/O

`internal/cli/help.go:10-16` uses `//go:embed` for help content:
```go
//go:embed help/agents.md
var agentsHelp string
```

This eliminates runtime file I/O for static content.

## Benchmarking Recommendations

The following operations would benefit from benchmarks if performance becomes a concern:

### Width calculation
```go
func BenchmarkStringWidth(b *testing.B) {
    texts := []string{
        "simple ascii",
        "with emoji 🚀",
        "日本語テキスト",
        strings.Repeat("x", 1000),
    }
    for _, text := range texts {
        b.Run(text[:min(20, len(text))], func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                StringWidth(text)
            }
        })
    }
}
```

### ADF conversion
```go
func BenchmarkMarkdownToADF(b *testing.B) {
    // Test with various markdown sizes
    small := "Simple paragraph"
    large := strings.Repeat("# Heading\nParagraph text.\n", 100)
    // ...
}
```

## Profiling Guidance

If performance issues are suspected:

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/converter/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/converter/
go tool pprof mem.prof

# Escape analysis (shows heap allocations)
go build -gcflags="-m" ./internal/converter/ 2>&1 | grep "escapes to heap"
```

## Recommendations Summary

| Priority | Issue | Recommendation | Expected Benefit |
|----------|-------|----------------|------------------|
| Medium | M1 | Cache validation metadata per invocation | 50-200ms per avoided call |
| Low | M2 | Move color objects to package level | Reduced allocations |
| Low | L1 | Use strings.Builder in loops | Marginal allocation reduction |
| Low | L2 | Pre-allocate slices with capacity | Fewer reallocations |
| Low | L3 | Binary search for unicode ranges | Faster CJK/emoji handling |

## Conclusion

ajira's performance is appropriate for its use case. The primary bottleneck is network latency to the Jira API, not local processing. The findings above are micro-optimisations that would provide marginal improvements.

**Recommended actions:**
1. **Consider** M1 (validation caching) if users report slow multi-flag commands
2. **Defer** other optimisations until profiling indicates a need

The codebase follows good practices: proper resource cleanup, bounded operations, and efficient patterns where they matter most (HTTP handling). No performance-related bugs or critical issues were identified.
