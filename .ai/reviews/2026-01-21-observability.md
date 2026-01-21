# Observability Review

Review assessing whether the ajira CLI can be understood and debugged in production use.

## Summary

The ajira codebase has **minimal observability instrumentation**, which is appropriate for a non-interactive CLI tool designed for automation. The `--verbose` flag provides HTTP-level debugging, error messages include useful context (method, path, status), and sensitive data (API tokens) is properly protected. However, there are gaps in structured logging and operation-level diagnostics that could improve debuggability.

## Context

ajira is a CLI tool, not a long-running service. This means:
- Health checks, liveness/readiness probes are not applicable
- Metrics endpoints are not applicable
- Distributed tracing across service boundaries is not applicable

The relevant observability concerns for a CLI are:
1. Can users diagnose failures from output?
2. Is there enough context to debug issues?
3. Are sensitive values protected from accidental exposure?

## Findings

### Critical Issues

None identified.

### High Severity

**H1: No structured logging**

The codebase uses `fmt.Print*` and `fmt.Fprintf` throughout. There is no structured logging library (slog, zerolog, zap). All output is unstructured text to stdout/stderr.

| Location | Pattern |
|----------|---------|
| `internal/api/client.go:138` | `fmt.Fprintf(verboseWriter, "%s %s error (%s)\n", ...)` |
| `internal/api/client.go:145` | `fmt.Fprintf(verboseWriter, "%s %s %s (%s)\n", ...)` |
| `internal/api/client.go:152` | `fmt.Fprintf(verboseWriter, "Rate limited, retrying...")` |
| `internal/cli/issue_view.go:160` | `fmt.Fprintf(os.Stderr, "warning: ...")` |

**Impact**:
- Log aggregation and parsing is difficult for automation consumers
- No consistent field names across output
- Cannot easily filter or query logs programmatically

**Recommendation**: For a CLI tool targeting AI agents, this is a style choice rather than a defect. If structured output is desired, consider:
1. A `--log-format json` flag for machine-parseable diagnostics
2. Using slog with a conditional JSON handler when verbose mode is enabled

**Priority**: Low - acceptable for current use case.

### Medium Severity

**M1: Verbose mode is limited to HTTP layer**

The `--verbose` flag only enables HTTP request/response logging in `internal/api/client.go`. No other operations are logged in verbose mode:
- No logging of markdown-to-ADF conversion
- No logging of validation checks (issue type, priority, status)
- No logging of batch operation progress
- No logging of context cancellation or signal handling

| Location | Coverage |
|----------|----------|
| `internal/cli/root.go:79` | `api.SetVerboseOutput(os.Stderr)` |
| `internal/api/client.go:137-145` | HTTP method, path, status, duration |
| All other code | No verbose output |

**Impact**: When debugging non-HTTP issues (e.g., ADF conversion failures, validation errors), verbose mode provides no additional insight.

**Recommendation**: Consider extending verbose output to cover:
1. Validation operations: `"validating issue type 'Story' for project 'PROJ'"`
2. Conversion operations: `"converting markdown to ADF (123 bytes)"`
3. Batch progress: `"processing 3/10 issues"`

**M2: Silent fallbacks hide failures**

Several operations silently fall back or return empty values without logging:

| Location | Behaviour |
|----------|-----------|
| `internal/converter/adf_to_markdown.go:13` | Returns `""` on JSON parse failure |
| `internal/cli/root.go:199-206` | Falls back to plain markdown on render failure |
| `internal/cli/issue_assign.go:184` | Returns `""` when no users found |

**Impact**: Callers cannot distinguish between "no data" and "error occurred". The comment fetching in `issue_view.go:158-164` correctly logs failures in verbose mode - this pattern should be applied elsewhere.

**Recommendation**: Log failures in verbose mode before returning fallback values:
```go
if verboseWriter != nil {
    fmt.Fprintf(verboseWriter, "warning: ADF parse failed: %v\n", err)
}
return ""
```

**M3: No request correlation for concurrent usage**

When running multiple ajira commands in parallel (common in AI agent workflows), there's no way to correlate output back to specific invocations.

**Impact**: Debugging concurrent batch operations is difficult when stderr output is interleaved.

**Recommendation**: Consider adding an optional `--request-id` flag that prefixes all verbose output:
```
[req-abc123] GET /rest/api/3/issue/PROJ-1 200 OK (45ms)
```

### Low Severity

**L1: No duration logging for non-HTTP operations**

Only HTTP requests log duration. Other potentially slow operations are not timed:
- Markdown rendering (`glamour.Render`)
- ADF conversion (for large documents)
- Stdin reading (could block indefinitely)

