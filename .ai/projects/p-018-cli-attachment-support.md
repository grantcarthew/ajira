# p-018: Issue Attachment Support

- Status: Pending
- Started:
- Completed:

## Overview

Add comprehensive file attachment management to ajira, enabling users to upload, list, download, and delete attachments on Jira issues without leaving the CLI. This completes the core issue management workflow by handling the final major issue field type not yet supported.

## Goals

1. Implement attachment upload with multipart form-data support
2. Add attachment listing and viewing capabilities
3. Enable attachment downloads with streaming for large files
4. Support attachment deletion with confirmation
5. Integrate attachment display into issue view output
6. Handle Jira-specific requirements (CSRF headers, size limits)

## Scope

In Scope:

- `ajira issue attachment list` - List attachments for an issue
- `ajira issue attachment add` - Upload one or more files
- `ajira issue attachment download` - Download attachment to local file
- `ajira issue attachment remove` - Delete attachments with confirmation
- Display attachments in `ajira issue view` output
- Multipart form-data support in API client
- Human-readable file size formatting
- Streaming downloads for memory efficiency
- Error handling for size limits and permissions

Out of Scope:

- Attachment renaming (Jira API limitation)
- Batch upload from file list via stdin (future enhancement)
- Thumbnail preview or display (metadata only)
- Attachment compression or transformation
- Progress bars for uploads/downloads (initial version)
- Recursive directory uploads

## Success Criteria

- [ ] Can upload single and multiple files to an issue
- [ ] Can list all attachments with metadata (ID, filename, size, author, date)
- [ ] Can download attachments to current directory
- [ ] Can delete attachments with confirmation prompt
- [ ] Attachments appear in `ajira issue view` output after links
- [ ] File sizes display in human-readable format (KB, MB, GB)
- [ ] Large files stream efficiently without loading into memory
- [ ] 413 errors map to user-friendly "file too large" messages
- [ ] All commands support --json output mode
- [ ] Created dr-016 documenting command structure and decisions
- [ ] Test suite covers upload, list, download, delete operations

## Deliverables

- `internal/api/client.go` - Add PostMultipart method for multipart uploads
- `internal/cli/issue_attachment.go` - Main attachment command implementation
- `internal/cli/issue_attachment_list.go` - List subcommand
- `internal/cli/issue_attachment_add.go` - Upload subcommand
- `internal/cli/issue_attachment_download.go` - Download subcommand
- `internal/cli/issue_attachment_remove.go` - Delete subcommand
- `internal/cli/issue_view.go` - Update to display attachments
- `internal/cli/format.go` or similar - Human-readable size formatting
- Test files for attachment operations
- dr-016: Issue Attachment Support (already created)

## Current State

Relevant codebase context:

- API client (`internal/api/client.go`) only supports JSON payloads
- Client methods: Get, Post, Put, Delete - all use `application/json`
- Issue view (`internal/cli/issue_view.go`) fetches all fields by default
- Issue view displays: description, links, comments (with -c flag)
- Comments use similar pattern: add, edit subcommands
- Links use similar pattern: add, remove, types, url subcommands
- File input pattern exists in comment command: -f flag, stdin support
- Batch operations use --stdin flag (comment add)
- Global flags: --json, --dry-run, --verbose, --quiet

## Technical Approach

API Client Enhancement:

- Add PostMultipart method to Client struct
- Use Go standard library mime/multipart package
- Stream file content with io.Copy (not ReadAll)
- Set Content-Type header from multipart.Writer.FormDataContentType()
- Add X-Atlassian-Token: no-check header for CSRF protection
- Maintain existing retry logic for rate limiting

Command Structure:

- Follow existing patterns from comment and link commands
- Use cobra for command definitions and flag parsing
- Add attachments alias for consistency
- Space-separated file paths (no comma support)
- Require issue key + attachment ID for remove/download operations
- Default to original filename for downloads

Response Handling:

- Parse Jira attachment metadata (id, filename, size, mimeType, author, created, content, thumbnail)
- Convert to internal AttachmentInfo struct for display
- Format file sizes: B, KB, MB, GB with appropriate precision
- Display attachment IDs in brackets for easy reference

Integration:

- Update IssueDetail struct to include Attachments field
- Update issueDetailFields to parse attachment array from API
- Add attachments section to printIssueDetail after links, before comments
- No additional API call needed (already in response)

## Testing Strategy

Unit Tests:

- Multipart form construction and boundary generation
- File size formatting (bytes, KB, MB, GB edge cases)
- Attachment response parsing
- Error handling for missing files, invalid IDs

Integration Tests:

- Upload small text file
- Upload multiple files in single command
- List attachments and verify metadata
- Download attachment and verify content
- Delete attachment with and without --force
- Verify attachments appear in issue view

Manual Testing:

- Large file upload (near size limits)
- Special characters in filenames
- Permission errors (403)
- Size limit errors (413)
- Network interruption handling

## Notes

Research findings:

- Jira Cloud max file size: 1 GB default, 2 GB absolute maximum
- No file type restrictions in Jira Cloud (unlike Server/Data Center)
- Thumbnails generated for GIF, JPEG, PNG only
- UTF-8 encoding recommended but special characters may have issues
- API v3 endpoints recommended (v2 being deprecated)
- Download URL requires same authentication as other endpoints
- Attachments field included in default issue response (no extra API call)

Implementation considerations:

- Use multipart.NewWriter for form construction
- Call writer.Close() before making HTTP request
- Stream large files with io.Copy to limit memory usage
- Format sizes with 1024-based units (KiB, MiB) or 1000-based (KB, MB)
- Include attachment IDs in all output for easy copy-paste to remove/download
- Confirmation prompt prevents accidental deletions
- --force flag enables scripting and automation

Future enhancements tracked in list.md:

- Add list subcommand to comment and link for consistency
- Support -f - for reading file paths from stdin
- Progress indicators for large uploads/downloads
- Attachment content search or grep
