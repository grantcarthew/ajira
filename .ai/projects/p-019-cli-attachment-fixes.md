# p-019: CLI Attachment Fixes

- Status: Pending
- Started:
- Completed:

## Overview

Address issues identified in external code review of the p-018 CLI Attachment Support implementation. The review found discrepancies between design documentation and implementation, missing validation, code style inconsistencies, and test fragility.

The most significant issue is that dr-016 explicitly promises streaming I/O for large file handling, but the implementation buffers entire files in memory. This could cause out-of-memory errors when users upload or download large attachments (log archives, database dumps, video recordings).

Secondary issues include missing validation that attachments belong to specified issues, which could lead to user confusion or accidental operations on wrong attachments.

## Goals

1. Implement true streaming for attachment downloads to match design promise
2. Implement streaming or chunked uploads for large files
3. Add validation that attachment IDs belong to specified issues before download/remove operations
4. Replace custom number formatting with standard library functions
5. Fix import ordering to follow Go conventions
6. Improve test robustness with proper ID generation
7. Standardise JSON output patterns across attachment commands

## Scope

In Scope:

- Refactoring `GetRaw` method or creating streaming alternative
- Refactoring `PostMultipart` method or creating streaming alternative
- Adding attachment ownership validation to download and remove commands
- Replacing `intToStr`, `formatFloat1`, `formatFloat2` with `strconv` functions
- Fixing import order in `issue_attachment_list.go`
- Fixing test ID generation in `issue_attachment_test.go`
- Updating `issue_attachment_remove.go` to use `PrintSuccessJSON`
- Updating dr-016 if streaming approach changes

Out of Scope:

- New attachment features (covered in future projects)
- Changes to other commands
- Performance optimisation beyond streaming fix

## Current State

### Download Implementation

File: `internal/cli/issue_attachment_download.go`

Current flow:
1. Fetch attachment metadata via `getAttachmentMeta()` - returns filename, size
2. Call `client.GetRaw()` which returns `[]byte` (entire file in memory)
3. Write to disk via `os.WriteFile(outputFile, content, 0644)`

Problem: For a 500MB file, this allocates 500MB+ heap memory.

### Upload Implementation

File: `internal/cli/issue_attachment_add.go`

Current flow:
1. Validate all files exist
2. Create `bytes.Buffer` for multipart body
3. Loop through files, `io.Copy` each into buffer via `multipart.Writer`
4. Call `client.PostMultipart()` with `body.Bytes()`

Problem: Multiple large files accumulate entirely in memory before HTTP request starts.

### API Client Methods

File: `internal/api/client.go`

`GetRaw` (lines 117-170):
- Uses `io.ReadAll(resp.Body)` at line 163
- Returns `[]byte, string, error`

`PostMultipart` (lines 109-113):
- Accepts `[]byte` body parameter
- Passes to `doMultipartRequest`

`doMultipartRequest` (lines 172-232):
- Uses `bytes.NewReader(body)` - body already in memory

### Issue Key Validation

Download (`issue_attachment_download.go:57`):
```go
_ = args[0] // issueKey - used for context but not required by API
```

Remove (`issue_attachment_remove.go:38`):
```go
issueKey := args[0]
// Used only for display, not validated
```

Neither command verifies the attachment belongs to the specified issue.

### Custom Number Formatting

File: `internal/cli/issue_attachment.go` (lines 67-108)

Functions:
- `intToStr(i int64) string` - custom integer to string conversion
- `formatFloat1(f float64) string` - 1 decimal place formatting
- `formatFloat2(f float64) string` - 2 decimal place formatting

Used by `FormatFileSize()` for human-readable size display.

### Import Ordering Issue

File: `internal/cli/issue_attachment_list.go` (lines 7-8):
```go
"text/tabwriter"
"os"
```

Should be alphabetically sorted: `os` before `text/tabwriter`.

### Test ID Generation

File: `internal/cli/issue_attachment_test.go` (line 394):
```go
"id": "1000" + string(rune('3'+i)),
```

For i >= 7, produces non-numeric characters (`:`, `;`, `<`, etc.).

### JSON Output Inconsistency

`issue_attachment_add.go:85` uses helper:
```go
PrintSuccessJSON(result)
```

`issue_attachment_remove.go:76-80` manually marshals:
```go
output, err := json.MarshalIndent(result, "", "  ")
if err != nil {
    return fmt.Errorf("failed to format JSON: %w", err)
}
fmt.Println(string(output))
```

## Technical Approach

### Streaming Download

Create new API client method that streams directly to a file:

```
func (c *Client) DownloadToFile(ctx context.Context, path string, dest io.Writer) error
```

Implementation approach:
1. Build HTTP request as in `GetRaw`
2. Execute request, get `resp.Body`
3. Use `io.Copy(dest, resp.Body)` to stream directly
4. Return error only (no body needed)

Update download command:
1. Create file with `os.Create(outputFile)`
2. Defer file close
3. Call `client.DownloadToFile(ctx, path, file)`
4. On error, remove partial file

### Streaming Upload

Two approaches possible:

