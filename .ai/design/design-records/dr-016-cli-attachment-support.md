# dr-016: Issue Attachment Support

- Date: 2026-01-23
- Status: Accepted
- Category: CLI

## Problem

Users need to manage file attachments on Jira issues without leaving the CLI:

- Upload files to issues (screenshots, logs, documentation, etc.)
- List attachments to see what files are present
- Download attachments for local review
- Remove attachments that are no longer needed
- View attachment information when viewing issue details

Without these capabilities, users must switch to the Jira web UI to manage attachments, breaking the CLI workflow.

## Decision

Add an `attachment` subcommand under `ajira issue` with four operations:

```
ajira issue attachment list KEY                         # List attachments for issue
ajira issue attachment add KEY FILE [FILE...]           # Upload one or more files
ajira issue attachment download KEY ID [-o OUTPUT]      # Download attachment
ajira issue attachment remove KEY ID [ID...] [--force]  # Delete attachments
```

Additionally:

- Display attachments in `ajira issue view` output by default
- Add `attachments` as plural alias for consistency with other commands
- Show attachment IDs in output for easy reference in download/remove operations

## Command Specifications

### ajira issue attachment list KEY

Lists all attachments for an issue.

API: `GET /rest/api/3/issue/{key}?fields=attachment`

Text output format:

```
Attachments for PROJ-123 (3):

ID      Filename           Size     Author      Created
10001   screenshot.png     45 KB    John Doe    2026-01-20 14:30
10002   requirements.pdf   1.2 MB   Jane Smith  2026-01-21 09:15
10003   data.csv          234 KB    John Doe    2026-01-23 08:00
```

JSON output includes all API fields: id, filename, size, mimeType, author, created, content (download URL), thumbnail (if available), self.

### ajira issue attachment add KEY FILE [FILE...]

Uploads one or more files to an issue.

API: `POST /rest/api/3/issue/{key}/attachments`

Request requirements:

- Content-Type: `multipart/form-data`
- Required header: `X-Atlassian-Token: no-check` (CSRF protection)
- Form field name: `file` (can repeat for multiple files)
- File paths are space-separated positional arguments

Examples:

```bash
ajira issue attachment add PROJ-123 screenshot.png
ajira issue attachment add PROJ-123 file1.pdf file2.png file3.doc
ajira issue attachment add PROJ-123 *.log
```

File size limits (Jira Cloud):

- Default maximum: 1 GB per file
- Absolute maximum: 2 GB per file
- Returns 413 Payload Too Large if exceeded

### ajira issue attachment download KEY ID [-o OUTPUT]

Downloads an attachment to the current directory.

API: `GET /rest/api/3/attachment/content/{id}`

Behavior:

- Without `-o`: Downloads to original filename (from API metadata)
- With `-o FILE`: Downloads to specified filename
- Streams download (does not load entire file into memory)

Examples:

```bash
ajira issue attachment download PROJ-123 10001                  # Downloads to ./screenshot.png
ajira issue attachment download PROJ-123 10001 -o custom.pdf    # Downloads to ./custom.pdf
```

### ajira issue attachment remove KEY ID [ID...] [--force]

Deletes one or more attachments from an issue.

API: `DELETE /rest/api/3/attachment/{id}`

Behavior:

- Accepts multiple attachment IDs
- Prompts for confirmation by default
- `--force` flag skips confirmation prompt
- Requires both issue key and attachment ID for safety

Examples:

```bash
ajira issue attachment remove PROJ-123 10001
ajira issue attachment remove PROJ-123 10001 10002 10003
ajira issue attachment remove PROJ-123 10001 --force
```

### Updated issue view

Add attachments section to `ajira issue view` output, displayed after links and before comments:

```
Description:
Need to add OAuth2 authentication for users...

Links:
  blocks PROJ-124 (To Do) - Fix login bug
  relates to PROJ-125 (Done) - Update documentation

Attachments (3):
  [10001] screenshot.png (45 KB) - John Doe, 2026-01-20 14:30
  [10002] requirements.pdf (1.2 MB) - Jane Smith, 2026-01-21 09:15
  [10003] data.csv (234 KB) - John Doe, 2026-01-23 08:00

------------------------------------------------------------
Comments (2):
...
```

Attachment data comes from the existing issue fetch (all fields returned by default). No additional API calls required.

Display order: Description → Links → Attachments → Comments

## Implementation Notes

### Multipart Upload

The API client requires a new method for multipart form-data requests since the existing client only handles JSON payloads.

Use Go standard library `mime/multipart` package:

1. Create `multipart.NewWriter()` with body buffer
2. Use `CreateFormFile("file", filename)` for each file
3. Stream file content with `io.Copy()` to avoid loading entire file into memory
4. Call `writer.Close()` to finalize multipart message
5. Set `Content-Type` header to `writer.FormDataContentType()` (includes boundary)
6. Add `X-Atlassian-Token: no-check` header

### Error Handling

Special cases to handle:

- 413 Payload Too Large: Map to user-friendly "file exceeds size limit" message
- 403 Forbidden: User lacks attachment permissions
- 404 Not Found: Issue or attachment doesn't exist
- File I/O errors: File not found, permission denied
- Network errors during large uploads/downloads

### Size Display

Format file sizes in human-readable units:

- Bytes: 0-999 B
- Kilobytes: 1.0 KB - 999 KB
- Megabytes: 1.0 MB - 999 MB
- Gigabytes: 1.0 GB+

## Why

- Space-separated file arguments match standard CLI conventions (cp, mv, tar)
- Requiring issue key + attachment ID prevents accidental deletions
- Including attachment IDs in display enables easy copy-paste for download/remove
- Downloading to original filename by default is intuitive (matches browser behavior)
- Optional `-o` flag provides flexibility when needed
- Showing attachments in issue view maintains context without extra commands
- Streaming uploads/downloads handles large files efficiently

## Trade-offs

Accept:

- Need new API client method for multipart uploads (existing client is JSON-only)
- File paths in add command must be shell-escaped if they contain spaces
- No support for piping file content from stdin (future enhancement)
- Cannot rename attachments (Jira API limitation - must delete and re-upload)
- Download always uses current directory (no recursive directory structure)

Gain:

- Simple, intuitive command structure matching CLI conventions
- Efficient memory usage for large files via streaming
- Complete attachment workflow without leaving CLI
- Attachment IDs visible in all output for easy reference
- Consistent with existing command patterns (comment, link)

## Alternatives

Comma-separated file paths:

- Pro: Alternative syntax some users might expect
- Con: Filenames can contain commas, causing ambiguity
- Con: Not a standard CLI convention
- Con: Breaks shell globbing and tab completion
- Rejected: Space-separated is standard and unambiguous

Just attachment ID for download (no issue key):

- Pro: Shorter command
- Con: Less consistent with remove command
- Con: Less clear context for the operation
- Rejected: Consistency and clarity outweigh brevity

Flag-based file input (`-f file.pdf`):

- Pro: Matches comment pattern
- Con: Redundant since positional args already represent files
- Con: Different semantic meaning (comment `-f` reads content, attachment takes path)
- Rejected: Positional arguments are clearer for file paths

Separate `get` command instead of `download`:

- Pro: Shorter name
- Con: Ambiguous - could mean "get metadata" not "download file"
- Con: `list` already provides metadata
- Rejected: `download` is explicit and unambiguous

No confirmation prompt for remove:

- Pro: Faster workflow
- Con: Easy to accidentally delete attachments
- Rejected: Confirmation improves safety, `--force` available for scripts