**Impact**: Performance bottlenecks outside HTTP calls are invisible.

**L2: Rate limit retry logging could include more context**

`internal/api/client.go:151-153` logs retry attempts but not the rate limit header values:

```go
fmt.Fprintf(verboseWriter, "Rate limited, retrying in %s (attempt %d/%d)\n", retryAfter, attempt+1, maxRetries)
```

**Recommendation**: Include the Retry-After header value when present:
```go
fmt.Fprintf(verboseWriter, "Rate limited (Retry-After: %s), retrying in %s (attempt %d/%d)\n", ...)
```

## Sensitive Data Protection

### Positive Findings

The codebase properly protects sensitive data:

**S1: API token never logged**

The token is stored in the `Client` struct and only used in `SetBasicAuth()`:
- `internal/api/client.go:126`: `req.SetBasicAuth(c.email, c.token)`
- Token is never printed, logged, or included in error messages
- `APIError` struct (`client.go:54-62`) does not include credentials

**S2: Verbose output excludes request body**

HTTP verbose logging only includes method, path, status, and duration - not request/response bodies. This prevents accidental logging of sensitive issue content.

**S3: Config validation errors are safe**

`internal/config/config.go:36-57` reports missing variables by name but never logs the actual values:
```go
errs = append(errs, errors.New("missing required environment variable: JIRA_API_TOKEN"))
```

**S4: Test files use dummy credentials**

Test files use clearly fake values:
- `internal/api/client_test.go:18`: `Email: "test@example.com"`, `APIToken: "test-token"`
- No real credentials in codebase

### Potential Concern

**P1: Email address displayed in output**

`internal/cli/me.go:66` displays the user's email address. This is intentional (verifying auth) but could be a concern in shared terminal logs.

**Recommendation**: No action required - this is expected behaviour for the `me` command.

## Debugging Support

### Current State

| Capability | Status | Notes |
|------------|--------|-------|
| Verbose mode | Partial | HTTP-only |
| Error context | Good | APIError includes method, path, status, messages |
| Exit codes | Excellent | 5 distinct codes for error classification |
| Dry-run mode | Good | `--dry-run` shows planned actions |
| JSON output | Good | `--json` for machine parsing |
| Runtime log level change | N/A | CLI exits after one command |
| Profiling | None | No pprof or similar |

### Positive Patterns

**Exit Code Classification** (`internal/cli/exitcodes.go`):
- 0: Success
- 1: User/input error
- 2: API error (4xx/5xx)
- 3: Network error
- 4: Auth error (401/403)
- 5: Partial failure (batch)

**Rich API Error Messages** (`internal/api/client.go:64-76`):
```
GET /rest/api/3/issue/INVALID-123: 404 Not Found - Issue does not exist
```

**Validation with Valid Options** (`internal/jira/validate.go`):
Validation errors list available options, making debugging straightforward.

## Metrics and Tracing

### Not Applicable

As a CLI tool, ajira does not need:
- Prometheus/OpenTelemetry metrics endpoints
- Distributed tracing spans
- Health check endpoints
- Goroutine/memory metrics

### Potential Future Enhancement

If ajira were to add a "server mode" or be used in high-frequency automation, consider:
- Command execution counters
- API call latency histograms
- Error rate tracking

This is not a current requirement.

## Recommendations Summary

| Priority | Issue | Recommendation |
|----------|-------|----------------|
| Medium | M1 | Extend verbose mode to cover validation and conversion |
| Medium | M2 | Log silent fallbacks in verbose mode |
| Medium | M3 | Add optional request-id for concurrent debugging |
| Low | L1 | Consider timing non-HTTP operations |
| Low | L2 | Include more context in rate limit logging |
| Low | H1 | Structured logging - nice-to-have for automation |

## Key Questions Assessment

| Question | Answer |
|----------|--------|
| Can you reconstruct what happened from output alone? | Partially - HTTP calls yes, other operations no |
| Can you identify the source of errors? | Yes - APIError includes method, path, status |
| Can you measure system health and performance? | Partially - HTTP duration only |
| Is sensitive data protected? | Yes - tokens never logged |

## Conclusion

ajira's observability is appropriate for its use case as a CLI tool. The `--verbose` flag, rich error messages, and distinct exit codes provide adequate debugging capability for most scenarios. The main improvement opportunities are:

1. Extending verbose mode beyond HTTP operations
2. Logging silent fallbacks instead of hiding failures
3. Adding request correlation for concurrent usage

None of these are blocking issues - they would improve debuggability in edge cases.