Approach A - Chunked Transfer Encoding:
- Use `io.Pipe()` with goroutine writing multipart form
- Set `Request.ContentLength = -1` for chunked encoding
- Risk: Some servers don't handle chunked uploads well

Approach B - Calculate Size First:
- Stat all files to get total size
- Calculate multipart overhead (boundaries, headers)
- Set `Content-Length` explicitly
- Stream via pipe with known length

Approach C - Accept Current Behaviour:
- Document that uploads buffer in memory
- Add file size validation/warning for large uploads
- Update dr-016 to reflect actual behaviour

Recommendation: Start with Approach B. If Jira Cloud has issues with streaming, fall back to Approach C with size limits.

### Attachment Ownership Validation

For both download and remove commands:

1. Fetch issue attachments: `GET /rest/api/3/issue/{key}?fields=attachment`
2. Extract attachment IDs from response
3. Verify requested ID(s) exist in the list
4. Proceed with operation or return descriptive error

This adds one API call per operation but ensures consistency and prevents accidents.

Helper function to add:
```
func validateAttachmentBelongsToIssue(ctx context.Context, client *api.Client, issueKey, attachmentID string) error
```

For remove with multiple IDs, validate all before deleting any.

### Number Formatting Replacement

Replace custom functions with:

```go
import "strconv"

// Instead of intToStr(i)
strconv.FormatInt(i, 10)

// Instead of formatFloat1(f)
strconv.FormatFloat(f, 'f', 1, 64)
// Then trim trailing ".0" if needed

// Instead of formatFloat2(f)
strconv.FormatFloat(f, 'f', 2, 64)
// Then trim trailing zeros
```

Maintain exact output format for existing tests.

### Import Fix

Reorder imports in `issue_attachment_list.go`:
```go
import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/gcarthew/ajira/internal/api"
    "github.com/gcarthew/ajira/internal/config"
    "github.com/spf13/cobra"
)
```

### Test ID Fix

Replace character arithmetic with explicit formatting:
```go
"id": fmt.Sprintf("1000%d", 3+i),
```

### JSON Output Standardisation

Replace manual marshalling in `issue_attachment_remove.go` with:
```go
PrintSuccessJSON(result)
```

## Success Criteria

- [ ] Download streams directly to file without buffering entire content in memory
- [ ] Upload either streams or has documented size limitations
- [ ] Download command validates attachment belongs to specified issue
- [ ] Remove command validates all attachment IDs belong to specified issue before deletion
- [ ] Custom number formatting functions replaced with strconv
- [ ] Import ordering follows Go conventions (verified by goimports)
- [ ] Test ID generation uses fmt.Sprintf for robustness
- [ ] All attachment commands use PrintSuccessJSON consistently
- [ ] All existing tests pass
- [ ] New tests added for validation behaviour
- [ ] dr-016 updated if streaming approach changes

## Deliverables

Files to modify:

- `internal/api/client.go` - Add streaming download method, possibly streaming upload
- `internal/cli/issue_attachment_download.go` - Use streaming download, add validation
- `internal/cli/issue_attachment_add.go` - Streaming upload or size limits
- `internal/cli/issue_attachment_remove.go` - Add validation, use PrintSuccessJSON
- `internal/cli/issue_attachment.go` - Replace custom formatting functions
- `internal/cli/issue_attachment_list.go` - Fix import ordering
- `internal/cli/issue_attachment_test.go` - Fix ID generation, add validation tests

Documentation:

- Update dr-016 if streaming approach changes from original design

## Testing Strategy

### Unit Tests

Streaming download:
- Mock server returns chunked response
- Verify file written correctly without full buffering
- Verify partial file cleanup on error

Streaming upload:
- Mock server receives multipart form
- Verify content integrity
- Test with multiple files

Validation:
- Test download with valid attachment ID for issue - succeeds
- Test download with invalid attachment ID for issue - fails with clear error
- Test remove with mix of valid/invalid IDs - fails before any deletion
- Test remove with all valid IDs - succeeds

### Manual Testing

Test with real Jira instance:
- Upload files of various sizes (1KB, 10MB, 100MB)
- Download same files, verify integrity
- Attempt download with wrong issue key - verify error
- Monitor memory usage during large file operations

## Decision Points

1. Upload streaming approach

- A: Chunked transfer encoding with io.Pipe (true streaming, may have server compatibility issues)
- B: Calculate content-length, stream with known size (safer, slightly more complex)
- C: Keep current buffering, add size warnings/limits (simplest, documents limitation)

2. Validation failure behaviour for remove with multiple IDs

- A: Fail fast - validate all IDs first, delete none if any invalid
- B: Partial success - delete valid IDs, report failures
- C: Transactional - if Jira API supports it, all-or-nothing

## Notes

The code review was generated by Claude Opus 4.5 as an external review of staged changes. All identified issues have been verified against the actual source code.

Jira Cloud attachment size limits:
- Default maximum: 1 GB per file
- Absolute maximum: 2 GB per file

These limits may influence the decision on upload streaming vs buffering with limits.
