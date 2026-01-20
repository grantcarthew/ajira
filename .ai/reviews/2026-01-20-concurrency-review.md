# Go Concurrency Code Review

**Date:** 2026-01-20
**Repository:** ajira
**Reviewer:** Claude Opus 4.5

## Executive Summary

**Concurrency Status: Low Risk**

The ajira codebase has minimal concurrency usage and passes all race detector tests. The architecture is fundamentally synchronous, appropriate for a CLI tool.

## 1. Goroutine Analysis

**Finding: No explicit goroutines spawned**

Searched for:
- `go func` patterns
- `go <functionName>(` patterns

Result: Zero goroutine creation points in application code.

## 2. Channel Usage

**Location:** `internal/api/client.go:155-159`

```go
select {
case <-ctx.Done():
    return nil, ctx.Err()
case <-time.After(retryAfter):
}
```

**Assessment: Correctly implemented**
- Properly handles context cancellation during rate-limit backoff
- No channel leaks possible (using `time.After` in `select` is fine here since the function returns after)
- Context errors propagated correctly

## 3. Shared State Analysis

**Package-level variables identified:**

| Location | Variable | Access Pattern | Risk |
|----------|----------|----------------|------|
| `internal/api/client.go:26` | `verboseWriter` | Write-once on startup, read-only during execution | **Low** |
| `internal/cli/root.go:16-29` | `jsonOutput`, `project`, `board`, `dryRun`, etc. | Set via Cobra flags before command execution | **Low** |
| `internal/width/width.go` | `combiningRanges`, `zeroWidthRanges`, `wideRanges` | Immutable lookup tables | **None** |

**verboseWriter Details:**
- Set in `PersistentPreRun` via `api.SetVerboseOutput(os.Stderr)` at `root.go:79`
- Only read thereafter during HTTP requests
- In CLI context: safe (single-threaded execution)
- Library usage note: Would need mutex protection if package used concurrently

## 4. Synchronisation Primitives

**Finding: None used**

Searched for:
- `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync.Map`
- `atomic.` operations

Result: No synchronisation primitives - consistent with the sequential execution model.

## 5. Context Handling

**Assessment: Good practices observed**

1. **Root command** (`internal/cli/root.go:131-133`):
   ```go
   ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
   defer cancel()
   return rootCmd.ExecuteContext(ctx)
   ```
   - Properly sets up signal-based cancellation
   - Context passed through Cobra to all commands

2. **API client** (`internal/api/client.go:121`):
   - Uses `http.NewRequestWithContext(ctx, ...)`
   - HTTP client respects context for cancellation/timeout

3. **Tests**: Properly test context cancellation (`client_test.go:204-246`)

## 6. Race Detector Results

```
go test -race ./...
```

**Result: All packages pass with no race conditions detected**

```
ok  github.com/gcarthew/ajira/internal/api      1.752s
ok  github.com/gcarthew/ajira/internal/cli      2.599s
ok  github.com/gcarthew/ajira/internal/config   3.538s
ok  github.com/gcarthew/ajira/internal/converter 1.489s
ok  github.com/gcarthew/ajira/internal/jira     2.792s
ok  github.com/gcarthew/ajira/internal/width    3.153s
```

Race-enabled build also succeeds.

## 7. Test Concurrency

**Finding: Tests are sequential**

- No `t.Parallel()` usage detected
- Each test creates isolated `httptest.NewServer` instances
- No shared test state between tests

## Issues Identified

**Severity: None critical, 1 informational**

### INFO-1: Global verboseWriter lacks thread-safety for library use

**Location:** `internal/api/client.go:26-30`

```go
var verboseWriter io.Writer

func SetVerboseOutput(w io.Writer) {
    verboseWriter = w
}
```

**Current Risk:** None for CLI usage
**Potential Risk:** If the `api` package is used as a library in a concurrent application, simultaneous calls to `SetVerboseOutput` and HTTP requests could race.

**Recommendation (only if library usage is planned):**
```go
var verboseWriter atomic.Pointer[io.Writer]

func SetVerboseOutput(w io.Writer) {
    verboseWriter.Store(&w)
}
```

Or use `sync.Once` if only set once at startup.

## Checklist Summary

| Category | Status | Notes |
|----------|--------|-------|
| Race conditions | Pass | Race detector finds none |
| Goroutine leaks | N/A | No goroutines created |
| Channel misuse | Pass | Single `select` correctly implemented |
| Context propagation | Pass | Context threaded through properly |
| Deadlock potential | N/A | No locks used |
| Shared state protection | Pass | CLI single-threaded model safe |

## Recommendations

1. **No immediate action required** - The codebase is concurrency-safe for its intended CLI use case.

2. **Consider adding `-race` to CI** if not already present:
   ```yaml
   go test -race ./...
   ```

3. **If exposing as library**: Add atomic/mutex protection to `verboseWriter` global.
